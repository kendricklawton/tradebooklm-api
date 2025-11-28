package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/api/option"
	"google.golang.org/genai"
)

type Clients struct {
	DB     *sql.DB
	Gemini *genai.Client
	Stripe string
	KMS    *KMSClient
}

type KMSClient struct {
	client  *kms.KeyManagementClient
	keyName string
	// Cache Key: String(Ciphertext), Value: []byte(Plaintext Key)
	dekCache *lru.LRU[string, []byte]
}

func InitializeConfig() (*Clients, error) {
	var (
		geminiApiKey    = mustGetenv("GEMINI_API_KEY")
		googleProjectID = mustGetenv("GOOGLE_PROJECT_ID")
		kmsKeyName      = mustGetenv("KMS_KEY_NAME")
		stripeApiKey    = mustGetenv("STRIPE_API_KEY")
	)

	ctx := context.Background()

	var db *sql.DB
	var err error

	if _, isRemote := os.LookupEnv("INSTANCE_CONNECTION_NAME"); isRemote {
		db, err = connectWithConnector()
	} else {
		db, err = connectWithLocalhost()
	}

	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	gemini, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error initializing gemini client: %w", err)
	}

	kmsClient, err := newKMSClient(ctx, googleProjectID, kmsKeyName)
	if err != nil {
		return nil, fmt.Errorf("error initializing KMS client: %w", err)
	}

	return &Clients{
		DB:     db,
		Stripe: stripeApiKey,
		KMS:    kmsClient,
		Gemini: gemini,
	}, nil
}

func (c *Clients) CloseDB() {
	if c.DB != nil {
		c.DB.Close()
	}
}

func connectWithConnector() (*sql.DB, error) {
	var (
		dbUser                 = mustGetenv("DB_USER")
		dbPwd                  = mustGetenv("DB_PASS")
		dbName                 = mustGetenv("DB_NAME")
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME")
		usePrivate             = os.Getenv("PRIVATE_IP")
	)

	dsn := fmt.Sprintf("user=%s password=%s database=%s", dbUser, dbPwd, dbName)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	var opts []cloudsqlconn.Option
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPrivateIP()))
	}
	// WithLazyRefresh() Option is used to perform refresh
	// when needed, rather than on a scheduled interval.
	// This is recommended for serverless environments to
	// avoid background refreshes from throttling CPU.
	opts = append(opts, cloudsqlconn.WithLazyRefresh())
	d, err := cloudsqlconn.NewDialer(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	// Use the Cloud SQL connector to handle connecting to the instance.
	// This approach does *NOT* require the Cloud SQL proxy.
	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	}
	dbURI := stdlib.RegisterConnConfig(config)
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// ---------------------------------------------------------
	// CRITICAL FOR SERVERLESS
	// ---------------------------------------------------------

	// Limit the number of open connections PER INSTANCE.
	// Since serverless scales horizontally (adding instances),
	// we keep vertical connections (per instance) extremely low.
	// 2 is usually sufficient: one for a read/write, one for a spare.
	dbPool.SetMaxOpenConns(2)

	// Limit idle connections.
	// In serverless, we don't want to hold onto connections we aren't using.
	dbPool.SetMaxIdleConns(1)

	// Set a lifetime to ensure connections are refreshed and don't get stale
	// or stuck in a bad state if the underlying infrastructure changes.
	dbPool.SetConnMaxLifetime(30 * time.Minute)

	return dbPool, nil
}

func connectWithLocalhost() (*sql.DB, error) {
	dbUser := getenvWithDefault("LOCAL_DB_USER", "postgres")
	dbPwd := os.Getenv("LOCAL_DB_PASS")
	dbTCPHost := getenvWithDefault("LOCAL_DB_HOST", "localhost")
	dbPort := getenvWithDefault("LOCAL_DB_PORT", "5432")
	dbName := getenvWithDefault("LOCAL_DB_NAME", "postgres")

	dbURI := fmt.Sprintf("host=%s user=%s password=%s port=%s database=%s sslmode=disable",
		dbTCPHost, dbUser, dbPwd, dbPort, dbName)

	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open (Local): %w", err)
	}

	log.Println("Local database connection pool initialized")
	return dbPool, nil
}

func getenvWithDefault(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Fatal Error in config.go: %s environment variable not set.\n", k)
	}
	return v
}

func newKMSClient(ctx context.Context, projectID, keyName string) (*KMSClient, error) {
	client, err := kms.NewKeyManagementClient(ctx, option.WithQuotaProject(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	if keyName == "" {
		return nil, fmt.Errorf("KMS key name cannot be empty")
	}

	// Initialize LRU Cache
	// 1000 items: Assuming ~32 bytes per key, this is negligible memory (KB range).
	// 5 minute TTL: Security trade-off. Revoked keys stay active for max 5 mins.
	cache := lru.NewLRU[string, []byte](1000, nil, 5*time.Minute)

	return &KMSClient{
		client:   client,
		keyName:  keyName,
		dekCache: cache,
	}, nil
}

// Encrypt simply calls KMS.
// Note: We generally do not cache encryption results because secure encryption
// schemes often introduce randomness (IVs), meaning the same plaintext yields
// different ciphertext every time.
func (k *KMSClient) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	req := &kmspb.EncryptRequest{
		Name:      k.keyName,
		Plaintext: plaintext,
	}

	resp, err := k.client.Encrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return resp.Ciphertext, nil
}

// Decrypt attempts to retrieve the decrypted key from the local cache.
// If missing, it calls Google KMS and caches the result.
func (k *KMSClient) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	// 1. Check Cache (Zero API Cost)
	// We use the encrypted bytes (as a string) as the lookup key.
	cacheKey := string(ciphertext)
	if plaintext, ok := k.dekCache.Get(cacheKey); ok {
		return plaintext, nil
	}

	// 2. Cache Miss: Call Google KMS (1 API Call)
	req := &kmspb.DecryptRequest{
		Name:       k.keyName,
		Ciphertext: ciphertext,
	}

	resp, err := k.client.Decrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// 3. Store in Cache for future calls
	k.dekCache.Add(cacheKey, resp.Plaintext)

	return resp.Plaintext, nil
}

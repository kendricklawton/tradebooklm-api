package config

import (
	"context"
	"crypto/aes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"tradebooklm-server/internal/encryption"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"google.golang.org/genai"
)

type Clients struct {
	DB     *sql.DB
	Stripe string
	Gemini *genai.Client
}

func connectWithConnector() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error in connect_connector.go: %s environment variable not set.\n", k)
		}
		return v
	}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep passwords and other secrets safe.
	var (
		dbUser                 = mustGetenv("DB_USER")                  // e.g. 'my-db-user'
		dbPwd                  = mustGetenv("DB_PASS")                  // e.g. 'my-db-password'
		dbName                 = mustGetenv("DB_NAME")                  // e.g. 'my-database'
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME") // e.g. 'project:region:instance'
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
	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	}
	dbURI := stdlib.RegisterConnConfig(config)
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		log.Println("Error opening database connection:", err)
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	log.Println("Database connection established")
	return dbPool, nil
}

func InitializeConfig() (*Clients, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error in connect_connector.go: %s environment variable not set.\n", k)
		}
		return v
	}
	var (
		geminiApiKey = mustGetenv("GEMINI_API_KEY")
		stripeApiKey = mustGetenv("STRIPE_API_KEY")
		dbKey        = mustGetenv("DB_KEY")
		// workosApiKey   = mustGetenv("WORKOS_API_KEY")
		// workosClientId = mustGetenv("WORKOS_CLIENT_ID")
	)

	ctx := context.Background()

	db, err := connectWithConnector()
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	gemini, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing gemini client: %w", err)
	}

	encryptionKey, err := base64.StdEncoding.DecodeString(dbKey)
	if err != nil {
		log.Fatalf("FATAL: Failed to decode encryption key: %v", err)
	}

	if len(encryptionKey) != 32 {
		log.Fatalf("FATAL: Encryption key must be 32 bytes long, but got %d", len(encryptionKey))
	}

	cipherBlock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		log.Fatalf("FATAL: Failed to create AES cipher: %v", err)
	}

	// sso.Configure(workosApiKey, workosClientId)
	// directorysync.SetAPIKey(workosApiKey)

	encryption.InitEncryption(cipherBlock)

	return &Clients{
		DB:     db,
		Stripe: stripeApiKey,
		Gemini: gemini,
	}, nil
}

func (c *Clients) CloseDB() {
	if c.DB != nil {
		c.DB.Close()
	}
}

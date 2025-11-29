package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"tradebooklm-api/internal/helpers"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/genai"
)

type Clients struct {
	DB     *sql.DB
	Gemini *genai.Client
	Stripe string
}

func InitializeConfig() (*Clients, error) {
	var (
		geminiApiKey = helpers.MustGetenv("GEMINI_API_KEY")
		stripeApiKey = helpers.MustGetenv("STRIPE_API_KEY")
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

func connectWithConnector() (*sql.DB, error) {
	var (
		dbUser                 = helpers.MustGetenv("DB_USER")
		dbPwd                  = helpers.MustGetenv("DB_PASS")
		dbName                 = helpers.MustGetenv("DB_NAME")
		instanceConnectionName = helpers.MustGetenv("INSTANCE_CONNECTION_NAME")
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

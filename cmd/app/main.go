package main

import (
	"context" // Added for graceful shutdown
	"log"
	"net/http" // Added for http.Server and ListenAndServe
	"os"
	"os/signal" // Added for catching signals
	"syscall"   // Added for catching SIGTERM
	"time"      // Added for shutdown timeout

	"tradebooklm-server/internal/handlers"
	"tradebooklm-server/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Attempt to load .env file for local development.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	log.Printf("Running in %s mode", gin.Mode())

	// Initialize application configuration, including database connection and clients.
	// (Uncomment this section when you have your config logic ready)
	// config, err := config.InitializeConfig()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize clients: %v", err)
	// }
	// defer config.CloseDB()

	// Set up the Gin router and middleware
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Cloud Run handles proxying, so setting trusted proxies to nil is fine.
	err = router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	router.Use(
		middleware.CORS(),
	)

	// --- Route Definitions (All unchanged) ---

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the TradeBook API!",
		})
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Test route",
		})
	})

	// Tradebook related routes
	router.POST("/tradebook", func(c *gin.Context) {
		handlers.CreateTradebookHandler(c)
	})

	router.DELETE("/tradebook/:tradebookId", func(c *gin.Context) {
		handlers.DeleteTradebookHandler(c)
	})

	router.GET("/tradebook/:tradebookId", func(c *gin.Context) {
		handlers.GetTradebooksHandler(c)
	})

	router.GET("/tradebooks", func(c *gin.Context) {
		handlers.GetTradebooksHandler(c)
	})

	router.PATCH("/tradebook/:tradebookId", func(c *gin.Context) {
		handlers.UpdateTradebookHandler(c)
	})

	// Trade related routes
	router.POST("/trade/:tradebookId", func(c *gin.Context) {
		handlers.CreateTradesHandler(c)
	})

	router.GET("/trade/:tradebookId", func(c *gin.Context) {
		handlers.GetTradesHandler(c)
	})

	router.PATCH("/trade/:tradebookId/:tradeId", func(c *gin.Context) {
		handlers.UpdateTradesHandler(c)
	})

	router.DELETE("/trade/:tradebookId", func(c *gin.Context) {
		handlers.DeleteTradesHandler(c)
	})

	// --- Server Start Logic (Cloud Run Graceful Shutdown) ---

	// Cloud Run sets the PORT environment variable (e.g., "8080").
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local development
	}

	addr := ":" + port
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Run the server in a goroutine so it doesn't block the main thread.
	go func() {
		log.Printf("Server starting and listening on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// This is a fatal error, e.g., port already in use.
			log.Fatalf("Server failed to listen: %v", err)
		}
	}()

	// Wait for interrupt signal (SIGINT for local, SIGTERM for Cloud Run).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server... Awaiting request completion.")

	// Create a context with a 5-second timeout for graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down the server.
	if err := srv.Shutdown(ctx); err != nil {
		// This happens if the server can't shut down within the timeout.
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

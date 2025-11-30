package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tradebooklm-api/internal/config"
	"tradebooklm-api/internal/services"
	"tradebooklm-api/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No env file found. Using System Environment Variables")
	}

	config, err := config.InitializeConfig()
	if err != nil {
		log.Fatalf("Failed to initialize clients: %v", err)
	}
	middleware.InitAuth()
	defer config.CloseDB()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	log.Printf("Running in %s mode", gin.Mode())

	// Cloud Run handles proxying
	err = router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	router.Use(
		middleware.CORS(),
	)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the TradeBook API!",
		})
	})

	webhookAPI := router.Group("/")
	webhookAPI.Use(middleware.WebhookMiddleware())
	{
		webhookAPI.POST("/user", func(c *gin.Context) {
			services.UpsertUser(c, config.DB)
		})

		webhookAPI.DELETE("/user", func(c *gin.Context) {
			services.DeleteUser(c, config.DB)
		})

		webhookAPI.PATCH("/user", func(c *gin.Context) {
			services.UpsertUser(c, config.DB)
		})
	}

	api := router.Group("/")
	api.Use(
		middleware.AuthMiddleware(),
	)
	{
		api.POST("/tradebook", func(c *gin.Context) {
			services.CreateTradebook(c, config.DB)
		})

		api.DELETE("/tradebook/:tradebookId", func(c *gin.Context) {
			services.DeleteTradebook(c, config.DB)
		})

		// api.DELETE("/tradebooks", func(c *gin.Context) {
		// 	services.DeleteTradebooks(c, config.DB)
		// })

		api.GET("/tradebook/:tradebookId", func(c *gin.Context) {
			services.GetTradebook(c, config.DB)
		})

		api.GET("/tradebooks", func(c *gin.Context) {
			services.GetTradebooks(c, config.DB)
			// c.JSON(200, []models.Tradebook{
			// 	{
			// 		ID:        "test-1",
			// 		Title:     "Test Tradebook 1",
			// 		CreatedAt: time.Now(),
			// 		UpdatedAt: time.Now(),
			// 		Role:      models.Owner,
			// 	},
			// 	{
			// 		ID:        "test-2",
			// 		Title:     "Test Tradebook 2",
			// 		CreatedAt: time.Now(),
			// 		UpdatedAt: time.Now(),
			// 		Role:      models.Editor,
			// 	},
			// })
		})

		api.PATCH("/tradebook/:tradebookId", func(c *gin.Context) {
			services.UpdateTradebook(c, config.DB)
		})

		api.POST("/trade/:tradebookId", func(c *gin.Context) {
			services.CreateTrades(c, config.DB)
		})

		api.GET("/trade/:tradebookId", func(c *gin.Context) {
			services.GetTrades(c, config.DB)
		})

		api.PATCH("/trade/:tradebookId/:tradeId", func(c *gin.Context) {
			services.UpdateTrades(c, config.DB)
		})

		api.DELETE("/trade/:tradebookId", func(c *gin.Context) {
			services.DeleteTrades(c, config.DB)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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

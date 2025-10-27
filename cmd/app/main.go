package main

import (
	"log"
	"tradebooklm-server/internal/config"
	"tradebooklm-server/internal/handlers"
	"tradebooklm-server/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	log.Printf("Running in %s mode", gin.Mode())

	config, err := config.InitializeConfig()
	if err != nil {
		log.Fatalf("Failed to initialize clients: %v", err)
	}

	defer config.CloseDB()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	err = router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	router.Use(
		middleware.CORS(),
	)

	// router.GET("/", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "Welcome to the TradeBook API!",
	// 	})
	// })

	// router.Static("/assets", "./assets")

	// router.GET("/favicon.ico", func(c *gin.Context) {
	// 	c.File("./assets/favicon.ico")
	// })

	// userRoutes := router.Group("/user")
	// userRoutes.Use(...WhatEverMiddleWare)
	// {
	// 	userRoutes.POST("", func(c *gin.Context) {
	// 		handlers.CreateUserHandler(c, config.CipherBlock, config.DB)
	// 	})

	// 	userRoutes.DELETE("/:workosId", func(c *gin.Context) {
	// 		handlers.DeleteUserHandler(c, config.DB)
	// 	})
	// }

	// User related routes
	router.POST("/user", func(c *gin.Context) {
		handlers.CreateUserHandler(c, config.DB)
	})

	router.DELETE("/user/:workosId", func(c *gin.Context) {
		handlers.DeleteUserHandler(c, config.DB)
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

	// stripeRoutes := router.Group("/stripe")
	// stripeRoutes.Use(middleware.WebhookAuth())
	// {
	// 	stripeRoutes.POST("", func(c *gin.Context) {
	// 		handlers.CreateUserHandler(c, cipherBlock.CipherBlock, clients.Database)
	// 	})

	// 	stripeRoutes.GET("/:workosId", func(c *gin.Context) {
	// 		handlers.GetUserHandler(c, cipherBlock.CipherBlock, clients.Database)
	// 	})

	// 	stripeRoutes.PATCH("/:workosId", func(c *gin.Context) {
	// 		handlers.UpdateUserHandler(c, cipherBlock.CipherBlock, clients.Database)
	// 	})

	// 	stripeRoutes.DELETE("/:workosId", func(c *gin.Context) {
	// 		handlers.DeleteUserHandler(c, cipherBlock.CipherBlock, clients.Database)
	// 	})
	// }

	// Run the server
	router.Run(":8080")
}

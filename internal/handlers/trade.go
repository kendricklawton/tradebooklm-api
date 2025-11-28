package handlers

import (
	"database/sql"
	"net/http"
	"tradebooklm-api/internal/config"

	"github.com/gin-gonic/gin"
)

func CreateTrades(c *gin.Context, db *sql.DB, kmsClient *config.KMSClient) {
	c.Status(http.StatusCreated)
}

func GetTrades(c *gin.Context, db *sql.DB, kmsClient *config.KMSClient) {
	c.Status(http.StatusOK)
}

func UpdateTrades(c *gin.Context, db *sql.DB, kmsClient *config.KMSClient) {
	c.Status(http.StatusOK)
}

func DeleteTrades(c *gin.Context, db *sql.DB) {
	c.Status(http.StatusNoContent)
}

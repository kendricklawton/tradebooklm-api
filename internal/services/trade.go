package services

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTrades(c *gin.Context, db *sql.DB) {
	c.Status(http.StatusCreated)
}

func GetTrades(c *gin.Context, db *sql.DB) {
	c.Status(http.StatusOK)
}

func UpdateTrades(c *gin.Context, db *sql.DB) {
	c.Status(http.StatusOK)
}

func DeleteTrades(c *gin.Context, db *sql.DB) {
	c.Status(http.StatusNoContent)
}

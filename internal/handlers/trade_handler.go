package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTradesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Trades created successfully"})
}

func GetTradesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Trades fetched successfully"})
}

func UpdateTradesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Trades updated successfully"})
}

func DeleteTradesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Trades deleted successfully"})
}

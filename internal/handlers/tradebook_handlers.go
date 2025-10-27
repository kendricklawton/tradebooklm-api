package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTradebookHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Tradebook created successfully"})
}

func GetTradebooksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Tradebooks fetched successfully"})
}

func UpdateTradebookHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Tradebook updated successfully"})
}

func DeleteTradebookHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Tradebook deleted successfully"})
}

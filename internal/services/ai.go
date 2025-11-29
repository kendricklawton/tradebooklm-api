package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OpenAI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OpenAI model created successfully"})
}

func GeminiAI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Gemini AI model created successfully"})
}

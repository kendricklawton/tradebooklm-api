package handlers

import (
	"net/http"

	"cloud.google.com/go/vertexai/genai"
	"github.com/gin-gonic/gin"
)

func OpenAIModelHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OpenAI model created successfully"})
}

func VertexAIModelHandler(c *gin.Context, vertexAiClient *genai.Client) {
	c.JSON(http.StatusOK, gin.H{"message": "Vertex AI model created successfully"})
}

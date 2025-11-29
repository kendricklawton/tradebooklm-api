package helpers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetWorkosID(c *gin.Context) (string, bool) {
	workosIdInterface, exists := c.Get("workos_id")
	if !exists {
		log.Println("workos_id not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WorkOS ID not found in context"})
		return "", false
	}

	workosId, ok := workosIdInterface.(string)
	if !ok {
		log.Println("workos_id in context is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid WorkOS ID format in context"})
		return "", false
	}

	return workosId, true
}

func GetPaginationParams(c *gin.Context) (int32, int32) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	// Hard limit of 100 items per page
	if limit > 100 {
		limit = 100
	}

	// Calculate offset
	offset := (page - 1) * limit

	return int32(limit), int32(offset)
}

func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func MustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Fatal Error in config.go: %s environment variable not set.\n", k)
	}
	return v
}

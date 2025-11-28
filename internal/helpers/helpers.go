package helpers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetWorkosID retrieves the workos_id from the Gin context.
// If the ID exists and is a string, it returns the ID and true.
// Otherwise, it logs the error, sends an appropriate JSON error response,
// and returns an empty string and false. This allows handlers to exit early.
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

// ParsePagination extracts limit and page from the query string,
// applies defaults/constraints, and calculates the SQL offset.
// Returns: (limit, offset)
func ParsePagination(c *gin.Context) (int, int) {
	// Parse "limit" with a default of 25
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if err != nil || limit <= 0 {
		limit = 25
	}
	// Hard cap limit to prevent massive queries
	if limit > 100 {
		limit = 100
	}

	// Parse "page" with a default of 1
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	// Calculate SQL offset
	offset := (page - 1) * limit

	return limit, offset
}

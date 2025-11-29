package services

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"tradebooklm-api/internal/database"
	"tradebooklm-api/internal/models"
)

func UpsertUser(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	var req models.CreateWorkosUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	q := database.New(conn)

	_, err := q.UpsertUser(ctx, req.ID)
	if err != nil {
		log.Printf("Error upserting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upsert user"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteUser(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	userID := c.Param("workosId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID cannot be empty"})
		return
	}

	log.Printf("Received request to delete user: %s", userID)

	query := `DELETE FROM users WHERE id = $1`

	result, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify deletion"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

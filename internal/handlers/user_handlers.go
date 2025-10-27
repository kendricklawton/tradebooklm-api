package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"tradebooklm-server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func CreateUserHandler(c *gin.Context, db *sql.DB) {
	var user models.WorkosUser

	log.Printf("User: %v", user)

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	dbUser := models.DBUser{
		WorkosID:  user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	log.Printf("Creating user with WorkOS ID: %s", dbUser.WorkosID)

	query := `
		INSERT INTO users (workos_id, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id, workos_id, created_at, updated_at
	`

	ctx := c.Request.Context()
	err := db.QueryRowContext(ctx, query, dbUser.WorkosID, dbUser.CreatedAt, dbUser.UpdatedAt).Scan(
		&dbUser.ID,
		&dbUser.WorkosID,
		&dbUser.CreatedAt,
		&dbUser.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("Conflict: User with WorkOS ID %s already exists", dbUser.WorkosID)
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	log.Printf("Successfully created dbUser: %+v", dbUser)
	c.Status(http.StatusCreated)
}

func DeleteUserHandler(c *gin.Context, db *sql.DB) {
	workosID := c.Param("workosId")

	log.Printf("Deleting user with WorkOS ID: %s", workosID)

	query := "DELETE FROM users WHERE workos_id = $1"

	ctx := c.Request.Context()
	result, err := db.ExecContext(ctx, query, workosID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if rowsAffected == 0 {
		log.Printf("User not found for deletion: %s", workosID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	log.Printf("Successfully deleted workos user: %s", workosID)
	c.Status(http.StatusNoContent)
}

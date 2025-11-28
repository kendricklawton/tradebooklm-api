package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"tradebooklm-api/internal/helpers"
	"tradebooklm-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Helper to convert string IDs to UUIDs safely
func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// setRLS sets the current user ID in the Postgres session for Row Level Security.
// This must be called within a transaction.
func setRLS(ctx context.Context, tx *sql.Tx, userID string) error {
	// The 'true' argument makes this setting local to the transaction
	_, err := tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID)
	return err
}

func CreateTradebook(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	log.Printf("Received request to create tradebook for user: %s", workosId)

	// Start a transaction (Required for RLS scoping)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// 1. Set RLS Context
	if err := setRLS(ctx, tx, workosId); err != nil {
		log.Printf("Error setting RLS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// 2. Ensure user exists (Upsert)
	upsertUserQuery := `
		INSERT INTO users (id) VALUES ($1)
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
	`
	if _, err := tx.ExecContext(ctx, upsertUserQuery, workosId); err != nil {
		log.Printf("Error upserting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	// 3. Create tradebook and member
	// Note: title is now plaintext
	tradebookTitle := "Untitled Tradebook"

	createQuery := `
		WITH new_tradebook AS (
			INSERT INTO tradebooks (owner_id, title)
			VALUES ($1, $2)
			RETURNING id, owner_id
		)
		INSERT INTO tradebook_members (tradebook_id, user_id, role)
		SELECT id, owner_id, 'owner'
		FROM new_tradebook
		RETURNING tradebook_id
	`

	var tradebookID uuid.UUID
	err = tx.QueryRowContext(ctx, createQuery, workosId, tradebookTitle).Scan(&tradebookID)
	if err != nil {
		log.Printf("Error creating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tradebook"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.JSON(http.StatusCreated, tradebookID)
}

func DeleteTradebook(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := parseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tradebook ID"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// Even with RLS, we explicitly check owner_id to ensure only owners can delete
	deleteQuery := `DELETE FROM tradebooks WHERE id = $1 AND owner_id = $2`

	result, err := tx.ExecContext(ctx, deleteQuery, tbUUID, workosId)
	if err != nil {
		log.Printf("Error deleting tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or access denied"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteTradebooks(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// RLS allows us to see the rows, but we restrict deletion to owner_id
	deleteQuery := `DELETE FROM tradebooks WHERE owner_id = $1`

	result, err := tx.ExecContext(ctx, deleteQuery, workosId)
	if err != nil {
		log.Printf("Error deleting tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tradebooks"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No tradebooks found"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

func GetTradebook(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := parseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// Plaintext selection. RLS handles the visibility check implicitly.
	// We join members to get the specific role of the current user.
	getQuery := `
		SELECT t.id, t.title, t.description, t.created_at, t.updated_at, tm.role
		FROM tradebooks t
		JOIN tradebook_members tm ON t.id = tm.tradebook_id
		WHERE t.id = $1 AND tm.user_id = $2
	`

	var id uuid.UUID
	var title string
	var description sql.NullString // Handle potential nulls
	var createdAt, updatedAt time.Time
	var role string

	err = tx.QueryRowContext(ctx, getQuery, tbUUID, workosId).Scan(&id, &title, &description, &createdAt, &updatedAt, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or access denied"})
			return
		}
		log.Printf("Error fetching tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	response := models.TradebookResponse{
		ID:    id.String(),
		Title: title,

		Role:      models.Role(role),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func GetTradebooks(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	limit, offset := helpers.ParsePagination(c)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// Plaintext list
	listQuery := `
		SELECT t.id, t.title, t.description, t.created_at, t.updated_at, tm.role
		FROM tradebooks t
		JOIN tradebook_members tm ON t.id = tm.tradebook_id
		WHERE tm.user_id = $1
		ORDER BY t.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := tx.QueryContext(ctx, listQuery, workosId, limit, offset)
	if err != nil {
		log.Printf("Error fetching tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	var responseList []models.TradebookResponse

	for rows.Next() {
		var id uuid.UUID
		var title string
		var description sql.NullString
		var createdAt, updatedAt time.Time
		var role string

		if err := rows.Scan(&id, &title, &description, &createdAt, &updatedAt, &role); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		responseList = append(responseList, models.TradebookResponse{
			ID:    id.String(),
			Title: title,

			Role:      models.Role(role),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	// Return empty array instead of null for consistency
	if responseList == nil {
		responseList = []models.TradebookResponse{}
	}

	c.JSON(http.StatusOK, responseList)
}

func GetTradebookTitle(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := parseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	var title string
	// Check membership implicitly via JOIN to ensure they have access
	query := `
		SELECT t.title
		FROM tradebooks t
		JOIN tradebook_members tm ON t.id = tm.tradebook_id
		WHERE t.id = $1 AND tm.user_id = $2
	`

	err = tx.QueryRowContext(ctx, query, tbUUID, workosId).Scan(&title)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"title": title})
}

func UpdateTradebook(c *gin.Context, db *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := parseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.TradebookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Body"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if err := setRLS(ctx, tx, workosId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security context error"})
		return
	}

	// Update query (Plaintext)
	// We check EXISTS in tradebook_members to ensure only members (editors/owners) can update.
	// You might want to refine this to only 'owner' or 'editor' roles specifically.
	updateQuery := `
		UPDATE tradebooks
		SET title = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		AND EXISTS (
			SELECT 1 FROM tradebook_members
			WHERE tradebook_id = $3
			AND user_id = $4
			AND role IN ('owner', 'editor')
		)
	`

	// Assuming req.Description exists in your model now.
	// If not, remove the second arg and fix the query.
	result, err := tx.ExecContext(ctx, updateQuery,
		strings.TrimSpace(req.Title),
		tbUUID,
		workosId,
	)

	if err != nil {
		log.Printf("Error updating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or access denied"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.Status(http.StatusOK)
}

package services

import (
	"database/sql"
	"log"
	"net/http"

	// Generated package
	"tradebooklm-api/internal/database"

	"tradebooklm-api/internal/helpers"
	"tradebooklm-api/internal/models" // API Response models

	"github.com/gin-gonic/gin"
)

func CreateTradebook(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	q := database.New(conn)

	// 1. Ensure User Exists
	_, err := q.UpsertUser(ctx, workosId)
	if err != nil {
		log.Printf("Error ensuring user exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 2. Start Transaction
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start failed"})
		return
	}
	defer tx.Rollback()

	qTx := q.WithTx(tx)

	// 3. Create Tradebook (Description ignored as requested)
	tb, err := qTx.CreateTradebook(ctx, database.CreateTradebookParams{
		OwnerID: workosId,
		Title:   "Untitled Tradebook",
	})
	if err != nil {
		log.Printf("Error creating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tradebook"})
		return
	}

	// 4. Add Member (Owner)
	_, err = qTx.UpsertTradebookMember(ctx, database.UpsertTradebookMemberParams{
		TradebookID: tb.ID,
		NewMemberID: workosId,
		Role:        database.TradebookRoleOwner,
		OwnerID:     workosId,
	})
	if err != nil {
		log.Printf("Error adding member: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign ownership"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusCreated, tb.ID)
}

func DeleteTradebook(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tradebook ID"})
		return
	}

	q := database.New(conn)

	err = q.DeleteTradebook(ctx, database.DeleteTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		log.Printf("Error deleting: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteTradebooks(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	q := database.New(conn)

	err := q.DeleteAllTradebooks(ctx, workosId)

	if err != nil {
		log.Printf("Error deleting all tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tradebooks"})
		return
	}

	c.Status(http.StatusNoContent)
}

func GetTradebook(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	q := database.New(conn)

	row, err := q.GetTradebook(ctx, database.GetTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or access denied"})
			return
		}
		log.Printf("Error fetching tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := models.Tradebook{
		ID:        row.ID.String(),
		Title:     row.Title,
		Role:      models.Role(row.UserRole),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func GetTradebooks(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	// 1. Calculate Pagination
	limit, offset := helpers.GetPaginationParams(c)

	q := database.New(conn)

	// 2. Fetch with Limit and Offset
	rows, err := q.ListTradebooks(ctx, database.ListTradebooksParams{
		UserID:    workosId,
		LimitVal:  limit,
		OffsetVal: offset,
	})
	if err != nil {
		log.Printf("Error fetching tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// 3. Map to Response
	responseList := make([]models.Tradebook, 0, len(rows))

	for _, row := range rows {
		responseList = append(responseList, models.Tradebook{
			ID:        row.ID.String(),
			Title:     row.Title,
			Role:      models.Role(row.UserRole),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	if responseList == nil {
		responseList = []models.Tradebook{}
	}

	c.JSON(http.StatusOK, responseList)
}

func UpdateTradebook(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.UpdateTradebookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Body"})
		return
	}

	q := database.New(conn)

	titleParam := sql.NullString{String: req.Title, Valid: req.Title != ""}

	_, err = q.UpdateTradebook(ctx, database.UpdateTradebookParams{
		Title:       titleParam,
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or permission denied"})
			return
		}
		log.Printf("Error updating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch the updated tradebook with role information
	updatedRow, err := q.GetTradebook(ctx, database.GetTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		log.Printf("Error fetching updated tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, models.Tradebook{
		ID:        updatedRow.ID.String(),
		Title:     updatedRow.Title,
		CreatedAt: updatedRow.CreatedAt,
		UpdatedAt: updatedRow.UpdatedAt,
		Role:      models.Role(updatedRow.UserRole),
	})
}

func CreateTradebookFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	q := database.New(conn)

	// 1. Ensure User Exists
	_, err := q.UpsertUser(ctx, workosId)
	if err != nil {
		log.Printf("Error ensuring user exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 2. Start Transaction
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start failed"})
		return
	}
	defer tx.Rollback()

	qTx := q.WithTx(tx)

	// 3. Create Tradebook (Description ignored as requested)
	tb, err := qTx.CreateTradebook(ctx, database.CreateTradebookParams{
		OwnerID: workosId,
		Title:   "Untitled Tradebook",
	})
	if err != nil {
		log.Printf("Error creating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tradebook"})
		return
	}

	// 4. Add Member (Owner)
	_, err = qTx.UpsertTradebookMember(ctx, database.UpsertTradebookMemberParams{
		TradebookID: tb.ID,
		NewMemberID: workosId,
		Role:        database.TradebookRoleOwner,
		OwnerID:     workosId,
	})
	if err != nil {
		log.Printf("Error adding member: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign ownership"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusCreated, tb.ID)
}

func DeleteTradebookFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tradebook ID"})
		return
	}

	q := database.New(conn)

	err = q.DeleteTradebook(ctx, database.DeleteTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		log.Printf("Error deleting: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteTradebooksFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	q := database.New(conn)

	err := q.DeleteAllTradebooks(ctx, workosId)

	if err != nil {
		log.Printf("Error deleting all tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tradebooks"})
		return
	}

	c.Status(http.StatusNoContent)
}

func GetTradebookFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	q := database.New(conn)

	row, err := q.GetTradebook(ctx, database.GetTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or access denied"})
			return
		}
		log.Printf("Error fetching tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := models.Tradebook{
		ID:        row.ID.String(),
		Title:     row.Title,
		Role:      models.Role(row.UserRole),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func GetTradebooksFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	// 1. Calculate Pagination
	limit, offset := helpers.GetPaginationParams(c)

	q := database.New(conn)

	// 2. Fetch with Limit and Offset
	rows, err := q.ListTradebooks(ctx, database.ListTradebooksParams{
		UserID:    workosId,
		LimitVal:  limit,
		OffsetVal: offset,
	})
	if err != nil {
		log.Printf("Error fetching tradebooks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// 3. Map to Response
	responseList := make([]models.Tradebook, 0, len(rows))

	for _, row := range rows {
		responseList = append(responseList, models.Tradebook{
			ID:        row.ID.String(),
			Title:     row.Title,
			Role:      models.Role(row.UserRole),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	if responseList == nil {
		responseList = []models.Tradebook{}
	}

	c.JSON(http.StatusOK, responseList)
}

func UpdateTradebookFirestore(c *gin.Context, conn *sql.DB) {
	ctx := c.Request.Context()

	workosId, ok := helpers.GetWorkosID(c)
	if !ok {
		return
	}

	tbUUID, err := helpers.ParseUUID(c.Param("tradebookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.UpdateTradebookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Body"})
		return
	}

	q := database.New(conn)

	titleParam := sql.NullString{String: req.Title, Valid: req.Title != ""}

	_, err = q.UpdateTradebook(ctx, database.UpdateTradebookParams{
		Title:       titleParam,
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tradebook not found or permission denied"})
			return
		}
		log.Printf("Error updating tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch the updated tradebook with role information
	updatedRow, err := q.GetTradebook(ctx, database.GetTradebookParams{
		TradebookID: tbUUID,
		UserID:      workosId,
	})

	if err != nil {
		log.Printf("Error fetching updated tradebook: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, models.Tradebook{
		ID:        updatedRow.ID.String(),
		Title:     updatedRow.Title,
		CreatedAt: updatedRow.CreatedAt,
		UpdatedAt: updatedRow.UpdatedAt,
		Role:      models.Role(updatedRow.UserRole),
	})
}

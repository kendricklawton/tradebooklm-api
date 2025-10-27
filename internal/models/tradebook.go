package models

import (
	"time"
)

// Tradebook corresponds to the 'tradebooks' table.
type Tradebook struct {
	ID             string    `json:"id" db:"id"`
	WorkspaceID    string    `json:"workspace_id" db:"workspace_id"`
	Title          string    `json:"title" db:"title"`
	TotalTrades    int64     `json:"total_trades" db:"total_trades"`
	TotalResources int64     `json:"total_resources" db:"total_resources"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

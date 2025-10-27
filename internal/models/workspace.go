package models

import (
	"time"
)

type WorkspaceType string

const (
	WorkspacePersonal     WorkspaceType = "personal"
	WorkspaceOrganization WorkspaceType = "organization"
)

type Workspace struct {
	ID              string        `json:"id" db:"id"`
	Name            string        `json:"name" db:"name"`
	Type            WorkspaceType `json:"type" db:"type"`
	PersonalOwnerID *string       `json:"personal_owner_id,omitempty" db:"personal_owner_id"`
	TotalTradebooks int64         `json:"total_tradebooks" db:"total_tradebooks"`
	TotalTrades     int64         `json:"total_trades" db:"total_trades"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type WorkspaceMember struct {
	WorkspaceID string `json:"workspace_id" db:"workspace_id"`
	UserID      string `json:"user_id" db:"user_id"`
	Role        string `json:"role" db:"role"`
}

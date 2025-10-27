package models

import "time"

type WorkosUser struct {
	ID        string    `json:"id" `
	CreatedAt time.Time `json:"created_at" `
	UpdatedAt time.Time `json:"updated_at" `
}

type DBUser struct {
	ID        *string   `json:"id" db:"id"`
	WorkosID  string    `json:"workos_id" db:"workos_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type AssetClass string

const (
	Equities    AssetClass = "equities"
	FixedIncome AssetClass = "fixed_income"
	Commodities AssetClass = "commodities"
	ETFs        AssetClass = "etfs"
	Forex       AssetClass = "forex"
	Derivatives AssetClass = "derivatives"
	Crypto      AssetClass = "crypto"
)

type OrderType string

const (
	Market    OrderType = "market"
	Limit     OrderType = "limit"
	Stop      OrderType = "stop"
	StopLimit OrderType = "stop_limit"
)

type PurchaseType string

const (
	Cash   PurchaseType = "cash"
	Margin PurchaseType = "margin"
)

type Role string

const (
	Editor Role = "editor"
	Owner  Role = "owner"
	Reader Role = "reader"
)

type UpdateTradebookRequest struct {
	Title string `json:"title" binding:"required"`
}

type CreateWorkosUserRequest struct {
	ID        string    `json:"id" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
	UpdatedAt time.Time `json:"updated_at" binding:"required"`
}

type User struct {
	ID        string    `json:"id"`
	WorkosID  string    `json:"workos_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tradebook struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      Role      `json:"role"` // This is usually injected during retrieval
}

type Trade struct {
	ID          string `json:"id"`
	TradebookID string `json:"tradebook_id"`

	IsOpen bool `json:"is_open"`

	AssetClass   AssetClass   `json:"asset_class"`
	PurchaseType PurchaseType `json:"purchase_type"`
	OrderType    OrderType    `json:"order_type"`

	EntryDate time.Time `json:"entry_date"`
	Symbol    string    `json:"symbol"`
	Currency  string    `json:"currency"` // Added: e.g. "USD", "BTC"

	// FINANCIAL FIELDS: Using Decimal instead of float/int
	// Quantity is Decimal to support Crypto/Forex (e.g. 0.05 BTC)
	EntryQuantity decimal.Decimal `json:"entry_quantity"`
	EntryPrice    decimal.Decimal `json:"entry_price"`
	EntryFees     decimal.Decimal `json:"entry_fees"`

	ExitLegs []*ExitLeg `json:"exit_legs"`

	Notes string `json:"notes,omitempty"` // omitempty if blank

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddTradeRequest struct {
	Title        string       `json:"title"`
	AssetClass   AssetClass   `json:"asset_class"`
	PurchaseType PurchaseType `json:"purchase_type"`
	OrderType    OrderType    `json:"order_type"`

	EntryDate time.Time `json:"entry_date"`
	Symbol    string    `json:"symbol"`
	Currency  string    `json:"currency"` // Added: e.g. "USD", "BTC"

	// FINANCIAL FIELDS
	EntryQuantity decimal.Decimal `json:"entry_quantity"`
	EntryPrice    decimal.Decimal `json:"entry_price"`
	EntryFees     decimal.Decimal `json:"entry_fees"`
}

type ExitLeg struct {
	ID      string `json:"id"`
	TradeID string `json:"trade_id"`

	ExitDate time.Time `json:"exit_date"`

	ExitQuantity decimal.Decimal `json:"exit_quantity"`
	ExitPrice    decimal.Decimal `json:"exit_price"`
	ExitFees     decimal.Decimal `json:"exit_fees"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddExitLegRequest struct {
	TradeID string `json:"trade_id"`

	ExitDate time.Time `json:"exit_date"`

	ExitQuantity decimal.Decimal `json:"exit_quantity"`
	ExitPrice    decimal.Decimal `json:"exit_price"`
	ExitFees     decimal.Decimal `json:"exit_fees"`

	Notes string `json:"notes,omitempty"`
}

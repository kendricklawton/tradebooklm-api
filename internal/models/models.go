package models

import "time"

type DatabaseUser struct {
	ID        string    `json:"id" db:"id"`
	WorkosID  string    `json:"workos_id" db:"workos_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type WorkosCreateUserRequest struct {
	ID        string    `json:"id" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
	UpdatedAt time.Time `json:"updated_at" binding:"required"`
}

type AssetClass string

type Role string

const (
	Editor Role = "editor"
	Owner  Role = "owner"
	Reader Role = "reader"
)

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

type TradebookUpdateRequest struct {
	Title string `json:"title" binding:"required"`
}

type TradebookResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatabaseTradebook struct {
	ID             string
	Role           Role
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EncryptedTitle []byte
}

type TradeResponse struct {
	ID            string             `json:"id"`
	TradebookID   string             `json:"tradebook_id"`
	AssetClass    AssetClass         `json:"asset_class"`
	PurchaseType  PurchaseType       `json:"purchase_type"`
	OrderType     OrderType          `json:"order_type"`
	Symbol        string             `json:"symbol"`
	EntryDate     time.Time          `json:"entry_date"`
	EntryQuantity int64              `json:"entry_quantity"`
	EntryPrice    float64            `json:"entry_price"`
	EntryFees     float64            `json:"entry_fees"`
	ExitLegs      []*ExitLegResponse `json:"exit_legs"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type ExitLegResponse struct {
	ID           string    `json:"id"`
	TradeID      string    `json:"trade_id"`
	ExitDate     time.Time `json:"exit_date"`
	ExitQuantity int64     `json:"exit_quantity"`
	ExitPrice    float64   `json:"exit_price"`
	ExitFees     float64   `json:"exit_fees"`
	CreatedAt    time.Time `json:"created_at"`
}

type DatabaseTrade struct {
	ID                     string
	TradebookID            string
	AssetClass             string
	PurchaseType           string
	OrderType              string
	EntryDate              time.Time
	Symbol                 string
	EntryQuantityEncrypted []byte
	EntryPriceEncrypted    []byte
	EntryFeesEncrypted     []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type DatabaseExitLeg struct {
	ID                    string
	TradeID               string
	ExitDate              time.Time
	ExitQuantityEncrypted []byte
	ExitPriceEncrypted    []byte
	ExitFeesEncrypted     []byte
	CreatedAt             time.Time
}

package models

import (
	"time"
	"tradebooklm-server/internal/encryption"
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

type Order string

const (
	Market    Order = "market"
	Limit     Order = "limit"
	Stop      Order = "stop"
	StopLimit Order = "stop_limit"
)

type Purchase string

const (
	Cash   Purchase = "cash"
	Margin Purchase = "margin"
)

type Trade struct {
	ID           string     `json:"id" db:"id"`
	TradebookID  string     `json:"tradebook_id" db:"tradebook_id"`
	AssetClass   AssetClass `json:"asset_class" db:"asset_class"`
	PurchaseType Purchase   `json:"purchase_type" db:"purchase"`
	OrderType    Order      `json:"order_type" db:"order"`
	EntryDate    time.Time  `json:"entry_date" db:"entry_date"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// These fields are now automatically encrypted!
	Symbol        encryption.EncryptedString          `json:"symbol" db:"symbol"`
	EntryQuantity encryption.EncryptedDecimal         `json:"entry_quantity" db:"entry_quantity"`
	EntryPrice    encryption.EncryptedDecimal         `json:"entry_price" db:"entry_price"`
	EntryFees     encryption.EncryptedNullableDecimal `json:"entry_fees" db:"entry_fees"`
}

// ExitLeg corresponds to the 'exit_legs' table.
type ExitLeg struct {
	ID       string    `json:"id" db:"id"`
	TradeID  string    `json:"tradeId" db:"trade_id"`
	ExitDate time.Time `json:"exitDate" db:"exit_date"`

	// These fields are decrypted by PostgreSQL.
	// UPDATED: All financial fields now use decimal.Decimal for precision.
	ExitQuantity encryption.EncryptedDecimal         `json:"exit_quantity" db:"exit_quantity"`
	ExitPrice    encryption.EncryptedDecimal         `json:"exit_price" db:"exit_price"`
	ExitFees     encryption.EncryptedNullableDecimal `json:"exit_fees" db:"exit_fees"`
}

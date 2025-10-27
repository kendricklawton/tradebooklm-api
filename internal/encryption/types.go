package encryption

import (
	"database/sql/driver"
	"fmt"

	"github.com/shopspring/decimal"
)

// --- EncryptedString ---

// EncryptedString is a string that is automatically encrypted/decrypted
type EncryptedString string

// Value implements the driver.Valuer interface.
// It encrypts the EncryptedString before writing to the database.
func (es EncryptedString) Value() (driver.Value, error) {
	if es == "" {
		return nil, nil // Store as NULL
	}
	return encrypt([]byte(string(es)))
}

// Scan implements the sql.Scanner interface.
// It decrypts the []byte from the database into the EncryptedString.
func (es *EncryptedString) Scan(value any) error {
	if value == nil {
		*es = ""
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte from database, got %T", value)
	}

	plaintext, err := decrypt(b)
	if err != nil {
		return fmt.Errorf("failed to scan encrypted value: %w", err)
	}

	*es = EncryptedString(plaintext)
	return nil
}

// --- EncryptedDecimal ---

// EncryptedDecimal is a decimal.Decimal that is automatically encrypted/decrypted
// It embeds decimal.Decimal so you can still use its methods.
type EncryptedDecimal struct {
	decimal.Decimal
}

// Value implements the driver.Valuer interface.
func (ed EncryptedDecimal) Value() (driver.Value, error) {
	// Convert the decimal to its string representation for encryption
	s := ed.Decimal.String()
	return encrypt([]byte(s))
}

// Scan implements the sql.Scanner interface.
func (ed *EncryptedDecimal) Scan(value any) error {
	if value == nil {
		ed.Decimal = decimal.Zero
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte from database, got %T", value)
	}

	plaintext, err := decrypt(b)
	if err != nil {
		return fmt.Errorf("failed to scan encrypted decimal: %w", err)
	}

	// Parse the decrypted string back into a decimal
	d, err := decimal.NewFromString(string(plaintext))
	if err != nil {
		return fmt.Errorf("failed to parse decrypted decimal: %w", err)
	}

	ed.Decimal = d
	return nil
}

// --- EncryptedNullableDecimal ---

// EncryptedNullableDecimal handles *decimal.Decimal (NULLable)
type EncryptedNullableDecimal struct {
	Decimal decimal.Decimal
	Valid   bool // True if not NULL
}

// Value implements the driver.Valuer interface.
func (end EncryptedNullableDecimal) Value() (driver.Value, error) {
	if !end.Valid {
		return nil, nil // Store as NULL
	}
	s := end.Decimal.String()
	return encrypt([]byte(s))
}

// Scan implements the sql.Scanner interface.
func (end *EncryptedNullableDecimal) Scan(value any) error {
	if value == nil {
		end.Valid = false
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte from database, got %T", value)
	}

	plaintext, err := decrypt(b)
	if err != nil {
		return fmt.Errorf("failed to scan encrypted nullable decimal: %w", err)
	}

	d, err := decimal.NewFromString(string(plaintext))
	if err != nil {
		return fmt.Errorf("failed to parse decrypted decimal: %w", err)
	}

	end.Decimal = d
	end.Valid = true
	return nil
}

// Helper methods to make it feel like a *decimal.Decimal
func (end EncryptedNullableDecimal) Ptr() *decimal.Decimal {
	if !end.Valid {
		return nil
	}
	// Create a copy to avoid pointer issues
	d := end.Decimal
	return &d
}

func NewEncryptedNullableDecimal(d *decimal.Decimal) EncryptedNullableDecimal {
	if d == nil {
		return EncryptedNullableDecimal{Valid: false}
	}
	return EncryptedNullableDecimal{Decimal: *d, Valid: true}
}

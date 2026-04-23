package model

import "time"

type LedgerEntry struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Type      string    `json:"type"` // "debit" or "credit"
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

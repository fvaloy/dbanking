package model

import "time"

type BankMovement struct {
	ID               string    `json:"id"`
	PaymentID        string    `json:"payment_id"`
	PaymentReference string    `json:"payment_reference"`
	Amount           int       `json:"amount"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

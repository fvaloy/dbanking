package model

import "time"

type Payment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    int       `json:"amount"`
	Currency  string    `json:"currency"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package repository

import (
	"database/sql"
	"fmt"

	"github.com/fvaloy/dbanking/ledger/internal/model"
	_ "github.com/lib/pq"
)

type LedgerRepository struct {
	db *sql.DB
}

type CreateLedgerEntryRequest struct {
	PaymentID string
	Type      string // "debit" or "credit"
	Amount    int
}

func NewLedgerRepository(db *sql.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) Create(
	req *CreateLedgerEntryRequest) (string, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO ledger (payment_id, type, amount)
		VALUES ($1, $2, $3) RETURNING id
	`, req.PaymentID, req.Type, req.Amount).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert ledger entry: %v", err)
	}
	return id, nil
}

func (r *LedgerRepository) GetByPaymentID(
	paymentID string) ([]model.LedgerEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, payment_id, type, amount, created_at
		FROM ledger WHERE payment_id = $1
		ORDER BY created_at DESC
	`, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ledger entries: %v", err)
	}
	defer rows.Close()

	var entries []model.LedgerEntry
	for rows.Next() {
		var entry model.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.PaymentID,
			&entry.Type,
			&entry.Amount,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %v", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning ledger entries: %v", err)
	}

	return entries, nil
}

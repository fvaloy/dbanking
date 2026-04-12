package repository

import (
	"database/sql"
	"fmt"

	"github.com/fvaloy/dbanking/payment/internal/model"
	_ "github.com/lib/pq"
)

type PaymentRepository struct {
	db *sql.DB
}

type CreatePaymentRequest struct {
	UserID    string
	Amount    int
	Currency  string
	Reference string
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(req *CreatePaymentRequest) (string, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO payments (user_id, amount, currency, reference, status)
		VALUES ($1, $2, $3, $4, 'pending') RETURNING id
	`, req.UserID, req.Amount, req.Currency, req.Reference).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert payment: %v", err)
	}
	return id, nil
}

func (r *PaymentRepository) GetByID(id string) (*model.Payment, error) {
	var p model.Payment
	err := r.db.QueryRow(`
		SELECT id, amount, currency, reference, status, created_at
		FROM payments WHERE id = $1
	`, id).Scan(&p.ID,
		&p.Amount,
		&p.Currency,
		&p.Reference,
		&p.Status,
		&p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found: %v", err)
		}
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}
	return &p, nil
}

func (r *PaymentRepository) ListByStatus(s string) ([]*model.Payment, error) {
	rows, err := r.db.Query(`
		SELECT id, amount, currency, reference, status, created_at
		FROM payments WHERE status = $1
	`, s)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %v", err)
	}
	defer rows.Close()

	var payments []*model.Payment
	for rows.Next() {
		var p model.Payment
		err := rows.Scan(&p.ID,
			&p.Amount,
			&p.Currency,
			&p.Reference,
			&p.Status,
			&p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %v", err)
		}
		payments = append(payments, &p)
	}
	return payments, nil
}

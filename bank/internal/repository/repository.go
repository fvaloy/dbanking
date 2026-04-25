package repository

import (
	"database/sql"
	"fmt"

	"github.com/fvaloy/dbanking/bank/internal/model"
	_ "github.com/lib/pq"
)

type BankRepository struct {
	db *sql.DB
}

type CreateBankMovementRequest struct {
	PaymentID        string
	PaymentReference string
	Amount           int
	Status           string
}

func NewBankRepository(db *sql.DB) *BankRepository {
	return &BankRepository{db: db}
}

func (r *BankRepository) CreatePending(
	req *CreateBankMovementRequest) (string, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO bank_movements (payment_reference, amount, status)
		VALUES ($1, $2, $3) RETURNING id
	`, req.PaymentReference, req.Amount, req.Status).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert bank movement: %v", err)
	}
	return id, nil
}

func (r *BankRepository) CompleteByReference(reference string) error {
	result, err := r.db.Exec(`
		UPDATE bank_movements
		SET status = 'completed'
		WHERE payment_reference = $1 AND status = 'pending'
	`, reference)
	if err != nil {
		return fmt.Errorf("failed to update bank movement status: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("bank movement not found or already completed: %s",
			reference)
	}
	return nil
}

func (r *BankRepository) GetByReference(
	reference string) (*model.BankMovement, error) {
	var m model.BankMovement
	err := r.db.QueryRow(`
		SELECT id, payment_reference, amount, status, created_at
		FROM bank_movements WHERE payment_reference = $1
	`, reference).Scan(
		&m.ID,
		&m.PaymentReference,
		&m.Amount,
		&m.Status,
		&m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bank movement not found: %v", err)
		}
		return nil, fmt.Errorf("failed to get bank movement: %v", err)
	}
	return &m, nil
}

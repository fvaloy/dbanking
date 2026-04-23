package service

import (
	"fmt"
	"log"

	"github.com/fvaloy/dbanking/ledger/internal/broker"
	"github.com/fvaloy/dbanking/ledger/internal/repository"
)

type LedgerService struct {
	repo *repository.LedgerRepository
}

func NewLedgerService(repo *repository.LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) ProcessPaymentCreatedEvent(
	event *broker.PaymentCreatedEvent) error {
	if event == nil {
		return fmt.Errorf("invalid event: event is nil")
	}

	if event.PaymentID == "" {
		return fmt.Errorf("invalid event: payment_id is empty")
	}

	if event.Amount <= 0 {
		return fmt.Errorf("invalid event: amount must be greater than 0")
	}

	debitEntryID, err := s.repo.Create(&repository.CreateLedgerEntryRequest{
		PaymentID: event.PaymentID,
		Type:      "debit",
		Amount:    event.Amount,
	})
	if err != nil {
		return fmt.Errorf("failed to create debit entry: %v", err)
	}

	log.Printf("Created debit entry: %s for payment: %s, amount: %d %s",
		debitEntryID, event.PaymentID, event.Amount, event.Currency)

	creditEntryID, err := s.repo.Create(&repository.CreateLedgerEntryRequest{
		PaymentID: event.PaymentID,
		Type:      "credit",
		Amount:    event.Amount,
	})
	if err != nil {
		return fmt.Errorf("failed to create credit entry: %v", err)
	}

	log.Printf("Created credit entry: %s for payment: %s, amount: %d %s",
		creditEntryID, event.PaymentID, event.Amount, event.Currency)

	return nil
}

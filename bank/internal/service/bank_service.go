package service

import (
	"context"
	"fmt"
	"log"

	"github.com/fvaloy/dbanking/bank/internal/broker"
	"github.com/fvaloy/dbanking/bank/internal/repository"
	"github.com/fvaloy/dbanking/bank/pb"
)

type BankService struct {
	repo          *repository.BankRepository
	broker        *broker.RabbitMQBroker
	paymentClient pb.PaymentServiceClient
}

func NewBankService(
	repo *repository.BankRepository,
	broker *broker.RabbitMQBroker,
	paymentClient pb.PaymentServiceClient) *BankService {
	return &BankService{
		repo:          repo,
		broker:        broker,
		paymentClient: paymentClient}
}

func (s *BankService) ProcessPaymentCreated(
	event *broker.PaymentCreatedEvent) error {
	if event == nil {
		return fmt.Errorf("invalid payment created event")
	}

	movementID, err := s.repo.CreatePending(
		&repository.CreateBankMovementRequest{
			PaymentID:        event.PaymentID,
			PaymentReference: event.Reference,
			Amount:           event.Amount,
			Status:           "pending",
		})
	if err != nil {
		return fmt.Errorf("failed to create bank movement: %w", err)
	}

	log.Printf("Created bank movement pending: %s for payment: %s",
		movementID,
		event.PaymentID)

	return s.broker.PublishBankMovement(&broker.BankMovementEvent{
		PaymentID: event.PaymentID,
		Reference: event.Reference,
		Amount:    event.Amount,
		Currency:  event.Currency,
		Status:    "pending",
	})
}

func (s *BankService) ProcessBankMovement(
	event *broker.BankMovementEvent) error {
	if event == nil {
		return fmt.Errorf("invalid bank movement event")
	}

	if err := s.repo.CompleteByReference(event.Reference); err != nil {
		return fmt.Errorf("failed to complete bank movement: %w", err)
	}

	log.Printf("Bank movement completed for payment: %s", event.PaymentID)

	_, err := s.paymentClient.MarkPaymentSucceeded(
		context.Background(),
		&pb.MarkPaymentSucceededRequest{
			PaymentId: event.PaymentID,
		})
	if err != nil {
		return fmt.Errorf("failed to mark payment succeeded: %w", err)
	}

	log.Printf("Called Payment.MarkPaymentSucceeded for payment: %s",
		event.PaymentID)
	return nil
}

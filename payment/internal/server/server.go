package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/fvaloy/dbanking/payment/internal/broker"
	"github.com/fvaloy/dbanking/payment/internal/repository"
	"github.com/fvaloy/dbanking/payment/pb"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	repo   *repository.PaymentRepository
	broker *broker.RabbitMQClient
}

func NewPaymentServer(
	repo *repository.PaymentRepository,
	brokerClient *broker.RabbitMQClient) *PaymentServer {
	return &PaymentServer{repo: repo, broker: brokerClient}
}

func (s *PaymentServer) CreatePayment(
	ctx context.Context,
	req *pb.CreatePaymentRequest) (*pb.PaymentResponse, error) {
	reference := fmt.Sprintf("REF-%d", rand.New(
		rand.NewSource(
			time.Now().UnixNano())).Intn(1000000))
	id, err := s.repo.Create(&repository.CreatePaymentRequest{
		UserID:    req.UserId,
		Amount:    int(req.Amount),
		Currency:  req.Currency,
		Reference: reference,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %v", err)
	}

	if s.broker != nil {
		event := &broker.PaymentEvent{
			PaymentID: id,
			UserID:    req.UserId,
			Amount:    int(req.Amount),
			Currency:  req.Currency,
			Reference: reference,
			Status:    "pending",
		}
		if err := s.broker.PublishPaymentCreated(event); err != nil {
			log.Printf("Warning: failed to publish payment event: %v", err)
		}
	}

	return &pb.PaymentResponse{
		PaymentId: id,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Reference: reference,
	}, nil
}

func (s *PaymentServer) GetPaymentByID(
	ctx context.Context,
	req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	p, err := s.repo.GetByID(req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}
	return &pb.PaymentResponse{
		PaymentId: p.ID,
		Status:    p.Status,
		Amount:    int32(p.Amount),
		Currency:  p.Currency,
		Reference: p.Reference,
	}, nil
}

func (s *PaymentServer) ListPaymentsByStatus(
	ctx context.Context,
	req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payments, err := s.repo.ListByStatus(req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %v", err)
	}
	resp := &pb.ListPaymentsResponse{}
	for _, p := range payments {
		resp.Payments = append(resp.Payments, &pb.PaymentResponse{
			PaymentId: p.ID,
			Status:    p.Status,
			Amount:    int32(p.Amount),
			Currency:  p.Currency,
			Reference: p.Reference,
		})
	}
	return resp, nil
}

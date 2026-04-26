package client

import (
	"context"
	"fmt"

	"github.com/fvaloy/dbanking/api-gateway/pb"
	"google.golang.org/grpc"
)

type PaymentClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentClient(conn *grpc.ClientConn) *PaymentClient {
	return &PaymentClient{client: pb.NewPaymentServiceClient(conn)}
}

func (p *PaymentClient) CreatePayment(ctx context.Context, userID string, amount int, currency string) (*pb.PaymentResponse, error) {
	resp, err := p.client.CreatePayment(ctx, &pb.CreatePaymentRequest{
		UserId:   userID,
		Amount:   int32(amount),
		Currency: currency,
	})
	if err != nil {
		return nil, fmt.Errorf("create payment failed: %w", err)
	}
	return resp, nil
}

func (p *PaymentClient) GetPaymentByID(ctx context.Context, paymentID string) (*pb.PaymentResponse, error) {
	resp, err := p.client.GetPaymentByID(ctx, &pb.GetPaymentRequest{PaymentId: paymentID})
	if err != nil {
		return nil, fmt.Errorf("get payment failed: %w", err)
	}
	return resp, nil
}

func (p *PaymentClient) ListPaymentsByStatus(ctx context.Context, status string) ([]*pb.PaymentResponse, error) {
	resp, err := p.client.ListPaymentsByStatus(ctx, &pb.ListPaymentsRequest{Status: status})
	if err != nil {
		return nil, fmt.Errorf("list payments failed: %w", err)
	}
	return resp.Payments, nil
}

package broker

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	PaymentID string `json:"payment_id"`
	UserID    string `json:"user_id"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
}

type RabbitMQClient struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

const (
	PaymentExchange = "payments"
)

func NewRabbitMQClient(
	rabbitURL,
	exchange string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		exchange,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQClient{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
	}, nil
}

func (r *RabbitMQClient) PublishPaymentCreated(event *PaymentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.channel.Publish(
		r.exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published payment created event: %s", event.PaymentID)
	return nil
}

func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	return nil
}

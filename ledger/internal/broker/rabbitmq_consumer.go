package broker

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCreatedEvent struct {
	PaymentID string `json:"payment_id"`
	UserID    string `json:"user_id"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
}

type PaymentEventHandler func(*PaymentCreatedEvent) error

type RabbitMQConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	exchange  string
	queueName string
}

const (
	PaymentCreatedRoutingKey = "payment.created"
	PaymentExchange          = "payments"
)

func NewRabbitMQConsumer(
	rabbitURL,
	exchange,
	queueName string) (*RabbitMQConsumer, error) {
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
		amqp.ExchangeTopic,
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

	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = channel.QueueBind(
		queue.Name,
		PaymentCreatedRoutingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQConsumer{
		conn:      conn,
		channel:   channel,
		queue:     queue,
		exchange:  exchange,
		queueName: queueName,
	}, nil
}

func (r *RabbitMQConsumer) StartConsuming(handler PaymentEventHandler) error {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"ledger",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	log.Printf("Ledger service started consuming from queue: %s", r.queueName)

	go func() {
		for msg := range msgs {
			var event PaymentCreatedEvent
			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				msg.Nack(false, true)
				continue
			}

			log.Printf("Processing payment event: %s", event.PaymentID)
			err = handler(&event)
			if err != nil {
				log.Printf("Error processing payment event: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

func (r *RabbitMQConsumer) Close() error {
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

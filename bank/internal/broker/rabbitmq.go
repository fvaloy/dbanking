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

type BankMovementEvent struct {
	PaymentID string `json:"payment_id"`
	Reference string `json:"payment_reference"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
}

type RabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

const (
	PaymentExchange          = "payments"
	PaymentCreatedRoutingKey = "payment.created"
	BankExchange             = "bank"
	BankMovementRoutingKey   = "bank.movements"
	PaymentQueueName         = "bank.payments"
	BankMovementQueueName    = "bank.movements"
)

func NewRabbitMQBroker(rabbitURL string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := channel.ExchangeDeclare(
		PaymentExchange,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare payment exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		PaymentQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare payment queue: %w", err)
	}

	if err := channel.QueueBind(
		PaymentQueueName,
		PaymentCreatedRoutingKey,
		PaymentExchange,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind payment queue: %w", err)
	}

	if err := channel.ExchangeDeclare(
		BankExchange,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare bank exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		BankMovementQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare bank movement queue: %w", err)
	}

	if err := channel.QueueBind(
		BankMovementQueueName,
		BankMovementRoutingKey,
		BankExchange,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind bank movement queue: %w", err)
	}

	return &RabbitMQBroker{conn: conn, channel: channel}, nil
}

func (b *RabbitMQBroker) StartConsumingPaymentCreated(
	handler func(*PaymentCreatedEvent) error) error {
	msgs, err := b.channel.Consume(
		PaymentQueueName,
		"bank-payment-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume payment created messages: %w", err)
	}

	log.Printf("Bank service consuming payment created events from queue: %s",
		PaymentQueueName)

	go func() {
		for msg := range msgs {
			var event PaymentCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Error decoding payment created event: %v", err)
				msg.Nack(false, true)
				continue
			}

			if err := handler(&event); err != nil {
				log.Printf("Error processing payment created event: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

func (b *RabbitMQBroker) StartConsumingBankMovements(
	handler func(*BankMovementEvent) error) error {
	msgs, err := b.channel.Consume(
		BankMovementQueueName,
		"bank-movement-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume bank movement messages: %w", err)
	}

	log.Printf("Bank service consuming bank movement events from queue: %s",
		BankMovementQueueName)

	go func() {
		for msg := range msgs {
			var event BankMovementEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Error decoding bank movement event: %v", err)
				msg.Nack(false, true)
				continue
			}

			if err := handler(&event); err != nil {
				log.Printf("Error processing bank movement event: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

func (b *RabbitMQBroker) PublishBankMovement(event *BankMovementEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal bank movement event: %w", err)
	}

	if err := b.channel.Publish(
		BankExchange,
		BankMovementRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish bank movement event: %w", err)
	}

	log.Printf("Published bank movement event for payment: %s", event.PaymentID)
	return nil
}

func (b *RabbitMQBroker) Close() error {
	if b.channel != nil {
		if err := b.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	return nil
}

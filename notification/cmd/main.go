package main

import (
	"log"
	"os"
	"time"

	"github.com/fvaloy/dbanking/notification/internal/broker"
	"github.com/fvaloy/dbanking/notification/internal/email"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (using system env)")
	}
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	consumer, err := connectRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer consumer.Close()

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	emailSender := email.NewEmailSender(smtpHost, smtpPort)

	err = consumer.StartConsuming(func(event *broker.PaymentSucceededEvent) error {
		return emailSender.SendPaymentSucceeded(event)
	})
	if err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}
	log.Println("Notification service is running and consuming messages...")
	select {}
}

func connectRabbitMQ(rabbitURL string) (*broker.RabbitMQConsumer, error) {
	var brokerClient *broker.RabbitMQConsumer
	var err error
	for i := range 10 {
		brokerClient, err = broker.NewRabbitMQConsumer(rabbitURL, "payments", "notification.payments")
		if err == nil {
			log.Println("Connected to RabbitMQ successfully")
			return brokerClient, nil
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

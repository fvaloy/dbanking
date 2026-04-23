package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/fvaloy/dbanking/ledger/internal/broker"
	"github.com/fvaloy/dbanking/ledger/internal/repository"
	"github.com/fvaloy/dbanking/ledger/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (using system env)")
	}

	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		log.Fatal("DB_CONN is required")
	}

	var db *sql.DB
	var errDb error
	for i := 0; i < 5; i++ {
		db, errDb = sql.Open("postgres", connStr)
		if errDb == nil && db.Ping() == nil {
			break
		}
		log.Println("Waiting for database connection...")
		time.Sleep(2 * time.Second)
	}
	if errDb != nil {
		log.Fatal(errDb)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database successfully")

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	consumer, err := connectRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer consumer.Close()

	repo := repository.NewLedgerRepository(db)
	ledgerService := service.NewLedgerService(repo)

	err = consumer.StartConsuming(func(event *broker.PaymentCreatedEvent) error {
		return ledgerService.ProcessPaymentCreatedEvent(event)
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Ledger Service is running and consuming payment events...")

	select {}
}

func connectRabbitMQ(rabbitURL string) (*broker.RabbitMQConsumer, error) {
	var consumer *broker.RabbitMQConsumer
	var err error
	for i := 0; i < 10; i++ {
		consumer, err = broker.NewRabbitMQConsumer(rabbitURL, "payments", "payment.created")
		if err == nil {
			log.Println("Connected to RabbitMQ successfully")
			return consumer, nil
		}
		log.Printf("Waiting for RabbitMQ connection (%d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

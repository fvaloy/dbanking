package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/fvaloy/dbanking/bank/internal/broker"
	"github.com/fvaloy/dbanking/bank/internal/repository"
	"github.com/fvaloy/dbanking/bank/internal/service"
	"github.com/fvaloy/dbanking/bank/pb"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	paymentAddr := os.Getenv("PAYMENT_SERVICE_ADDR")
	if paymentAddr == "" {
		paymentAddr = "payment:8080"
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
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

	brokerClient, err := connectRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer brokerClient.Close()

	grpcConn, err := grpc.NewClient(
		paymentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial payment service: %v", err)
	}
	defer grpcConn.Close()

	paymentClient := pb.NewPaymentServiceClient(grpcConn)

	repo := repository.NewBankRepository(db)
	bankService := service.NewBankService(repo, brokerClient, paymentClient)

	if err := brokerClient.StartConsumingPaymentCreated(
		bankService.ProcessPaymentCreated); err != nil {
		log.Fatalf("failed to start payment created consumer: %v", err)
	}

	if err := brokerClient.StartConsumingBankMovements(
		bankService.ProcessBankMovement); err != nil {
		log.Fatalf("failed to start bank movement consumer: %v", err)
	}

	log.Println("Bank service is running")

	select {}
}

func connectRabbitMQ(rabbitURL string) (*broker.RabbitMQBroker, error) {
	var brokerClient *broker.RabbitMQBroker
	var err error
	for i := range 10 {
		brokerClient, err = broker.NewRabbitMQBroker(rabbitURL)
		if err == nil {
			return brokerClient, nil
		}
		log.Printf("Waiting for RabbitMQ connection (%d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

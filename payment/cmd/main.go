package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"github.com/fvaloy/dbanking/payment/internal/repository"
	"github.com/fvaloy/dbanking/payment/internal/server"
	"github.com/fvaloy/dbanking/payment/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPaymentRepository(db)
	server := server.NewPaymentServer(repo)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPaymentServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Printf("Payment Service listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/fvaloy/dbanking/api-gateway/docs"
	"github.com/fvaloy/dbanking/api-gateway/internal/client"
	"github.com/fvaloy/dbanking/api-gateway/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (using system env)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	paymentAddr := os.Getenv("PAYMENT_SERVICE_ADDR")
	if paymentAddr == "" {
		paymentAddr = "payment:8080"
	}

	grpcConn, err := grpc.NewClient(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to payment service: %v", err)
	}
	defer grpcConn.Close()

	paymentClient := client.NewPaymentClient(grpcConn)
	handler := handler.NewPaymentHandler(paymentClient)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/payments", handler.CreatePayment)
	router.GET("/payments", handler.ListPaymentsByStatus)
	router.GET("/payments/:id", handler.GetPaymentByID)
	router.POST("/stress-test", handler.StressTest)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API Gateway is running"})
	})

	log.Printf("API Gateway listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run api gateway: %v", err)
	}
}

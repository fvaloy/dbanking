package handler

import (
	"context"
	"log"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fvaloy/dbanking/api-gateway/internal/client"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentClient *client.PaymentClient
}

func NewPaymentHandler(paymentClient *client.PaymentClient) *PaymentHandler {
	return &PaymentHandler{paymentClient: paymentClient}
}

// CreatePaymentRequest represents the request body for creating a payment.
// swagger:model CreatePaymentRequest
type CreatePaymentRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Amount   int    `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required"`
}

// StressTestRequest represents the request body for the stress test endpoint.
// swagger:model StressTestRequest
type StressTestRequest struct {
	DurationSeconds int `json:"duration_seconds" binding:"required,gt=0"`
	Concurrency     int `json:"concurrency" binding:"omitempty,gt=1"`
}

// StressTestResponse represents the result of a stress test.
// swagger:model StressTestResponse
type StressTestResponse struct {
	DurationSeconds int   `json:"duration_seconds"`
	Concurrency     int   `json:"concurrency"`
	TotalRequests   int64 `json:"total_requests"`
	Successes       int64 `json:"successes"`
	Failures        int64 `json:"failures"`
}

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a payment in the Payment microservice
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Create payment request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.paymentClient.CreatePayment(c.Request.Context(), req.UserID, req.Amount, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPaymentByID godoc
// @Summary Get a payment by ID
// @Description Retrieve payment details by payment ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPaymentByID(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	resp, err := h.paymentClient.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListPaymentsByStatus godoc
// @Summary List payments by status
// @Description List payments filtered by payment status
// @Tags payments
// @Accept json
// @Produce json
// @Param status query string false "Payment status"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments [get]
func (h *PaymentHandler) ListPaymentsByStatus(c *gin.Context) {
	status := c.Query("status")
	resp, err := h.paymentClient.ListPaymentsByStatus(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": resp})
}

// StressTest godoc
// @Summary Run a payment stress test
// @Description Execute concurrent CreatePayment calls for load testing
// @Tags payments
// @Accept json
// @Produce json
// @Param request body StressTestRequest true "Stress test request"
// @Success 200 {object} StressTestResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /stress-test [post]
func (h *PaymentHandler) StressTest(c *gin.Context) {
	var req StressTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(req.DurationSeconds)*time.Second)
	defer cancel()

	var total int64
	var success int64
	var failures int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					atomic.AddInt64(&total, 1)
					callCtx, callCancel := context.WithTimeout(ctx, 10*time.Second)
					_, err := h.paymentClient.CreatePayment(callCtx, "f30728f9-c6ba-4506-a0eb-bc2f4c70acb3", 1000, "USD")
					callCancel()
					if err != nil {
						atomic.AddInt64(&failures, 1)
						log.Printf("stress create failed: %v", err)
						continue
					}
					atomic.AddInt64(&success, 1)
				}
			}
		})
	}

	wg.Wait()

	c.JSON(http.StatusOK, StressTestResponse{
		DurationSeconds: req.DurationSeconds,
		Concurrency:     concurrency,
		TotalRequests:   total,
		Successes:       success,
		Failures:        failures,
	})
}

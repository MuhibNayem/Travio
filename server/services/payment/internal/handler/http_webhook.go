package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// IPNHandler handles Instant Payment Notification callbacks from payment gateways
type IPNHandler struct {
	paymentService *service.PaymentService
	registry       *gateway.Registry
	repo           *repository.TransactionRepository
	refundRepo     *repository.RefundRepository
	configRepo     *repository.PaymentConfigRepository
	redisClient    *redis.Client
	kafkaProducer  *kafka.Producer
}

// NewIPNHandler creates a new IPN handler
func NewIPNHandler(
	paymentService *service.PaymentService,
	registry *gateway.Registry,
	repo *repository.TransactionRepository,
	refundRepo *repository.RefundRepository,
	configRepo *repository.PaymentConfigRepository,
	redisClient *redis.Client,
	kafkaProducer *kafka.Producer,
) *IPNHandler {
	return &IPNHandler{
		paymentService: paymentService,
		registry:       registry,
		repo:           repo,
		refundRepo:     refundRepo,
		configRepo:     configRepo,
		redisClient:    redisClient,
		kafkaProducer:  kafkaProducer,
	}
}

// HandleIPN handles POST /v1/payments/ipn/{gateway}
// This endpoint receives callbacks from payment gateways (SSLCommerz, bKash, Nagad)
func (h *IPNHandler) HandleIPN(w http.ResponseWriter, r *http.Request) {
	gatewayName := chi.URLParam(r, "gateway")
	if gatewayName == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "gateway parameter required"})
		return
	}

	// Parse the request body
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Convert payload to bytes for gateway validation
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to marshal payload"})
		return
	}

	// Get the gateway factory
	providerName := h.registry.ResolveProvider(gatewayName)
	factory, err := h.registry.GetFactory(providerName)
	if err != nil {
		logger.Error("Unknown gateway for IPN", "gateway", gatewayName, "error", err)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown gateway"})
		return
	}

	// Extract order ID from payload
	payloadStr := make(map[string]string)
	for k, v := range payload {
		if s, ok := v.(string); ok {
			payloadStr[k] = s
		}
	}
	
	orderID, err := factory.ParseOrderID(payloadStr)
	if err != nil {
		logger.Error("Failed to parse order ID from IPN", "gateway", gatewayName, "error", err)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	// Check for duplicate IPN (idempotency) - TASK-020
	idempotencyKey := fmt.Sprintf("ipn:%s:%s", gatewayName, orderID)
	if h.redisClient != nil {
		exists, err := h.redisClient.SetNX(r.Context(), idempotencyKey, "processing", 5*time.Minute).Result()
		if err != nil {
			logger.Warn("Redis check failed for IPN idempotency", "error", err)
			// Proceed anyway - fail open for idempotency
		} else if !exists {
			// Already processed this IPN
			logger.Info("Duplicate IPN detected, returning cached response", "order_id", orderID)
			respondJSON(w, http.StatusOK, map[string]string{
				"status":  "duplicate",
				"message": "IPN already processed",
			})
			return
		}
	}

	// Get the transaction
	tx, err := h.repo.GetByOrderID(r.Context(), orderID)
	if err != nil {
		logger.Error("Transaction not found for IPN", "order_id", orderID, "error", err)
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "transaction not found"})
		return
	}

	// Get payment config for the organization
	payConfig, err := h.configRepo.GetConfig(r.Context(), tx.OrganizationID)
	if err != nil {
		logger.Error("Failed to get payment config for IPN", "org_id", tx.OrganizationID, "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Create gateway instance
	gw, err := factory.Create(payConfig.Credentials, false)
	if err != nil {
		logger.Error("Failed to create gateway for IPN validation", "gateway", gatewayName, "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "gateway creation failed"})
		return
	}

	// Validate IPN signature
	ipnData, err := gw.ValidateIPN(r.Context(), payloadBytes)
	if err != nil {
		logger.Error("IPN validation failed", "order_id", orderID, "gateway", gatewayName, "error", err)
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"valid":   "false",
			"reason":  "invalid signature",
		})
		return
	}

	if !ipnData.IsValid {
		logger.Warn("IPN signature invalid", "order_id", orderID, "gateway", gatewayName)
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"valid":  "false",
			"reason": "signature verification failed",
		})
		return
	}

	// Update transaction status based on IPN data
	newStatus := mapGatewayStatus(ipnData.Status)
	if err := h.repo.UpdateStatus(r.Context(), tx.ID, string(newStatus), ipnData.GatewayRef); err != nil {
		logger.Error("Failed to update transaction status from IPN", "order_id", orderID, "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update status"})
		return
	}

	// Publish payment event to Kafka (TASK-019)
	if h.kafkaProducer != nil {
		eventType := kafka.EventOrderConfirmed
		if newStatus == gateway.StatusFailed || newStatus == gateway.StatusCancelled {
			eventType = kafka.EventPaymentFailed
		}

		event := &kafka.Event{
			ID:          fmt.Sprintf("evt-%s-%d", orderID, time.Now().UnixNano()),
			Type:        eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Payload: map[string]interface{}{
				"order_id":       orderID,
				"transaction_id": tx.ID,
				"gateway_ref":    ipnData.GatewayRef,
				"bank_tran_id":   ipnData.BankTranID,
				"status":         string(newStatus),
				"amount_paisa":   ipnData.AmountPaisa,
				"card_type":      ipnData.CardType,
				"risk_level":     ipnData.RiskLevel,
			},
		}

		if err := h.kafkaProducer.Publish(r.Context(), kafka.TopicOrders, event); err != nil {
			logger.Error("Failed to publish payment event", "order_id", orderID, "error", err)
			// Don't fail the IPN - status update is more important
		}
	}

	// Update idempotency key with success status
	if h.redisClient != nil {
		h.redisClient.Set(r.Context(), idempotencyKey, "completed", 24*time.Hour)
	}

	logger.Info("IPN processed successfully",
		"order_id", orderID,
		"gateway", gatewayName,
		"status", newStatus,
		"gateway_ref", ipnData.GatewayRef,
	)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":          true,
		"transaction_id": tx.ID,
		"order_id":       orderID,
		"status":         string(newStatus),
	})
}

// HandleIPNReturn handles GET requests (some gateways use GET for return URLs)
func (h *IPNHandler) HandleIPNReturn(w http.ResponseWriter, r *http.Request) {
	gatewayName := chi.URLParam(r, "gateway")
	orderID := r.URL.Query().Get("order_id")
	
	if orderID == "" {
		http.Redirect(w, r, "/payment/return?status=error&message=missing_order_id", http.StatusSeeOther)
		return
	}

	// Redirect to frontend payment return page
	http.Redirect(w, r, fmt.Sprintf("/payment/return?gateway=%s&order_id=%s", gatewayName, orderID), http.StatusSeeOther)
}

// GetRefundStatus handles GET /v1/payments/{orderId}/refund (TASK-021)
func (h *IPNHandler) GetRefundStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	if orderID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "orderId parameter required"})
		return
	}

	refund, err := h.refundRepo.GetByOrderID(r.Context(), orderID)
	if err != nil {
		if err == repository.ErrRefundNotFound {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "refund not found"})
			return
		}
		logger.Error("Failed to get refund status", "order_id", orderID, "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	respondJSON(w, http.StatusOK, refund)
}

// RegisterRoutes registers IPN routes
func (h *IPNHandler) RegisterRoutes(r chi.Router) {
	r.Post("/{gateway}", h.HandleIPN)
	r.Get("/{gateway}", h.HandleIPNReturn)
	r.Get("/{orderId}/refund", h.GetRefundStatus)
}

// --- Helpers ---

func mapGatewayStatus(status gateway.Status) gateway.Status {
	// Normalize various gateway statuses to our internal status
	switch status {
	case "VALID", "SUCCESS", "COMPLETED", "CAPTURED":
		return gateway.StatusCaptured
	case "FAILED", "CANCELLED", "EXPIRED", "DECLINED":
		return gateway.StatusFailed
	case "PENDING", "INITIATED", "PROCESSING":
		return gateway.StatusProcessing
	case "REFUNDED":
		return gateway.StatusRefunded
	default:
		return gateway.StatusPending
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}

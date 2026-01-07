package handler

import (
	"encoding/json"
	"net/http"

	paymentv1 "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

// PaymentHandler handles payment-related requests via gRPC
type PaymentHandler struct {
	client *client.PaymentClient
}

// NewPaymentHandler creates a new payment handler with gRPC client
func NewPaymentHandler(paymentClient *client.PaymentClient) *PaymentHandler {
	return &PaymentHandler{client: paymentClient}
}

// GetPaymentStatus returns payment status for an order
func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")

	resp, err := h.client.VerifyPayment(r.Context(), &paymentv1.VerifyPaymentRequest{
		TransactionId: orderID,
	})
	if err != nil {
		logger.Error("Failed to get payment status", "error", err)
		http.Error(w, `{"error": "payment service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ProcessPayment initiates a payment via gRPC
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req paymentv1.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreatePayment(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to process payment", "error", err)
		http.Error(w, `{"error": "payment service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPaymentMethods returns available payment methods
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	methods := []map[string]interface{}{
		{"id": "card", "name": "Credit/Debit Card", "enabled": true},
		{"id": "bkash", "name": "bKash", "enabled": true},
		{"id": "nagad", "name": "Nagad", "enabled": true},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"methods": methods})
}

// UpdatePaymentConfig handles updating payment configuration
func (h *PaymentHandler) UpdatePaymentConfig(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	var req paymentv1.UpdatePaymentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.OrganizationId = orgID // Ensure ID from URL is used

	resp, err := h.client.UpdatePaymentConfig(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to update payment config", "org_id", orgID, "error", err)
		http.Error(w, `{"error": "failed to update config"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPaymentConfig handles retrieving payment configuration
func (h *PaymentHandler) GetPaymentConfig(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	resp, err := h.client.GetPaymentConfig(r.Context(), &paymentv1.GetPaymentConfigRequest{
		OrganizationId: orgID,
	})
	if err != nil {
		logger.Error("Failed to get payment config", "org_id", orgID, "error", err)
		http.Error(w, `{"error": "config not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

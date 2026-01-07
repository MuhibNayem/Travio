package handler

import (
	"encoding/json"
	"net/http"

	pricingv1 "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
)

// PricingHandler handles pricing requests via gRPC
type PricingHandler struct {
	client *client.PricingClient
}

// NewPricingHandler creates a new pricing handler with gRPC client
func NewPricingHandler(pricingClient *client.PricingClient) *PricingHandler {
	return &PricingHandler{client: pricingClient}
}

// CalculatePriceRequest is the HTTP request body for price calculation
type CalculatePriceRequest struct {
	TripID         string  `json:"trip_id"`
	SeatClass      string  `json:"seat_class"`
	Date           string  `json:"date"`
	Quantity       int32   `json:"quantity"`
	BasePricePaisa int64   `json:"base_price_paisa"`
	OccupancyRate  float64 `json:"occupancy_rate"`
}

// CalculatePrice calculates dynamic price via gRPC
func (h *PricingHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	var req CalculatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	grpcReq := &pricingv1.CalculatePriceRequest{
		TripId:         req.TripID,
		SeatClass:      req.SeatClass,
		Date:           req.Date,
		Quantity:       req.Quantity,
		BasePricePaisa: req.BasePricePaisa,
		OccupancyRate:  req.OccupancyRate,
	}

	resp, err := h.client.CalculatePrice(r.Context(), grpcReq)
	if err != nil {
		logger.Error("Failed to calculate price", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPricingRules returns all pricing rules via gRPC
func (h *PricingHandler) GetPricingRules(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetRules(r.Context(), false)
	if err != nil {
		logger.Error("Failed to get pricing rules", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

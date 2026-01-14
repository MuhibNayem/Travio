package handler

import (
	"encoding/json"
	"net/http"

	pricingv1 "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
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
	SeatCategory   string  `json:"seat_category"`
	Date           string  `json:"date"`
	Quantity       int32   `json:"quantity"`
	BasePricePaisa int64   `json:"base_price_paisa"`
	OccupancyRate  float64 `json:"occupancy_rate"`
	OrganizationID string  `json:"organization_id"`
	DepartureTime  int64   `json:"departure_time"`
	RouteID        string  `json:"route_id"`
	ScheduleID     string  `json:"schedule_id"`
	FromStationID  string  `json:"from_station_id"`
	ToStationID    string  `json:"to_station_id"`
	VehicleType    string  `json:"vehicle_type"`
	VehicleClass   string  `json:"vehicle_class"`
	PromoCode      string  `json:"promo_code"`
}

type PricingRuleRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Condition       string  `json:"condition"`
	Multiplier      float64 `json:"multiplier"`
	AdjustmentType  string  `json:"adjustment_type"`
	AdjustmentValue float64 `json:"adjustment_value"`
	Priority        int32   `json:"priority"`
	IsActive        bool    `json:"is_active"`
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
		SeatCategory:   req.SeatCategory,
		Date:           req.Date,
		Quantity:       req.Quantity,
		BasePricePaisa: req.BasePricePaisa,
		OccupancyRate:  req.OccupancyRate,
		OrganizationId: req.OrganizationID,
		DepartureTime:  req.DepartureTime,
		RouteId:        req.RouteID,
		ScheduleId:     req.ScheduleID,
		FromStationId:  req.FromStationID,
		ToStationId:    req.ToStationID,
		VehicleType:    req.VehicleType,
		VehicleClass:   req.VehicleClass,
		PromoCode:      req.PromoCode,
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
	orgID := middleware.GetOrgID(r.Context())
	resp, err := h.client.GetRules(r.Context(), r.URL.Query().Get("include_inactive") == "true", orgID)
	if err != nil {
		logger.Error("Failed to get pricing rules", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PricingHandler) CreatePricingRule(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	var req PricingRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateRule(r.Context(), &pricingv1.CreateRuleRequest{
		OrganizationId:  orgID,
		Name:            req.Name,
		Description:     req.Description,
		Condition:       req.Condition,
		Multiplier:      req.Multiplier,
		AdjustmentType:  req.AdjustmentType,
		AdjustmentValue: req.AdjustmentValue,
		Priority:        req.Priority,
	})
	if err != nil {
		logger.Error("Failed to create pricing rule", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *PricingHandler) UpdatePricingRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, `{"error": "rule_id is required"}`, http.StatusBadRequest)
		return
	}

	var req PricingRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdateRule(r.Context(), &pricingv1.UpdateRuleRequest{
		Id:              ruleID,
		Name:            req.Name,
		Description:     req.Description,
		Condition:       req.Condition,
		Multiplier:      req.Multiplier,
		AdjustmentType:  req.AdjustmentType,
		AdjustmentValue: req.AdjustmentValue,
		Priority:        req.Priority,
		IsActive:        req.IsActive,
	})
	if err != nil {
		logger.Error("Failed to update pricing rule", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PricingHandler) DeletePricingRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, `{"error": "rule_id is required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.DeleteRule(r.Context(), &pricingv1.DeleteRuleRequest{Id: ruleID})
	if err != nil {
		logger.Error("Failed to delete pricing rule", "error", err)
		http.Error(w, `{"error": "pricing service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

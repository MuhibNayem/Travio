package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/order/internal/service"
	"github.com/go-chi/chi/v5"
)

// CheckoutHandler handles HTTP requests for checkout operations
type CheckoutHandler struct {
	checkoutService *service.CheckoutService
}

// NewCheckoutHandler creates a new checkout HTTP handler
func NewCheckoutHandler(checkoutService *service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
	}
}

// CreateCheckout handles POST /v1/checkout
func (h *CheckoutHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TripID        string                            `json:"trip_id"`
		HoldID        string                            `json:"hold_id"`
		FromStationID string                            `json:"from_station_id"`
		ToStationID   string                            `json:"to_station_id"`
		Passengers    []service.CheckoutPassengerInput  `json:"passengers"`
		CouponCode    string                            `json:"coupon_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract user context from middleware
	orgID := r.Context().Value("org_id")
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	checkoutReq := &service.CreateCheckoutRequest{
		OrganizationID: toString(orgID),
		UserID:         toString(userID),
		TripID:         req.TripID,
		HoldID:         req.HoldID,
		FromStationID:  req.FromStationID,
		ToStationID:    req.ToStationID,
		Passengers:     req.Passengers,
		CouponCode:     req.CouponCode,
	}

	session, err := h.checkoutService.CreateCheckout(r.Context(), checkoutReq)
	if err != nil {
		logger.Error("Failed to create checkout session", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// GetCheckout handles GET /v1/checkout/{id}
func (h *CheckoutHandler) GetCheckout(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	session, err := h.checkoutService.GetCheckout(r.Context(), sessionID, toString(userID))
	if err != nil {
		if err.Error() == "checkout session has expired" {
			respondError(w, http.StatusGone, "checkout session has expired")
			return
		}
		respondError(w, http.StatusNotFound, "checkout session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// GetCheckoutByHoldID handles GET /v1/checkout/hold/{holdId}
func (h *CheckoutHandler) GetCheckoutByHoldID(w http.ResponseWriter, r *http.Request) {
	holdID := chi.URLParam(r, "holdId")
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	session, err := h.checkoutService.GetCheckoutByHoldID(r.Context(), holdID, toString(userID))
	if err != nil {
		respondError(w, http.StatusNotFound, "checkout session not found")
		return
	}

	if session == nil {
		respondError(w, http.StatusNotFound, "checkout session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// UpdateCheckout handles PATCH /v1/checkout/{id}
func (h *CheckoutHandler) UpdateCheckout(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req service.UpdateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session, err := h.checkoutService.UpdateCheckout(r.Context(), sessionID, toString(userID), &req)
	if err != nil {
		logger.Error("Failed to update checkout session", "error", err)
		if err.Error() == "checkout session has expired" {
			respondError(w, http.StatusGone, "checkout session has expired")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// ConfirmCheckout handles POST /v1/checkout/{id}/confirm
func (h *CheckoutHandler) ConfirmCheckout(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	session, err := h.checkoutService.ConfirmCheckout(r.Context(), sessionID, toString(userID))
	if err != nil {
		logger.Error("Failed to confirm checkout session", "error", err)
		if err.Error() == "checkout session has expired" {
			respondError(w, http.StatusGone, "checkout session has expired")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// ListCheckouts handles GET /v1/checkout
func (h *CheckoutHandler) ListCheckouts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse query parameters
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	sessions, total, err := h.checkoutService.ListCheckouts(r.Context(), toString(userID), status, limit, offset)
	if err != nil {
		logger.Error("Failed to list checkout sessions", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list checkout sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// RegisterRoutes registers all checkout routes
func (h *CheckoutHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateCheckout)
	r.Get("/", h.ListCheckouts)
	r.Get("/{id}", h.GetCheckout)
	r.Get("/hold/{holdId}", h.GetCheckoutByHoldID)
	r.Patch("/{id}", h.UpdateCheckout)
	r.Post("/{id}/confirm", h.ConfirmCheckout)
}

// --- Helpers ---

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/service"
)

// HTTPHandler provides HTTP endpoints for the Pricing Service
type HTTPHandler struct {
	svc *service.PricingService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc *service.PricingService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

// RegisterRoutes registers HTTP routes
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/pricing/calculate", h.handleCalculatePrice)
	mux.HandleFunc("/api/v1/pricing/rules", h.handleRules)
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *HTTPHandler) handleCalculatePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcHandler := &GRPCHandler{svc: h.svc}
	resp, err := grpcHandler.CalculatePrice(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to calculate price", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HTTPHandler) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRules(w, r)
	case http.MethodPost:
		h.createRule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) getRules(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	organizationID := r.URL.Query().Get("organization_id")

	rules, err := h.svc.GetRules(r.Context(), includeInactive, organizationID)
	if err != nil {
		logger.Error("Failed to get rules", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rules": rules})
}

func (h *HTTPHandler) createRule(w http.ResponseWriter, r *http.Request) {
	var rule repository.PricingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateRule(r.Context(), &rule); err != nil {
		logger.Error("Failed to create rule", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

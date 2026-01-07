package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	operatorv1 "github.com/MuhibNayem/Travio/server/api/proto/operator/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
)

// OperatorHandler handles operator requests via gRPC
type OperatorHandler struct {
	client *client.OperatorClient
}

// NewOperatorHandler creates a new operator handler with gRPC client
func NewOperatorHandler(operatorClient *client.OperatorClient) *OperatorHandler {
	return &OperatorHandler{client: operatorClient}
}

// CreateVendorRequest is the HTTP request body for creating a vendor
type CreateVendorRequest struct {
	Name           string  `json:"name"`
	ContactEmail   string  `json:"contact_email"`
	ContactPhone   string  `json:"contact_phone"`
	Address        string  `json:"address"`
	CommissionRate float64 `json:"commission_rate"`
}

// CreateVendor creates a new vendor via gRPC
func (h *OperatorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var req CreateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	grpcReq := &operatorv1.CreateVendorRequest{
		Name:           req.Name,
		ContactEmail:   req.ContactEmail,
		ContactPhone:   req.ContactPhone,
		Address:        req.Address,
		CommissionRate: req.CommissionRate,
	}

	resp, err := h.client.CreateVendor(r.Context(), grpcReq)
	if err != nil {
		logger.Error("Failed to create vendor", "error", err)
		http.Error(w, `{"error": "failed to create vendor"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetVendor retrieves a vendor by ID via gRPC
func (h *OperatorHandler) GetVendor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error": "missing vendor id"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetVendor(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get vendor", "id", id, "error", err)
		http.Error(w, `{"error": "vendor not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateVendorRequest is the HTTP request body for updating a vendor
type UpdateVendorRequest struct {
	Name           string  `json:"name"`
	ContactEmail   string  `json:"contact_email"`
	ContactPhone   string  `json:"contact_phone"`
	Address        string  `json:"address"`
	Status         string  `json:"status"`
	CommissionRate float64 `json:"commission_rate"`
}

// UpdateVendor updates a vendor via gRPC
func (h *OperatorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error": "missing vendor id"}`, http.StatusBadRequest)
		return
	}

	var req UpdateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	grpcReq := &operatorv1.UpdateVendorRequest{
		Id:             id,
		Name:           req.Name,
		ContactEmail:   req.ContactEmail,
		ContactPhone:   req.ContactPhone,
		Address:        req.Address,
		Status:         req.Status,
		CommissionRate: req.CommissionRate,
	}

	resp, err := h.client.UpdateVendor(r.Context(), grpcReq)
	if err != nil {
		logger.Error("Failed to update vendor", "id", id, "error", err)
		http.Error(w, `{"error": "failed to update vendor"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListVendors lists vendors via gRPC
func (h *OperatorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.client.ListVendors(r.Context(), int32(page), int32(limit))
	if err != nil {
		logger.Error("Failed to list vendors", "error", err)
		http.Error(w, `{"error": "failed to list vendors"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteVendor deletes a vendor via gRPC
func (h *OperatorHandler) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error": "missing vendor id"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.DeleteVendor(r.Context(), id)
	if err != nil {
		logger.Error("Failed to delete vendor", "id", id, "error", err)
		http.Error(w, `{"error": "failed to delete vendor"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": resp.Success})
}

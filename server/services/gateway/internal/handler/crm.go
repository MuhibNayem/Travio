package handler

import (
	"encoding/json"
	"net/http"

	crmv1 "github.com/MuhibNayem/Travio/server/api/proto/crm/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

type CRMHandler struct {
	client *client.CRMClient
}

func NewCRMHandler(c *client.CRMClient) *CRMHandler {
	return &CRMHandler{client: c}
}

// --- Coupons ---

func (h *CRMHandler) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req crmv1.CreateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateCoupon(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CRMHandler) ListCoupons(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	resp, err := h.client.ListCoupons(r.Context(), &crmv1.ListCouponsRequest{OrganizationId: orgID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CRMHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req crmv1.ValidateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ValidateCoupon(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Support ---

func (h *CRMHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req crmv1.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateTicket(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CRMHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	resp, err := h.client.ListTickets(r.Context(), &crmv1.ListTicketsRequest{OrganizationId: orgID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CRMHandler) AddTicketMessage(w http.ResponseWriter, r *http.Request) {
	var req crmv1.AddTicketMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ticket ID from URL param
	ticketID := chi.URLParam(r, "id")
	req.TicketId = ticketID

	resp, err := h.client.AddTicketMessage(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

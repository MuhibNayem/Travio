package handler

import (
	"encoding/json"
	"net/http"

	subscriptionv1 "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	client *client.SubscriptionClient
}

func NewSubscriptionHandler(client *client.SubscriptionClient) *SubscriptionHandler {
	return &SubscriptionHandler{client: client}
}

// CreatePlan - Admin
func (h *SubscriptionHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req subscriptionv1.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreatePlan(r.Context(), &req)
	if err != nil {
		// In a real app, map gRPC error to HTTP code
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	resp, err := h.client.ListPlans(r.Context(), includeInactive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Plan ID required", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetPlan(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	var req subscriptionv1.UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdatePlan(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req subscriptionv1.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify organization ownership from JWT claims
	orgID := middleware.GetOrgID(r.Context())
	if orgID != "" && orgID != req.OrganizationId {
		http.Error(w, "Unauthorized: organization mismatch", http.StatusForbidden)
		return
	}

	resp, err := h.client.CreateSubscription(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	if orgID == "" {
		http.Error(w, "Organization ID required", http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetSubscription(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	if orgID == "" {
		http.Error(w, "Organization ID required", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CancelSubscription(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	planID := r.URL.Query().Get("plan_id")
	status := r.URL.Query().Get("status")

	resp, err := h.client.ListSubscriptions(r.Context(), planID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SubscriptionHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "subID")
	if subID == "" {
		http.Error(w, "Subscription ID required", http.StatusBadRequest)
		return
	}

	resp, err := h.client.ListInvoices(r.Context(), subID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

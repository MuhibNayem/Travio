package handler

import (
	"encoding/json"
	"net/http"

	fleetv1 "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

type FleetHandler struct {
	client *client.FleetClient
}

func NewFleetHandler(c *client.FleetClient) *FleetHandler {
	return &FleetHandler{client: c}
}

// --- Assets ---

func (h *FleetHandler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	var req fleetv1.RegisterAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	asset, err := h.client.RegisterAsset(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(asset)
}

func (h *FleetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := h.client.GetAsset(r.Context(), &fleetv1.GetAssetRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(asset)
}

func (h *FleetHandler) UpdateAssetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req fleetv1.UpdateAssetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Id = id

	asset, err := h.client.UpdateAssetStatus(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(asset)
}

func (h *FleetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	// orgID should come from token (or query param if admin/operator handling multiple orgs)
	// For now, let's assume "organization_id" query param or extract from context if auth middleware sets it.
	// Since we are in Gateway, we usually rely on claims.
	// However, simple implementation: query param or header?
	// Let's use query param if provided, otherwise assume caller handles it or it's global?
	// ACTUALLY: Backend requires OrgID.
	// Let's take it from query param "organization_id".
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		// Try to fallback to user's org if available in context?
		// For now http 400 if missing.
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	req := &fleetv1.ListAssetsRequest{
		OrganizationId: orgID,
	}

	resp, err := h.client.ListAssets(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Location ---

func (h *FleetHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var req fleetv1.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdateLocation(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FleetHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	loc, err := h.client.GetLocation(r.Context(), &fleetv1.GetLocationRequest{AssetId: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loc)
}

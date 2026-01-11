package handler

import (
	"io"
	"net/http"

	fleetv1 "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

type FleetHandler struct {
	client *client.FleetClient
}

func NewFleetHandler(c *client.FleetClient) *FleetHandler {
	return &FleetHandler{client: c}
}

// --- Assets ---

func (h *FleetHandler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req fleetv1.RegisterAssetRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		logger.Error("Failed to unmarshal request", "error", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Inject Organization ID from context
	req.OrganizationId = middleware.GetOrgID(r.Context())
	if req.OrganizationId == "" {
		http.Error(w, "Unauthorized: missing organization context", http.StatusUnauthorized)
		return
	}

	asset, err := h.client.RegisterAsset(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Use protojson for response marshaling as well to handle Enums correctly
	respBytes, err := protojson.Marshal(asset)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

func (h *FleetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := h.client.GetAsset(r.Context(), &fleetv1.GetAssetRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respBytes, err := protojson.Marshal(asset)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

func (h *FleetHandler) UpdateAssetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req fleetv1.UpdateAssetStatusRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	asset, err := h.client.UpdateAssetStatus(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respBytes, err := protojson.Marshal(asset)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

func (h *FleetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	// 1. Try to get Org ID from Context (JWT)
	orgID := middleware.GetOrgID(r.Context())
	logger.Info("ListAssets Debug", "context_org_id", orgID, "query_org_id", r.URL.Query().Get("organization_id"))

	// 2. If not in context (e.g. admin overriding), check query param?
	// For now, enforce context for security unless empty.
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	if orgID == "" {
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
	respBytes, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

// --- Location ---

func (h *FleetHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req fleetv1.UpdateLocationRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdateLocation(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respBytes, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

func (h *FleetHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	loc, err := h.client.GetLocation(r.Context(), &fleetv1.GetLocationRequest{AssetId: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respBytes, err := protojson.Marshal(loc)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}

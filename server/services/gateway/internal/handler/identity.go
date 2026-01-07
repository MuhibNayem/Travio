package handler

import (
	"encoding/json"
	"net/http"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
)

// IdentityHandler handles auth/identity requests via gRPC
type IdentityHandler struct {
	client *client.IdentityClient
}

// NewIdentityHandler creates a new identity handler with gRPC client
func NewIdentityHandler(identityClient *client.IdentityClient) *IdentityHandler {
	return &IdentityHandler{client: identityClient}
}

// Register handles user registration
func (h *IdentityHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req identityv1.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.Register(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to register user", "error", err)
		http.Error(w, `{"error": "identity service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handles user authentication
func (h *IdentityHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req identityv1.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.Login(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to login", "error", err)
		http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RefreshToken handles token refresh
func (h *IdentityHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req identityv1.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.RefreshToken(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to refresh token", "error", err)
		http.Error(w, `{"error": "invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout
func (h *IdentityHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req identityv1.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	_, err := h.client.Logout(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to logout", "error", err)
		http.Error(w, `{"error": "logout failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateOrganization handles org creation
func (h *IdentityHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req identityv1.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateOrganization(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to create organization", "error", err)
		http.Error(w, `{"error": "identity service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

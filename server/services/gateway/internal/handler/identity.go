package handler

import (
	"encoding/json"
	"net/http"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	subscriptionv1 "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

// IdentityHandler handles auth/identity requests via gRPC
type IdentityHandler struct {
	client    *client.IdentityClient
	subClient *client.SubscriptionClient
}

// NewIdentityHandler creates a new identity handler with gRPC client
func NewIdentityHandler(identityClient *client.IdentityClient, subClient *client.SubscriptionClient) *IdentityHandler {
	return &IdentityHandler{
		client:    identityClient,
		subClient: subClient,
	}
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

	// Default to Free Tier if no plan selected
	if req.PlanId == "" {
		req.PlanId = "plan_free"
	}

	// 1. Create Organization in Identity Service
	resp, err := h.client.CreateOrganization(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to create organization", "error", err)
		http.Error(w, `{"error": "identity service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// 2. Create Subscription in Subscription Service
	// We use the PlanID from the request (which is now guaranteed to be set)
	logger.Info("Creating subscription for organization", "org_id", resp.OrganizationId, "plan_id", req.PlanId)
	subReq := &subscriptionv1.CreateSubscriptionRequest{
		OrganizationId: resp.OrganizationId,
		PlanId:         req.PlanId,
	}

	_, subErr := h.subClient.CreateSubscription(r.Context(), subReq)
	if subErr != nil {
		// Critical Error: Org created but subscription failed.
		// In a real system, we might want to rollback the Org or queue a retry.
		// For now, we log an error. The user will be blocked by EntitlementMiddleware until fixed.
		logger.Error("Failed to create default subscription for new org",
			"org_id", resp.OrganizationId,
			"plan_id", req.PlanId,
			"error", subErr)
	} else {
		logger.Info("Created default subscription for new org",
			"org_id", resp.OrganizationId,
			"plan_id", req.PlanId)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// AcceptInvite handles invite acceptance
func (h *IdentityHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req identityv1.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.AcceptInvite(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to accept invite", "error", err)
		http.Error(w, `{"error": "failed to accept invite"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateInvite handles sending invites
func (h *IdentityHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var req identityv1.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateInvite(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to create invite", "error", err)
		http.Error(w, `{"error": "failed to create invite"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ListInvites handles listing invites
func (h *IdentityHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")

	req := &identityv1.ListInvitesRequest{
		OrganizationId: orgID,
	}

	resp, err := h.client.ListInvites(r.Context(), req)
	if err != nil {
		logger.Error("Failed to list invites", "error", err)
		http.Error(w, `{"error": "failed to list invites"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListMembers handles listing members
func (h *IdentityHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")

	req := &identityv1.ListMembersRequest{
		OrganizationId: orgID,
		Page:           1,
		Limit:          10,
	}

	resp, err := h.client.ListMembers(r.Context(), req)
	if err != nil {
		logger.Error("Failed to list members", "error", err)
		http.Error(w, `{"error": "failed to list members"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateUserRole handles role updates
func (h *IdentityHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var req identityv1.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.UserId = userID

	resp, err := h.client.UpdateUserRole(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to update role", "error", err)
		http.Error(w, `{"error": "failed to update role"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RemoveMember handles removing a member
func (h *IdentityHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	orgID := r.URL.Query().Get("org_id")

	req := &identityv1.RemoveMemberRequest{
		UserId:         userID,
		OrganizationId: orgID,
	}

	resp, err := h.client.RemoveMember(r.Context(), req)
	if err != nil {
		logger.Error("Failed to remove member", "error", err)
		http.Error(w, `{"error": "failed to remove member"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

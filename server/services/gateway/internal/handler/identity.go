package handler

import (
	"encoding/json"
	"net/http"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	subscriptionv1 "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
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

	// Set Cookies
	setAuthCookies(w, resp.AccessToken, resp.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RefreshToken handles token refresh
func (h *IdentityHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req identityv1.RefreshTokenRequest
	// Try to decode body, but ignore error if empty (might use cookie)
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken == "" {
		cookie, err := r.Cookie("refresh_token")
		if err == nil {
			req.RefreshToken = cookie.Value
		}
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error": "missing refresh token"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.RefreshToken(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to refresh token", "error", err)
		http.Error(w, `{"error": "invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	// Update Cookies
	setAuthCookies(w, resp.AccessToken, resp.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout
func (h *IdentityHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req identityv1.LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken == "" {
		cookie, err := r.Cookie("refresh_token")
		if err == nil {
			req.RefreshToken = cookie.Value
		}
	}

	// Only call backend if we have a token
	if req.RefreshToken != "" {
		_, err := h.client.Logout(r.Context(), &req)
		if err != nil {
			logger.Error("Failed to logout", "error", err)
			// Proceed to clear cookies anyway
		}
	}

	// Clear Cookies
	clearAuthCookies(w)
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

// GetOrganization retrieves organization profile
func (h *IdentityHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetOrganization(r.Context(), &identityv1.GetOrganizationRequest{
		OrganizationId: orgID,
	})
	if err != nil {
		logger.Error("Failed to get organization", "error", err)
		http.Error(w, `{"error": "identity service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateOrganization updates organization profile
func (h *IdentityHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id required"}`, http.StatusBadRequest)
		return
	}

	var req identityv1.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.OrganizationId = orgID

	resp, err := h.client.UpdateOrganization(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to update organization", "error", err)
		http.Error(w, `{"error": "identity service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

// GetMe returns current user info from context (cookie auth)
func (h *IdentityHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if userID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":              userID,
		"organization_id": orgID,
		"role":            role,
	})
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
}

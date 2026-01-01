package handler

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle OrganizationID from request
	logger.Info("Register endpoint called")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"registered"}`))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.Info("Login endpoint called")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"token":"dummy-jwt"}`))
}

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	logger.Info("CreateOrganization endpoint called")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"organization_id":"org-123", "status":"active"}`))
}

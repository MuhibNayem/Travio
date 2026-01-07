package handler

import (
	"context"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	identityv1.UnimplementedIdentityServiceServer
	authService *service.AuthService
}

func NewGrpcHandler(authService *service.AuthService) *GrpcHandler {
	return &GrpcHandler{authService: authService}
}

func (h *GrpcHandler) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	user, err := h.authService.Register(req.Email, req.Password, req.OrganizationId)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &identityv1.RegisterResponse{
		UserId: user.ID,
	}, nil
}

func (h *GrpcHandler) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	// Note: In production, extract metadata from gRPC context for userAgent/IP
	tokenPair, err := h.authService.Login(req.Email, req.Password, "grpc-client", "0.0.0.0")
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Proto currently only has Token field; return access_token for now
	// TODO: Update proto to include refresh_token and expires_in
	return &identityv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (h *GrpcHandler) RefreshToken(ctx context.Context, req *identityv1.RefreshTokenRequest) (*identityv1.RefreshTokenResponse, error) {
	// Note: In production, extract metadata from gRPC context for userAgent/IP
	tokenPair, err := h.authService.RefreshTokens(req.RefreshToken, "grpc-client", "0.0.0.0")
	if err != nil {
		if err == service.ErrRefreshTokenReused {
			return nil, status.Error(codes.PermissionDenied, "session terminated due to security concern")
		}
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &identityv1.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (h *GrpcHandler) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	_ = h.authService.Logout(ctx, req.RefreshToken)
	return &identityv1.LogoutResponse{}, nil
}

func (h *GrpcHandler) CreateOrganization(ctx context.Context, req *identityv1.CreateOrgRequest) (*identityv1.CreateOrgResponse, error) {
	org, err := h.authService.CreateOrganization(req.Name, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create organization")
	}

	return &identityv1.CreateOrgResponse{
		OrganizationId: org.ID,
	}, nil
}

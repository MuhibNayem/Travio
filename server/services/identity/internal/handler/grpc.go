package handler

import (
	"context"
	"time"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
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

	var newOrg *domain.Organization
	if req.NewOrganization != nil {
		newOrg = &domain.Organization{
			Name:     req.NewOrganization.Name,
			Address:  req.NewOrganization.Address,
			Phone:    req.NewOrganization.Phone,
			Email:    req.NewOrganization.Email,
			Website:  req.NewOrganization.Website,
			Currency: req.NewOrganization.Currency,
			PlanID:   "plan_free", // Default plan
		}
	}

	user, err := h.authService.Register(req.Email, req.Password, req.OrganizationId, req.Name, newOrg)
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
	tokenPair, err := h.authService.Login(req.Email, req.Password, "grpc-client", "0.0.0.0")
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &identityv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (h *GrpcHandler) RefreshToken(ctx context.Context, req *identityv1.RefreshTokenRequest) (*identityv1.RefreshTokenResponse, error) {
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
	org, err := h.authService.CreateOrganization(req.Name, req.PlanId, req.Address, req.Phone, req.Email, req.Website, req.Currency)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create organization")
	}

	return &identityv1.CreateOrgResponse{
		OrganizationId: org.ID,
		Currency:       org.Currency,
	}, nil
}

func (h *GrpcHandler) GetOrganization(ctx context.Context, req *identityv1.GetOrganizationRequest) (*identityv1.Organization, error) {
	org, err := h.authService.GetOrganization(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "organization not found")
	}
	return mapOrganizationToProto(org), nil
}

func (h *GrpcHandler) UpdateOrganization(ctx context.Context, req *identityv1.UpdateOrganizationRequest) (*identityv1.Organization, error) {
	org := &domain.Organization{
		ID:       req.OrganizationId,
		Name:     req.Name,
		Address:  req.Address,
		Phone:    req.Phone,
		Email:    req.Email,
		Website:  req.Website,
		Status:   req.Status,
		Currency: req.Currency,
	}
	updated, err := h.authService.UpdateOrganization(org)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update organization")
	}
	return mapOrganizationToProto(updated), nil
}

// --- Member Management ---

func (h *GrpcHandler) ListMembers(ctx context.Context, req *identityv1.ListMembersRequest) (*identityv1.ListMembersResponse, error) {
	users, total, err := h.authService.ListMembers(req.OrganizationId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var members []*identityv1.Member
	for _, u := range users {
		members = append(members, &identityv1.Member{
			UserId:   u.ID,
			Email:    u.Email,
			Role:     u.Role,
			Status:   u.Status,
			JoinedAt: u.CreatedAt.String(),
		})
	}

	return &identityv1.ListMembersResponse{
		Members: members,
		Total:   int32(total),
	}, nil
}

func mapOrganizationToProto(org *domain.Organization) *identityv1.Organization {
	if org == nil {
		return nil
	}
	return &identityv1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		PlanId:    org.PlanID,
		Address:   org.Address,
		Phone:     org.Phone,
		Email:     org.Email,
		Website:   org.Website,
		Status:    org.Status,
		Currency:  org.Currency,
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
	}
}

func (h *GrpcHandler) UpdateUserRole(ctx context.Context, req *identityv1.UpdateUserRoleRequest) (*identityv1.UpdateUserRoleResponse, error) {
	err := h.authService.UpdateMemberRole(req.UserId, req.OrganizationId, req.NewRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &identityv1.UpdateUserRoleResponse{Success: true}, nil
}

func (h *GrpcHandler) RemoveMember(ctx context.Context, req *identityv1.RemoveMemberRequest) (*identityv1.RemoveMemberResponse, error) {
	err := h.authService.RemoveMember(req.UserId, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &identityv1.RemoveMemberResponse{Success: true}, nil
}

// --- Invite Management ---

func (h *GrpcHandler) CreateInvite(ctx context.Context, req *identityv1.CreateInviteRequest) (*identityv1.CreateInviteResponse, error) {
	invite, err := h.authService.CreateInvite(req.Email, req.Role, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.CreateInviteResponse{
		InviteId: invite.ID,
		Token:    invite.Token,
	}, nil
}

func (h *GrpcHandler) AcceptInvite(ctx context.Context, req *identityv1.AcceptInviteRequest) (*identityv1.AcceptInviteResponse, error) {
	tokenPair, user, err := h.authService.AcceptInvite(req.Token, req.Email, req.Password, req.Name)
	if err != nil {
		if err == service.ErrInviteNotFound {
			return nil, status.Error(codes.NotFound, "invite not found")
		}
		if err == service.ErrInviteExpired {
			return nil, status.Error(codes.FailedPrecondition, "invite expired")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.AcceptInviteResponse{
		UserId:         user.ID,
		OrganizationId: user.OrganizationID,
		AccessToken:    tokenPair.AccessToken,
		RefreshToken:   tokenPair.RefreshToken,
	}, nil
}

func (h *GrpcHandler) ListInvites(ctx context.Context, req *identityv1.ListInvitesRequest) (*identityv1.ListInvitesResponse, error) {
	invites, err := h.authService.ListInvites(req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoInvites []*identityv1.Invite
	for _, i := range invites {
		protoInvites = append(protoInvites, &identityv1.Invite{
			Id:        i.ID,
			Email:     i.Email,
			Role:      i.Role,
			Status:    i.Status,
			Token:     i.Token,
			CreatedAt: i.CreatedAt.String(),
		})
	}

	return &identityv1.ListInvitesResponse{Invites: protoInvites}, nil
}

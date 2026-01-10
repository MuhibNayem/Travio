package service

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/auth"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrRefreshTokenReused = errors.New("refresh token reuse detected, session terminated")
)

// TokenPair represents the response from Login/Refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token TTL in seconds
}

type AuthService struct {
	UserRepo         *repository.UserRepository
	OrgRepo          *repository.OrgRepository
	RefreshTokenRepo *repository.RefreshTokenRepository
	RedisRepo        *repository.RedisRepository
	InviteRepo       *repository.InviteRepository
	Notifier         NotificationService
}

func NewAuthService(userRepo *repository.UserRepository, orgRepo *repository.OrgRepository, rtRepo *repository.RefreshTokenRepository, redisRepo *repository.RedisRepository, inviteRepo *repository.InviteRepository, notifier NotificationService) *AuthService {
	return &AuthService{
		UserRepo:         userRepo,
		OrgRepo:          orgRepo,
		RefreshTokenRepo: rtRepo,
		RedisRepo:        redisRepo,
		InviteRepo:       inviteRepo,
		Notifier:         notifier,
	}
}

func (s *AuthService) Register(email, password, orgID, name string, newOrgDetails *domain.Organization) (*domain.User, error) {
	existing, _ := s.UserRepo.FindByEmail(email)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// 1. Create Organization if provided
	if newOrgDetails != nil {
		// Set default plan if missing? handled in CreateOrg?
		// We just create it.
		if err := s.OrgRepo.Create(newOrgDetails); err != nil {
			return nil, err // wrap error?
		}
		orgID = newOrgDetails.ID
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := "user"
	if newOrgDetails != nil {
		role = "admin" // Creator of org is admin
	}

	user := &domain.User{
		Email:          email,
		Name:           name,
		PasswordHash:   string(hashedPwd),
		OrganizationID: orgID,
		Role:           role,
	}

	if err := s.UserRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login authenticates and returns both Access and Refresh tokens
func (s *AuthService) Login(email, password, userAgent, ipAddress string) (*TokenPair, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate Access Token
	accessToken, err := auth.GenerateAccessToken(user.ID, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token (new family)
	refreshToken, jti, err := auth.GenerateRefreshToken(user.ID, "")
	if err != nil {
		return nil, err
	}

	// Store Refresh Token in DB for revocation
	rtRecord := &domain.RefreshToken{
		ID:         jti,
		UserID:     user.ID,
		FamilyID:   jti, // First token in family uses its own JTI as FamilyID
		TokenHash:  auth.HashToken(refreshToken),
		Revoked:    false,
		ExpiresAt:  time.Now().Add(auth.RefreshTokenTTL),
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
	if err := s.RefreshTokenRepo.Create(rtRecord); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenTTL.Seconds()),
	}, nil
}

// RefreshTokens performs token rotation: validates old refresh, issues new pair, revokes old
func (s *AuthService) RefreshTokens(refreshTokenString, userAgent, ipAddress string) (*TokenPair, error) {
	// Validate the incoming refresh token
	claims, err := auth.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	// Look up the token in DB
	storedToken, err := s.RefreshTokenRepo.FindByID(claims.ID)
	if err != nil {
		// Token not found = stolen or already rotated
		return nil, auth.ErrInvalidToken
	}

	// CRITICAL: Check if token was already revoked (Reuse Detection)
	if storedToken.Revoked {
		// Potential token theft! Revoke entire family
		_ = s.RefreshTokenRepo.RevokeFamily(storedToken.FamilyID)
		return nil, ErrRefreshTokenReused
	}

	// Revoke the old token (Single Use)
	if err := s.RefreshTokenRepo.Revoke(storedToken.ID); err != nil {
		return nil, err
	}

	// Get user info for new Access Token
	user, err := s.UserRepo.FindByID(storedToken.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new Access Token
	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.OrganizationID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate new Refresh Token (same family)
	newRefreshToken, newJTI, err := auth.GenerateRefreshToken(user.ID, storedToken.FamilyID)
	if err != nil {
		return nil, err
	}

	// Store new Refresh Token
	newRTRecord := &domain.RefreshToken{
		ID:         newJTI,
		UserID:     user.ID,
		FamilyID:   storedToken.FamilyID, // Same family
		TokenHash:  auth.HashToken(newRefreshToken),
		Revoked:    false,
		ExpiresAt:  time.Now().Add(auth.RefreshTokenTTL),
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
	if err := s.RefreshTokenRepo.Create(newRTRecord); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(auth.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout revokes a specific refresh token (single device logout)
func (s *AuthService) Logout(ctx context.Context, refreshTokenString string) error {
	claims, err := auth.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil // Silently succeed
	}

	// 1. DB Update (Source of Truth)
	if err := s.RefreshTokenRepo.Revoke(claims.ID); err != nil {
		return err
	}

	// 2. Cache Update (Immediate consistency)
	ttl := 7 * 24 * time.Hour
	return s.RedisRepo.BlacklistToken(ctx, claims.ID, ttl)
}

// LogoutAll revokes all refresh tokens for a user (all devices)
func (s *AuthService) LogoutAll(userID string) error {
	return s.RefreshTokenRepo.RevokeAllForUser(userID)
}

// GetActiveSessions returns all active sessions for the user
func (s *AuthService) GetActiveSessions(userID string) ([]*domain.RefreshToken, error) {
	return s.RefreshTokenRepo.GetActiveSessionsForUser(userID)
}

// RevokeSession revokes a specific session (for "Revoke Session" button in UI)
func (s *AuthService) RevokeSession(sessionID string) error {
	return s.RefreshTokenRepo.Revoke(sessionID)
}

func (s *AuthService) CreateOrganization(name, planID, address, phone, email, website string) (*domain.Organization, error) {
	org := &domain.Organization{
		Name:    name,
		PlanID:  planID,
		Address: address,
		Phone:   phone,
		Email:   email,
		Website: website,
	}
	if err := s.OrgRepo.Create(org); err != nil {
		return nil, err
	}
	return org, nil
}

// ValidateToken parses and validates the token, checking the blacklist cache
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*auth.AccessTokenClaims, error) {
	// 1. Parse locally (CPU cheap compared to DB)
	claims, err := auth.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 2. Check Blacklist Cache (Fast Redis lookup)
	if claims.ID != "" {
		isRevoked, err := s.RedisRepo.IsBlacklisted(ctx, claims.ID)
		if err != nil {
			return nil, err
		}
		if isRevoked {
			return nil, auth.ErrInvalidToken
		}
	}

	return claims, nil
}

// GetUser returns user profile with Read-Through Caching
func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	// 1. Check Cache
	cachedUser, err := s.RedisRepo.GetUser(ctx, userID)
	if err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 2. Cache Miss - Hit DB
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 3. Populate Cache
	_ = s.RedisRepo.CacheUser(ctx, user, 1*time.Hour)

	return user, nil
}

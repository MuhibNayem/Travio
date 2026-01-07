package service

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/auth"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInviteNotFound = errors.New("invite not found or expired")
	ErrInviteExpired  = errors.New("invite has expired")
)

// CreateInvite generates a secure token and stores the invite
func (s *AuthService) CreateInvite(email, role, orgID string) (*domain.Invite, error) {
	// 1. Check if user already exists in org
	existingUser, _ := s.UserRepo.FindByEmail(email)
	if existingUser != nil && existingUser.OrganizationID == orgID {
		return nil, errors.New("user is already a member of this organization")
	}

	// 2. Check for pending invite
	pending, _ := s.InviteRepo.FindByEmailAndOrg(email, orgID)
	if pending != nil {
		// Return existing (or regenerate?) - For now return existing
		return pending, nil
	}

	// 3. Generate Token
	token, _ := auth.GenerateOpaqueToken(32) // Crypto secure random string

	invite := &domain.Invite{
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		Token:          token,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
	}

	if err := s.InviteRepo.Create(invite); err != nil {
		return nil, err
	}

	orgName := "Travio Organization"
	if org, err := s.OrgRepo.FindByID(orgID); err == nil {
		orgName = org.Name
	}

	_ = s.Notifier.SendInviteEmail(context.Background(), email, token, orgName)

	return invite, nil
}

// AcceptInvite validates token and creates user
func (s *AuthService) AcceptInvite(token, email, password, name string) (*TokenPair, *domain.User, error) {
	invite, err := s.InviteRepo.FindByToken(token)
	if err != nil || invite == nil {
		return nil, nil, ErrInviteNotFound
	}

	if invite.ExpiresAt.Before(time.Now()) {
		return nil, nil, ErrInviteExpired
	}

	if invite.Status != "pending" {
		return nil, nil, errors.New("invite already used")
	}

	if invite.Email != email {
		return nil, nil, errors.New("email mismatch")
	}

	// Check if user exists (maybe invite was sent to existing user to join org?)
	existingUser, _ := s.UserRepo.FindByEmail(email)
	if existingUser != nil {
		// Link existing user to organization
		existingUser.OrganizationID = invite.OrganizationID
		existingUser.Role = invite.Role
		if err := s.UserRepo.Update(existingUser); err != nil {
			return nil, nil, err
		}

		// Mark invite as accepted
		_ = s.InviteRepo.UpdateStatus(invite.ID, "accepted")

		// Log them in
		tokens, err := s.Login(email, password, "InviteAccept", "0.0.0.0")
		return tokens, existingUser, err
	}

	// Create User
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		Email:          email,
		PasswordHash:   string(hashedPwd),
		OrganizationID: invite.OrganizationID,
		Role:           invite.Role,
	}

	if err := s.UserRepo.Create(user); err != nil {
		return nil, nil, err
	}

	// Mark invite as accepted
	if err := s.InviteRepo.UpdateStatus(invite.ID, "accepted"); err != nil {
		// Log error but proceed
	}

	// Generate Tokens
	tokens, err := s.Login(email, password, "InviteAccept", "0.0.0.0")
	return tokens, user, err
}

func (s *AuthService) ListInvites(orgID string) ([]*domain.Invite, error) {
	return s.InviteRepo.ListByOrg(orgID)
}

func (s *AuthService) ListMembers(orgID string, page, limit int) ([]*domain.User, int, error) {
	offset := (page - 1) * limit
	return s.InviteRepo.ListMembers(orgID, limit, offset)
}

func (s *AuthService) UpdateMemberRole(userID, orgID, newRole string) error {
	// Prevent Lockout: If changing an admin to non-admin, ensure they are not the last admin
	if newRole != "admin" {
		user, err := s.UserRepo.FindByID(userID)
		if err == nil && user.Role == "admin" {
			adminCount, err := s.InviteRepo.CountAdmins(orgID)
			if err == nil && adminCount <= 1 {
				return errors.New("cannot remove the last admin from the organization")
			}
		}
	}
	return s.InviteRepo.UpdateMemberRole(userID, orgID, newRole)
}

func (s *AuthService) RemoveMember(userID, orgID string) error {
	// Prevent Lockout: If removing an admin, ensure they are not the last admin
	user, err := s.UserRepo.FindByID(userID)
	if err == nil && user.Role == "admin" {
		adminCount, err := s.InviteRepo.CountAdmins(orgID)
		if err == nil && adminCount <= 1 {
			return errors.New("cannot remove the last admin from the organization")
		}
	}
	return s.InviteRepo.RemoveMember(userID, orgID)
}

package domain

import (
	"time"
)

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PlanID    string    `json:"plan_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"` // "admin", "agent", "user"
	CreatedAt      time.Time `json:"created_at"`
}

// RefreshToken represents a stored refresh token for revocation and rotation tracking
type RefreshToken struct {
	ID         string    `json:"id"` // JTI of the token
	UserID     string    `json:"user_id"`
	FamilyID   string    `json:"family_id"` // Token family for rotation
	TokenHash  string    `json:"-"`         // SHA256 hash of the actual token (for secure lookup)
	Revoked    bool      `json:"revoked"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	UserAgent  string    `json:"user_agent"` // For session tracking
	IPAddress  string    `json:"ip_address"`
}

// Session represents an active user session (for "View Active Sessions" feature)
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FamilyID   string    `json:"family_id"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
}

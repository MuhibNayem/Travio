package domain

import "time"

// Invite represents an invitation for a user to join an organization
type Invite struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Email          string    `json:"email" db:"email"`
	Role           string    `json:"role" db:"role"`
	Token          string    `json:"token" db:"token"`
	Status         string    `json:"status" db:"status"` // pending, accepted, expired
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

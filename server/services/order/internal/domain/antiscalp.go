package domain

import (
	"errors"
	"time"
)

var (
	ErrMaxTicketsExceeded = errors.New("maximum tickets per user exceeded")
	ErrDuplicateNID       = errors.New("NID already registered for this trip")
	ErrInvalidNID         = errors.New("invalid NID format")
)

// Passenger represents a ticket holder with identity verification
type Passenger struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`

	// Identity Verification
	NID            string    `json:"nid"`             // National ID Number
	PassportNumber string    `json:"passport_number"` // Alternative for foreigners
	FullName       string    `json:"full_name"`
	DateOfBirth    time.Time `json:"date_of_birth"`
	Gender         string    `json:"gender"`
	Phone          string    `json:"phone"`

	// Verification Status
	VerificationStatus string    `json:"verification_status"` // "pending", "verified", "failed"
	VerifiedAt         time.Time `json:"verified_at,omitempty"`

	// Seat Assignment
	SeatNumber string `json:"seat_number"`
	SeatClass  string `json:"seat_class"`

	CreatedAt time.Time `json:"created_at"`
}

// TicketLimit defines per-trip limits for anti-scalping
type TicketLimit struct {
	// Per User limits
	MaxTicketsPerUser int `json:"max_tickets_per_user"` // e.g., 6 tickets per user per trip
	MaxTicketsPerIP   int `json:"max_tickets_per_ip"`   // e.g., 10 tickets per IP per trip
	MaxTicketsPerNID  int `json:"max_tickets_per_nid"`  // e.g., 1 ticket per NID per trip (prevent resale)

	// Time-based limits
	MaxTicketsPerHour int `json:"max_tickets_per_hour"` // e.g., 20 across all trips
	MaxHoldsPerUser   int `json:"max_holds_per_user"`   // e.g., 2 concurrent holds
}

// DefaultTicketLimits returns production-safe default limits
func DefaultTicketLimits() TicketLimit {
	return TicketLimit{
		MaxTicketsPerUser: 6,
		MaxTicketsPerIP:   10,
		MaxTicketsPerNID:  1, // CRITICAL: One ticket per NID
		MaxTicketsPerHour: 20,
		MaxHoldsPerUser:   2,
	}
}

// Hold represents a temporary seat reservation (expires after TTL)
type Hold struct {
	ID        string `json:"id"`
	TripID    string `json:"trip_id"`
	SegmentID string `json:"segment_id"`
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"` // Browser session for anonymous holds
	IPAddress string `json:"ip_address"`

	Status    string    `json:"status"` // "active", "expired", "converted"
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// HoldTTL is the default hold expiration time
const HoldTTL = 10 * time.Minute

// NIDValidator validates Bangladesh National ID format
type NIDValidator struct{}

// Validate checks if NID is in valid format
func (v *NIDValidator) Validate(nid string) error {
	// Bangladesh NID: 10 or 17 digits
	if len(nid) != 10 && len(nid) != 17 {
		return ErrInvalidNID
	}

	// Must be all digits
	for _, c := range nid {
		if c < '0' || c > '9' {
			return ErrInvalidNID
		}
	}

	// TODO: Add checksum validation for 17-digit smart card NIDs

	return nil
}

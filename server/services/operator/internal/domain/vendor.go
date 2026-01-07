package domain

import "time"

// Vendor represents a bus/train operator in the system
type Vendor struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	ContactEmail   string    `json:"contact_email" db:"contact_email"`
	ContactPhone   string    `json:"contact_phone" db:"contact_phone"`
	Address        string    `json:"address" db:"address"`
	Status         string    `json:"status" db:"status"` // active, inactive, suspended
	CommissionRate float64   `json:"commission_rate" db:"commission_rate"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

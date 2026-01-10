package domain

import (
	"time"
)

type Asset struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	LicensePlate   string    `json:"license_plate"`
	VIN            string    `json:"vin"`
	Make           string    `json:"make"`
	Model          string    `json:"model"`
	Year           int32     `json:"year"`
	Type           string    `json:"type"`   // BUS, TRAIN, LAUNCH
	Status         string    `json:"status"` // ACTIVE, MAINTENANCE
	Config         string    `json:"config"` // JSONB for layouts
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AssetLocation struct {
	AssetID        string    `json:"asset_id"`
	OrganizationID string    `json:"organization_id"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Speed          float64   `json:"speed"`
	Heading        float64   `json:"heading"` // Degrees 0-360
	Timestamp      time.Time `json:"timestamp"`
}

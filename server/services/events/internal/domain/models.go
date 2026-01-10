package domain

import (
	"time"
)

type Venue struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	City           string    `json:"city"`
	Country        string    `json:"country"`
	Capacity       int32     `json:"capacity"`
	Type           string    `json:"type"`     // Enum in proto, string in DB
	Sections       string    `json:"sections"` // JSONB
	MapImageURL    string    `json:"map_image_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Event struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	VenueID        string    `json:"venue_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	Images         []string  `json:"images"` // PostgreSQL Array
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TicketType struct {
	ID                string    `json:"id"`
	EventID           string    `json:"event_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	PricePaisa        int64     `json:"price_paisa"`
	TotalQuantity     int32     `json:"total_quantity"`
	AvailableQuantity int32     `json:"available_quantity"`
	MaxPerUser        int32     `json:"max_per_user"`
	SalesStartTime    time.Time `json:"sales_start_time"`
	SalesEndTime      time.Time `json:"sales_end_time"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

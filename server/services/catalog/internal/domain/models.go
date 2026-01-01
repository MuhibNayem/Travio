package domain

import (
	"time"
)

// Station represents a transportation terminal (bus stand, railway station, ferry ghat)
type Station struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"` // Unique short code like "DHA"
	Name           string    `json:"name"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Country        string    `json:"country"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Timezone       string    `json:"timezone"`
	Address        string    `json:"address"`
	Amenities      []string  `json:"amenities"`
	Status         string    `json:"status"` // active, inactive, under_construction
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Route represents a path between stations with possible intermediate stops
type Route struct {
	ID                   string      `json:"id"`
	OrganizationID       string      `json:"organization_id"`
	Code                 string      `json:"code"` // e.g., "DHA-CTG-001"
	Name                 string      `json:"name"`
	OriginStationID      string      `json:"origin_station_id"`
	DestinationStationID string      `json:"destination_station_id"`
	IntermediateStops    []RouteStop `json:"intermediate_stops"`
	DistanceKm           int         `json:"distance_km"`
	EstimatedDurationMin int         `json:"estimated_duration_minutes"`
	Status               string      `json:"status"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// RouteStop represents an intermediate stop on a route
type RouteStop struct {
	StationID              string `json:"station_id"`
	Sequence               int    `json:"sequence"`
	ArrivalOffsetMinutes   int    `json:"arrival_offset_minutes"`
	DepartureOffsetMinutes int    `json:"departure_offset_minutes"`
	DistanceFromOriginKm   int    `json:"distance_from_origin_km"`
}

// Trip represents a specific scheduled journey on a route
type Trip struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organization_id"`
	RouteID        string        `json:"route_id"`
	VehicleID      string        `json:"vehicle_id"`    // Bus/Train number
	VehicleType    string        `json:"vehicle_type"`  // bus, train, ferry
	VehicleClass   string        `json:"vehicle_class"` // economy, business, ac
	DepartureTime  time.Time     `json:"departure_time"`
	ArrivalTime    time.Time     `json:"arrival_time"`
	TotalSeats     int           `json:"total_seats"`
	AvailableSeats int           `json:"available_seats"`
	Pricing        TripPricing   `json:"pricing"`
	Status         string        `json:"status"`
	Segments       []TripSegment `json:"segments"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// TripPricing contains pricing information for a trip
type TripPricing struct {
	BasePricePaisa  int64            `json:"base_price_paisa"`
	TaxPaisa        int64            `json:"tax_paisa"`
	BookingFeePaisa int64            `json:"booking_fee_paisa"`
	Currency        string           `json:"currency"`
	ClassPrices     map[string]int64 `json:"class_prices"`
}

// TripSegment represents a segment of a trip for intermediate boarding
type TripSegment struct {
	SegmentIndex   int       `json:"segment_index"`
	FromStationID  string    `json:"from_station_id"`
	ToStationID    string    `json:"to_station_id"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	AvailableSeats int       `json:"available_seats"`
}

// Vehicle represents a transport vehicle
type Vehicle struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	RegistrationNo string     `json:"registration_no"`
	Type           string     `json:"type"`  // bus, train, ferry
	Class          string     `json:"class"` // ac, non-ac, sleeper
	Capacity       int        `json:"capacity"`
	SeatLayout     SeatLayout `json:"seat_layout"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

// SeatLayout defines the seating arrangement
type SeatLayout struct {
	Rows       int      `json:"rows"`
	Columns    int      `json:"columns"`
	SeatMap    [][]Seat `json:"seat_map"`
	TotalSeats int      `json:"total_seats"`
}

// Seat represents a single seat
type Seat struct {
	ID       string `json:"id"`
	Row      int    `json:"row"`
	Column   int    `json:"column"`
	Number   string `json:"number"`    // Display number like "A1"
	Type     string `json:"type"`      // window, aisle, middle
	Class    string `json:"class"`     // economy, business
	IsActive bool   `json:"is_active"` // Can be disabled for maintenance
}

// Status constants
const (
	StationStatusActive            = "active"
	StationStatusInactive          = "inactive"
	StationStatusUnderConstruction = "under_construction"

	RouteStatusActive    = "active"
	RouteStatusInactive  = "inactive"
	RouteStatusSuspended = "suspended"

	TripStatusScheduled = "scheduled"
	TripStatusBoarding  = "boarding"
	TripStatusDeparted  = "departed"
	TripStatusInTransit = "in_transit"
	TripStatusArrived   = "arrived"
	TripStatusCancelled = "cancelled"
	TripStatusDelayed   = "delayed"
)

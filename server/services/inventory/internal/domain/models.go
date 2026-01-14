package domain

import (
	"time"
)

// Segment represents one leg of a trip between two consecutive stops
// This is the core unit for segment-based inventory (IRCTC style)
// For a trip A -> B -> C -> D, segments are: A-B, B-C, C-D
type Segment struct {
	OrganizationID string    `json:"organization_id"`
	TripID        string    `json:"trip_id"`
	SegmentIndex  int       `json:"segment_index"`
	FromStationID string    `json:"from_station_id"`
	ToStationID   string    `json:"to_station_id"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
}

// SeatInventory represents the availability of a specific seat on a specific segment
// Stored in ScyllaDB with partition key = (trip_id, segment_index)
type SeatInventory struct {
	OrganizationID string    `json:"organization_id"`
	TripID       string    `json:"trip_id"`
	SegmentIndex int       `json:"segment_index"`
	SeatID       string    `json:"seat_id"`
	SeatNumber   string    `json:"seat_number"`
	SeatClass    string    `json:"seat_class"` // economy, business, ac
	SeatType     string    `json:"seat_type"`  // window, aisle, middle
	Status       string    `json:"status"`     // available, held, booked, blocked
	HoldID       string    `json:"hold_id,omitempty"`
	HoldUserID   string    `json:"hold_user_id,omitempty"`
	HoldExpiry   time.Time `json:"hold_expiry,omitempty"`
	BookingID    string    `json:"booking_id,omitempty"`
	PricePaisa   int64     `json:"price_paisa"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SeatHold represents a temporary reservation across multiple segments
type SeatHold struct {
	HoldID        string    `json:"hold_id"`
	OrganizationID string    `json:"organization_id"`
	TripID        string    `json:"trip_id"`
	UserID        string    `json:"user_id"`
	SessionID     string    `json:"session_id"`
	FromStationID string    `json:"from_station_id"`
	ToStationID   string    `json:"to_station_id"`
	SeatIDs       []string  `json:"seat_ids"`
	SegmentRange  []int     `json:"segment_range"` // e.g., [0,1,2] for A-D via B,C
	Status        string    `json:"status"`        // active, expired, converted, released
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	IPAddress     string    `json:"ip_address"`
}

// Booking represents confirmed seat reservations
type Booking struct {
	BookingID     string       `json:"booking_id"`
	OrderID       string       `json:"order_id"`
	OrganizationID string       `json:"organization_id"`
	TripID        string       `json:"trip_id"`
	UserID        string       `json:"user_id"`
	FromStationID string       `json:"from_station_id"`
	ToStationID   string       `json:"to_station_id"`
	Seats         []BookedSeat `json:"seats"`
	SegmentRange  []int        `json:"segment_range"`
	TotalPaisa    int64        `json:"total_paisa"`
	Status        string       `json:"status"` // confirmed, cancelled, completed
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// BookedSeat represents a single booked seat with passenger info
type BookedSeat struct {
	SeatID        string `json:"seat_id"`
	SeatNumber    string `json:"seat_number"`
	SeatClass     string `json:"seat_class"`
	TicketID      string `json:"ticket_id"`
	PassengerNID  string `json:"passenger_nid"`
	PassengerName string `json:"passenger_name"`
	PricePaisa    int64  `json:"price_paisa"`
}

// Status constants
const (
	SeatStatusAvailable = "available"
	SeatStatusHeld      = "held"
	SeatStatusBooked    = "booked"
	SeatStatusBlocked   = "blocked"

	HoldStatusActive    = "active"
	HoldStatusExpired   = "expired"
	HoldStatusConverted = "converted"
	HoldStatusReleased  = "released"

	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
)

// SegmentRange calculates which segment indices are covered for a journey
// For trip with stops [A, B, C, D] (indices 0-3):
// - Journey A->D covers segments [0, 1, 2]
// - Journey B->D covers segments [1, 2]
// - Journey A->B covers segments [0]
func CalculateSegmentRange(stops []string, fromStation, toStation string) ([]int, error) {
	fromIdx := -1
	toIdx := -1

	for i, stop := range stops {
		if stop == fromStation {
			fromIdx = i
		}
		if stop == toStation {
			toIdx = i
		}
	}

	if fromIdx == -1 || toIdx == -1 || fromIdx >= toIdx {
		return nil, ErrInvalidStationRange
	}

	var segments []int
	for i := fromIdx; i < toIdx; i++ {
		segments = append(segments, i)
	}
	return segments, nil
}

var ErrInvalidStationRange = &DomainError{Message: "invalid station range"}
var ErrSeatNotAvailable = &DomainError{Message: "seat not available"}
var ErrHoldExpired = &DomainError{Message: "hold has expired"}
var ErrHoldNotFound = &DomainError{Message: "hold not found"}

type DomainError struct {
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

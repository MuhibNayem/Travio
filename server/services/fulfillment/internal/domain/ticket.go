package domain

import "time"

type Ticket struct {
	ID             string `json:"id"`
	BookingID      string `json:"booking_id"`
	OrderID        string `json:"order_id"`
	OrganizationID string `json:"organization_id"`

	// Trip Info
	TripID        string    `json:"trip_id"`
	RouteName     string    `json:"route_name"`
	FromStation   string    `json:"from_station"`
	ToStation     string    `json:"to_station"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`

	// Passenger
	PassengerNID  string `json:"passenger_nid"`
	PassengerName string `json:"passenger_name"`
	SeatNumber    string `json:"seat_number"`
	SeatClass     string `json:"seat_class"`

	// Pricing
	PricePaisa int64  `json:"price_paisa"`
	Currency   string `json:"currency"`

	// QR Code
	QRCodeData string `json:"qr_code_data"`
	QRCodeURL  string `json:"qr_code_url"`

	// PDF
	PDFURL string `json:"pdf_url"`

	// Status
	Status     TicketStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	ValidUntil time.Time    `json:"valid_until"`

	// Boarding
	IsBoarded bool      `json:"is_boarded"`
	BoardedAt time.Time `json:"boarded_at,omitempty"`
	BoardedBy string    `json:"boarded_by,omitempty"`
}

type TicketStatus string

const (
	TicketStatusActive    TicketStatus = "active"
	TicketStatusUsed      TicketStatus = "used"
	TicketStatusCancelled TicketStatus = "cancelled"
	TicketStatusExpired   TicketStatus = "expired"
)

// QRPayload is the data encoded in the ticket QR code
type QRPayload struct {
	Version      int    `json:"v"`
	TicketID     string `json:"tid"`
	BookingID    string `json:"bid"`
	PassengerNID string `json:"nid"`
	SeatNumber   string `json:"seat"`
	TripID       string `json:"trip"`
	Departure    int64  `json:"dep"`
	Signature    string `json:"sig"`
}

package events

import (
	"context"
	"database/sql"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/outbox"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
)

// Publisher handles event publishing for catalog domain events
type Publisher struct {
	outbox *outbox.Publisher
}

// NewPublisher creates a new event publisher
func NewPublisher(db *sql.DB) *Publisher {
	return &Publisher{
		outbox: outbox.NewPublisher(db),
	}
}

// TripCreatedPayload is the event payload for trip created
type TripCreatedPayload struct {
	TripID          string `json:"trip_id"`
	OrganizationID  string `json:"organization_id"`
	RouteID         string `json:"route_id"`
	VehicleID       string `json:"vehicle_id"`
	VehicleType     string `json:"vehicle_type"`
	Date            string `json:"date"`
	FromStationID   string `json:"from_station_id"`
	ToStationID     string `json:"to_station_id"`
	FromStationName string `json:"from_station_name"`
	ToStationName   string `json:"to_station_name"`
	DepartureTime   int64  `json:"departure_time"`
	TotalSeats      int    `json:"total_seats"`
	Status          string `json:"status"`
}

// PublishTripCreated publishes trip created event within a transaction
func (p *Publisher) PublishTripCreated(ctx context.Context, tx *sql.Tx, trip *domain.Trip) error {
	payload := TripCreatedPayload{
		TripID:          trip.ID,
		OrganizationID:  trip.OrganizationID,
		RouteID:         trip.RouteID,
		VehicleID:       trip.VehicleID,
		VehicleType:     trip.VehicleType,
		Date:            trip.ServiceDate,
		FromStationID:   trip.OriginStationID,
		ToStationID:     trip.DestinationStationID,
		FromStationName: trip.OriginStationName,
		ToStationName:   trip.DestinationStationName,
		DepartureTime:   trip.DepartureTime.Unix(),
		TotalSeats:      trip.TotalSeats,
		Status:          trip.Status,
	}
	// Note: We need to ensure TopicTrips and EventTripCreated exist in pkg/kafka
	// Using hardcoded strings or we'll check pkg/kafka next.
	// OrderService uses kafka.TopicOrders.
	// I'll assume kafka.TopicTrips/Catalog exists or I will add it.
	return p.outbox.Publish(ctx, tx, kafka.TopicCatalog, kafka.EventTripCreated, trip.ID, payload)
}

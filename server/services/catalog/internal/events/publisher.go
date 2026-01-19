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
	TripID          string               `json:"trip_id"`
	OrganizationID  string               `json:"organization_id"`
	RouteID         string               `json:"route_id"`
	VehicleID       string               `json:"vehicle_id"`
	VehicleType     string               `json:"vehicle_type"`
	VehicleClass    string               `json:"vehicle_class"`
	Date            string               `json:"date"`
	ServiceDate     string               `json:"service_date"`
	FromStationID   string               `json:"from_station_id"`
	ToStationID     string               `json:"to_station_id"`
	FromStationName string               `json:"from_station_name"`
	ToStationName   string               `json:"to_station_name"`
	FromCity        string               `json:"from_city"`
	ToCity          string               `json:"to_city"`
	DepartureTime   int64                `json:"departure_time"`
	ArrivalTime     int64                `json:"arrival_time"`
	TotalSeats      int                  `json:"total_seats"`
	AvailableSeats  int                  `json:"available_seats"`
	PricePaisa      int64                `json:"price_paisa"`
	Pricing         domain.TripPricing   `json:"pricing"`
	Segments        []domain.TripSegment `json:"segments"`
	Status          string               `json:"status"`
}

// PublishTripCreated publishes trip created event within a transaction
func (p *Publisher) PublishTripCreated(ctx context.Context, tx *sql.Tx, trip *domain.Trip) error {
	payload := TripCreatedPayload{
		TripID:          trip.ID,
		OrganizationID:  trip.OrganizationID,
		RouteID:         trip.RouteID,
		VehicleID:       trip.VehicleID,
		VehicleType:     trip.VehicleType,
		VehicleClass:    trip.VehicleClass,
		Date:            trip.ServiceDate,
		ServiceDate:     trip.ServiceDate,
		FromStationID:   trip.OriginStationID,
		ToStationID:     trip.DestinationStationID,
		FromStationName: trip.OriginStationName,
		ToStationName:   trip.DestinationStationName,
		FromCity:        trip.OriginStationCity,
		ToCity:          trip.DestinationStationCity,
		DepartureTime:   trip.DepartureTime.Unix(),
		ArrivalTime:     trip.ArrivalTime.Unix(),
		TotalSeats:      trip.TotalSeats,
		AvailableSeats:  trip.AvailableSeats,
		PricePaisa:      trip.Pricing.BasePricePaisa,
		Status:          trip.Status,
		Pricing:         trip.Pricing,
		Segments:        trip.Segments,
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicCatalog, kafka.EventTripCreated, trip.ID, payload)
}

// TripUpdatedPayload matches TripCreatedPayload for full state updates
type TripUpdatedPayload TripCreatedPayload

// PublishTripUpdated publishes trip updated event
func (p *Publisher) PublishTripUpdated(ctx context.Context, tx *sql.Tx, trip *domain.Trip) error {
	payload := TripUpdatedPayload{
		TripID:          trip.ID,
		OrganizationID:  trip.OrganizationID,
		RouteID:         trip.RouteID,
		VehicleID:       trip.VehicleID,
		VehicleType:     trip.VehicleType,
		VehicleClass:    trip.VehicleClass,
		Date:            trip.ServiceDate,
		ServiceDate:     trip.ServiceDate,
		FromStationID:   trip.OriginStationID,
		ToStationID:     trip.DestinationStationID,
		FromStationName: trip.OriginStationName,
		ToStationName:   trip.DestinationStationName,
		FromCity:        trip.OriginStationCity,
		ToCity:          trip.DestinationStationCity,
		DepartureTime:   trip.DepartureTime.Unix(),
		ArrivalTime:     trip.ArrivalTime.Unix(),
		TotalSeats:      trip.TotalSeats,
		AvailableSeats:  trip.AvailableSeats,
		PricePaisa:      trip.Pricing.BasePricePaisa,
		Status:          trip.Status,
		Pricing:         trip.Pricing,
		Segments:        trip.Segments,
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicCatalog, kafka.EventTripUpdated, trip.ID, payload)
}

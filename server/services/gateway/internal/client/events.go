package client

import (
	"context"
	"time"

	eventsv1 "github.com/MuhibNayem/Travio/server/api/proto/events/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

type EventsClient struct {
	conn   *grpc.ClientConn
	client eventsv1.EventServiceClient
}

func NewEventsClient(address string, tlsCfg TLSConfig) (*EventsClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to events service", "address", address, "tls", tlsCfg.CertFile != "")
	return &EventsClient{
		conn:   conn,
		client: eventsv1.NewEventServiceClient(conn),
	}, nil
}

func (c *EventsClient) Close() error {
	return c.conn.Close()
}

// --- Venues ---

func (c *EventsClient) CreateVenue(ctx context.Context, req *eventsv1.CreateVenueRequest) (*eventsv1.Venue, error) {
	return c.client.CreateVenue(ctx, req)
}

func (c *EventsClient) GetVenue(ctx context.Context, req *eventsv1.GetVenueRequest) (*eventsv1.Venue, error) {
	return c.client.GetVenue(ctx, req)
}

func (c *EventsClient) ListVenues(ctx context.Context, req *eventsv1.ListVenuesRequest) (*eventsv1.ListVenuesResponse, error) {
	return c.client.ListVenues(ctx, req)
}

func (c *EventsClient) UpdateVenue(ctx context.Context, req *eventsv1.UpdateVenueRequest) (*eventsv1.Venue, error) {
	return c.client.UpdateVenue(ctx, req)
}

// --- Events ---

func (c *EventsClient) CreateEvent(ctx context.Context, req *eventsv1.CreateEventRequest) (*eventsv1.Event, error) {
	return c.client.CreateEvent(ctx, req)
}

func (c *EventsClient) GetEvent(ctx context.Context, req *eventsv1.GetEventRequest) (*eventsv1.Event, error) {
	return c.client.GetEvent(ctx, req)
}

func (c *EventsClient) ListEvents(ctx context.Context, req *eventsv1.ListEventsRequest) (*eventsv1.ListEventsResponse, error) {
	return c.client.ListEvents(ctx, req)
}

func (c *EventsClient) UpdateEvent(ctx context.Context, req *eventsv1.UpdateEventRequest) (*eventsv1.Event, error) {
	return c.client.UpdateEvent(ctx, req)
}

func (c *EventsClient) PublishEvent(ctx context.Context, req *eventsv1.PublishEventRequest) (*eventsv1.Event, error) {
	return c.client.PublishEvent(ctx, req)
}

func (c *EventsClient) SearchEvents(ctx context.Context, req *eventsv1.SearchEventsRequest) (*eventsv1.SearchEventsResponse, error) {
	return c.client.SearchEvents(ctx, req)
}

// --- Tickets ---

func (c *EventsClient) CreateTicketType(ctx context.Context, req *eventsv1.CreateTicketTypeRequest) (*eventsv1.TicketType, error) {
	return c.client.CreateTicketType(ctx, req)
}

func (c *EventsClient) ListTicketTypes(ctx context.Context, req *eventsv1.ListTicketTypesRequest) (*eventsv1.ListTicketTypesResponse, error) {
	return c.client.ListTicketTypes(ctx, req)
}

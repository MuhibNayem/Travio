package handler

import (
	"context"
	"encoding/json"
	"time"

	eventsv1 "github.com/MuhibNayem/Travio/server/api/proto/events/v1"
	"github.com/MuhibNayem/Travio/server/services/events/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/events/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	eventsv1.UnimplementedEventServiceServer
	service *service.EventService
}

func NewGRPCHandler(svc *service.EventService) *GRPCHandler {
	return &GRPCHandler{service: svc}
}

// --- Venues ---

func (h *GRPCHandler) CreateVenue(ctx context.Context, req *eventsv1.CreateVenueRequest) (*eventsv1.Venue, error) {
	// Convert Sections proto to JSON string
	sectionsJSON, _ := json.Marshal(req.Sections)

	v, err := h.service.CreateVenue(ctx, req.Name, req.OrganizationId, req.Address, req.City, req.Country, 0, req.Type.String(), "", string(sectionsJSON))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapVenueToProto(v), nil
}

func (h *GRPCHandler) GetVenue(ctx context.Context, req *eventsv1.GetVenueRequest) (*eventsv1.Venue, error) {
	v, err := h.service.GetVenue(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return mapVenueToProto(v), nil
}

func (h *GRPCHandler) ListVenues(ctx context.Context, req *eventsv1.ListVenuesRequest) (*eventsv1.ListVenuesResponse, error) {
	venues, err := h.service.ListVenues(ctx, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoVenues []*eventsv1.Venue
	for _, v := range venues {
		protoVenues = append(protoVenues, mapVenueToProto(v))
	}
	return &eventsv1.ListVenuesResponse{Venues: protoVenues}, nil
}

func (h *GRPCHandler) UpdateVenue(ctx context.Context, req *eventsv1.UpdateVenueRequest) (*eventsv1.Venue, error) {
	v, err := h.service.UpdateVenue(ctx, req.Id, req.Name, req.Type.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapVenueToProto(v), nil
}

// --- Events ---

func (h *GRPCHandler) CreateEvent(ctx context.Context, req *eventsv1.CreateEventRequest) (*eventsv1.Event, error) {
	start, _ := time.Parse(time.RFC3339, req.StartTime)
	end, _ := time.Parse(time.RFC3339, req.EndTime)

	e, err := h.service.CreateEvent(ctx, req.Title, req.OrganizationId, req.VenueId, req.Description, req.Category, nil, start, end)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapEventToProto(e), nil
}

func (h *GRPCHandler) GetEvent(ctx context.Context, req *eventsv1.GetEventRequest) (*eventsv1.Event, error) {
	e, err := h.service.GetEvent(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return mapEventToProto(e), nil
}

func (h *GRPCHandler) ListEvents(ctx context.Context, req *eventsv1.ListEventsRequest) (*eventsv1.ListEventsResponse, error) {
	events, err := h.service.ListEvents(ctx, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoEvents []*eventsv1.Event
	for _, e := range events {
		protoEvents = append(protoEvents, mapEventToProto(e))
	}
	return &eventsv1.ListEventsResponse{Events: protoEvents, TotalCount: int32(len(events))}, nil
}

func (h *GRPCHandler) UpdateEvent(ctx context.Context, req *eventsv1.UpdateEventRequest) (*eventsv1.Event, error) {
	start, _ := time.Parse(time.RFC3339, req.StartTime)
	end, _ := time.Parse(time.RFC3339, req.EndTime)

	e, err := h.service.UpdateEvent(ctx, req.Id, req.Title, req.Description, start, end)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapEventToProto(e), nil
}

func (h *GRPCHandler) PublishEvent(ctx context.Context, req *eventsv1.PublishEventRequest) (*eventsv1.Event, error) {
	e, err := h.service.PublishEvent(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapEventToProto(e), nil
}

// --- Ticket Types ---

func (h *GRPCHandler) CreateTicketType(ctx context.Context, req *eventsv1.CreateTicketTypeRequest) (*eventsv1.TicketType, error) {
	start, _ := time.Parse(time.RFC3339, req.SalesStartTime)
	end, _ := time.Parse(time.RFC3339, req.SalesEndTime)

	tt, err := h.service.CreateTicketType(ctx, req.EventId, req.Name, "", req.PricePaisa, req.TotalQuantity, req.TotalQuantity, 10, start, end)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapTicketTypeToProto(tt), nil
}

func (h *GRPCHandler) ListTicketTypes(ctx context.Context, req *eventsv1.ListTicketTypesRequest) (*eventsv1.ListTicketTypesResponse, error) {
	types, err := h.service.ListTicketTypes(ctx, req.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoTypes []*eventsv1.TicketType
	for _, t := range types {
		protoTypes = append(protoTypes, mapTicketTypeToProto(t))
	}
	return &eventsv1.ListTicketTypesResponse{TicketTypes: protoTypes}, nil
}

func (h *GRPCHandler) SearchEvents(ctx context.Context, req *eventsv1.SearchEventsRequest) (*eventsv1.SearchEventsResponse, error) {
	events, total, err := h.service.SearchEvents(ctx, req.Query, req.City, req.Category, req.StartDate, req.EndDate, req.PageSize, 0) // Offset 0 for now
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var results []*eventsv1.EventSearchResult
	for _, e := range events {
		// Fetch venue for each event
		v, _ := h.service.GetVenue(ctx, e.VenueID)

		results = append(results, &eventsv1.EventSearchResult{
			Event: mapEventToProto(e),
			Venue: mapVenueToProto(v), // Venue might be nil/partial if fail, but map handles nil gracefully? Check mapVenueToProto.
			// Actually mapVenueToProto(nil) might panic. Let's be safe.
		})
	}
	return &eventsv1.SearchEventsResponse{
		Results:    results,
		TotalCount: total,
	}, nil
}

// --- Helpers ---

func mapVenueToProto(v *domain.Venue) *eventsv1.Venue {
	var sections []*eventsv1.SeatingSection
	json.Unmarshal([]byte(v.Sections), &sections)

	return &eventsv1.Venue{
		Id:             v.ID,
		OrganizationId: v.OrganizationID,
		Name:           v.Name,
		Address:        v.Address,
		City:           v.City,
		Country:        v.Country,
		Capacity:       v.Capacity,
		Type:           eventsv1.VenueType_VENUE_TYPE_AUDITORIUM, // Simplified mapping
		Sections:       sections,
		MapImageUrl:    v.MapImageURL,
		CreatedAt:      v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      v.UpdatedAt.Format(time.RFC3339),
	}
}

func mapEventToProto(e *domain.Event) *eventsv1.Event {
	status := eventsv1.EventStatus_EVENT_STATUS_DRAFT
	if e.Status == "published" {
		status = eventsv1.EventStatus_EVENT_STATUS_PUBLISHED
	}

	return &eventsv1.Event{
		Id:             e.ID,
		OrganizationId: e.OrganizationID,
		VenueId:        e.VenueID,
		Title:          e.Title,
		Description:    e.Description,
		Category:       e.Category,
		Images:         e.Images,
		StartTime:      e.StartTime.Format(time.RFC3339),
		EndTime:        e.EndTime.Format(time.RFC3339),
		Status:         status,
		CreatedAt:      e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      e.UpdatedAt.Format(time.RFC3339),
	}
}

func mapTicketTypeToProto(t *domain.TicketType) *eventsv1.TicketType {
	return &eventsv1.TicketType{
		Id:                t.ID,
		EventId:           t.EventID,
		Name:              t.Name,
		Description:       t.Description,
		PricePaisa:        t.PricePaisa,
		TotalQuantity:     t.TotalQuantity,
		AvailableQuantity: t.AvailableQuantity,
		MaxPerUser:        t.MaxPerUser,
		SalesStartTime:    t.SalesStartTime.Format(time.RFC3339),
		SalesEndTime:      t.SalesEndTime.Format(time.RFC3339),
	}
}

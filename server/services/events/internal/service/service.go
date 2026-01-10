package service

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/events/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/events/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// --- Venues ---

func (s *EventService) CreateVenue(ctx context.Context, name, orgID, address, city, country string, capacity int32, vType, mapImageURL, sections string) (*domain.Venue, error) {
	venue := &domain.Venue{
		OrganizationID: orgID,
		Name:           name,
		Address:        address,
		City:           city,
		Country:        country,
		Capacity:       capacity,
		Type:           vType,
		Sections:       sections,
		MapImageURL:    mapImageURL,
	}

	if err := s.repo.CreateVenue(venue); err != nil {
		return nil, err
	}
	return venue, nil
}

func (s *EventService) GetVenue(ctx context.Context, id string) (*domain.Venue, error) {
	return s.repo.GetVenue(id)
}

func (s *EventService) ListVenues(ctx context.Context, orgID string) ([]*domain.Venue, error) {
	return s.repo.ListVenues(orgID)
}

// --- Events ---

func (s *EventService) CreateEvent(ctx context.Context, title, orgID, venueID, description, category string, images []string, start, end time.Time) (*domain.Event, error) {
	event := &domain.Event{
		OrganizationID: orgID,
		VenueID:        venueID,
		Title:          title,
		Description:    description,
		Category:       category,
		Images:         images,
		StartTime:      start,
		EndTime:        end,
		Status:         "draft",
	}

	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	return s.repo.GetEvent(id)
}

func (s *EventService) ListEvents(ctx context.Context, orgID string) ([]*domain.Event, error) {
	return s.repo.ListEvents(orgID)
}

// --- Ticket Types ---

func (s *EventService) CreateTicketType(ctx context.Context, eventID, name, desc string, price int64, total, avail, max int32, start, end time.Time) (*domain.TicketType, error) {
	tt := &domain.TicketType{
		EventID:           eventID,
		Name:              name,
		Description:       desc,
		PricePaisa:        price,
		TotalQuantity:     total,
		AvailableQuantity: avail,
		MaxPerUser:        max,
		SalesStartTime:    start,
		SalesEndTime:      end,
	}

	if err := s.repo.CreateTicketType(tt); err != nil {
		return nil, err
	}
	return tt, nil
}

func (s *EventService) ListTicketTypes(ctx context.Context, eventID string) ([]*domain.TicketType, error) {
	return s.repo.ListTicketTypes(eventID)
}

package service

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/services/events/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/events/internal/repository"
	"github.com/google/uuid"
)

type EventService struct {
	repo     *repository.EventRepository
	producer *kafka.Producer
}

func NewEventService(repo *repository.EventRepository, producer *kafka.Producer) *EventService {
	return &EventService{repo: repo, producer: producer}
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

func (s *EventService) UpdateVenue(ctx context.Context, id, name, vType string) (*domain.Venue, error) {
	venue, err := s.repo.GetVenue(id)
	if err != nil {
		return nil, err
	}
	venue.Name = name
	venue.Type = vType
	if err := s.repo.UpdateVenue(venue); err != nil {
		return nil, err
	}
	return venue, nil
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
	// Publish to Kafka
	s.producer.Publish(ctx, kafka.TopicEvents, &kafka.Event{
		ID:          uuid.New().String(),
		Type:        kafka.EventEventCreated,
		AggregateID: event.ID,
		Timestamp:   time.Now(),
		Payload:     event,
	})
	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	return s.repo.GetEvent(id)
}

func (s *EventService) ListEvents(ctx context.Context, orgID string) ([]*domain.Event, error) {
	return s.repo.ListEvents(orgID)
}

func (s *EventService) UpdateEvent(ctx context.Context, id, title, description string, start, end time.Time) (*domain.Event, error) {
	event, err := s.repo.GetEvent(id)
	if err != nil {
		return nil, err
	}
	event.Title = title
	event.Description = description
	event.StartTime = start
	event.EndTime = end

	if err := s.repo.UpdateEvent(event); err != nil {
		return nil, err
	}

	// Publish to Kafka
	s.producer.Publish(ctx, kafka.TopicEvents, &kafka.Event{
		ID:          uuid.New().String(),
		Type:        kafka.EventEventUpdated,
		AggregateID: event.ID,
		Timestamp:   time.Now(),
		Payload:     event,
	})

	return event, nil
}

func (s *EventService) PublishEvent(ctx context.Context, id string) (*domain.Event, error) {
	if err := s.repo.UpdateEventStatus(id, "published"); err != nil {
		return nil, err
	}
	event, err := s.repo.GetEvent(id)
	if err == nil {
		s.producer.Publish(ctx, kafka.TopicEvents, &kafka.Event{
			ID:          uuid.New().String(),
			Type:        kafka.EventEventPublished,
			AggregateID: event.ID,
			Timestamp:   time.Now(),
			Payload:     event,
		})
	}
	return event, err
}

func (s *EventService) SearchEvents(ctx context.Context, query, city, category, start, end string, limit, offset int32) ([]*domain.Event, int32, error) {
	events, total, err := s.repo.SearchEvents(query, city, category, start, end, int(limit), int(offset))
	if err != nil {
		return nil, 0, err
	}
	return events, int32(total), nil
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

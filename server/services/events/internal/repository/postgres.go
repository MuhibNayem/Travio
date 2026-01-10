package repository

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/MuhibNayem/Travio/server/services/events/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrVenueNotFound = errors.New("venue not found")
	ErrEventNotFound = errors.New("event not found")
)

type EventRepository struct {
	DB *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{DB: db}
}

// --- Venue Methods ---

func (r *EventRepository) CreateVenue(venue *domain.Venue) error {
	venue.ID = uuid.New().String()
	venue.CreatedAt = time.Now()
	venue.UpdatedAt = time.Now()

	query := `INSERT INTO venues (id, organization_id, name, address, city, country, capacity, type, sections, map_image_url, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.DB.Exec(query, venue.ID, venue.OrganizationID, venue.Name, venue.Address, venue.City, venue.Country, venue.Capacity, venue.Type, venue.Sections, venue.MapImageURL, venue.CreatedAt, venue.UpdatedAt)
	return err
}

func (r *EventRepository) GetVenue(id string) (*domain.Venue, error) {
	query := `SELECT id, organization_id, name, address, city, country, capacity, type, sections, map_image_url, created_at, updated_at FROM venues WHERE id = $1`
	row := r.DB.QueryRow(query, id)

	var v domain.Venue
	var sectionsJSON []byte
	err := row.Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Address, &v.City, &v.Country, &v.Capacity, &v.Type, &sectionsJSON, &v.MapImageURL, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVenueNotFound
		}
		return nil, err
	}
	v.Sections = string(sectionsJSON)
	return &v, nil
}

func (r *EventRepository) ListVenues(orgID string) ([]*domain.Venue, error) {
	query := `SELECT id, organization_id, name, address, city, country, capacity, type, sections, map_image_url, created_at, updated_at FROM venues WHERE organization_id = $1`
	rows, err := r.DB.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		var v domain.Venue
		var sectionsJSON []byte
		if err := rows.Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Address, &v.City, &v.Country, &v.Capacity, &v.Type, &sectionsJSON, &v.MapImageURL, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		v.Sections = string(sectionsJSON)
		venues = append(venues, &v)
	}
	return venues, nil
}

func (r *EventRepository) UpdateVenue(venue *domain.Venue) error {
	venue.UpdatedAt = time.Now()
	query := `UPDATE venues SET name = $1, type = $2, updated_at = $3 WHERE id = $4`
	_, err := r.DB.Exec(query, venue.Name, venue.Type, venue.UpdatedAt, venue.ID)
	return err
}

// --- Event Methods ---

func (r *EventRepository) CreateEvent(event *domain.Event) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	query := `INSERT INTO events (id, organization_id, venue_id, title, description, category, images, start_time, end_time, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.DB.Exec(query, event.ID, event.OrganizationID, event.VenueID, event.Title, event.Description, event.Category, pq.Array(event.Images), event.StartTime, event.EndTime, event.Status, event.CreatedAt, event.UpdatedAt)
	return err
}

func (r *EventRepository) GetEvent(id string) (*domain.Event, error) {
	query := `SELECT id, organization_id, venue_id, title, description, category, images, start_time, end_time, status, created_at, updated_at FROM events WHERE id = $1`
	row := r.DB.QueryRow(query, id)

	var e domain.Event
	err := row.Scan(&e.ID, &e.OrganizationID, &e.VenueID, &e.Title, &e.Description, &e.Category, pq.Array(&e.Images), &e.StartTime, &e.EndTime, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *EventRepository) ListEvents(orgID string) ([]*domain.Event, error) {
	query := `SELECT id, organization_id, venue_id, title, description, category, images, start_time, end_time, status, created_at, updated_at FROM events WHERE organization_id = $1`
	rows, err := r.DB.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.VenueID, &e.Title, &e.Description, &e.Category, pq.Array(&e.Images), &e.StartTime, &e.EndTime, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

func (r *EventRepository) UpdateEvent(event *domain.Event) error {
	event.UpdatedAt = time.Now()
	query := `UPDATE events SET title = $1, description = $2, start_time = $3, end_time = $4, updated_at = $5 WHERE id = $6`
	_, err := r.DB.Exec(query, event.Title, event.Description, event.StartTime, event.EndTime, event.UpdatedAt, event.ID)
	return err
}

func (r *EventRepository) UpdateEventStatus(id, status string) error {
	query := `UPDATE events SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.DB.Exec(query, status, time.Now(), id)
	return err
}

func (r *EventRepository) SearchEvents(queryStr, city, category, start, end string, limit, offset int) ([]*domain.Event, int, error) {
	// Base query structure
	baseQuery := `SELECT e.id, e.organization_id, e.venue_id, e.title, e.description, e.category, e.images, e.start_time, e.end_time, e.status, e.created_at, e.updated_at 
	              FROM events e 
	              JOIN venues v ON e.venue_id = v.id 
	              WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM events e JOIN venues v ON e.venue_id = v.id WHERE 1=1`

	var args []interface{}
	idx := 1

	// Helper to add condition
	addCondition := func(condition string, val interface{}) {
		baseQuery += " AND " + condition
		countQuery += " AND " + condition
		args = append(args, val)
		idx++
	}

	if queryStr != "" {
		// Use manual positional parameters based on current idx
		cond := "(e.title ILIKE $" + strconv.Itoa(idx) + " OR e.description ILIKE $" + strconv.Itoa(idx) + ")"
		addCondition(cond, "%"+queryStr+"%")
	}
	if city != "" {
		cond := "v.city ILIKE $" + strconv.Itoa(idx)
		addCondition(cond, city)
	}
	if category != "" {
		cond := "e.category = $" + strconv.Itoa(idx)
		addCondition(cond, category)
	}
	if start != "" {
		cond := "e.start_time >= $" + strconv.Itoa(idx)
		addCondition(cond, start)
	}
	if end != "" {
		cond := "e.end_time <= $" + strconv.Itoa(idx)
		addCondition(cond, end)
	}

	// 1. Get Total Count
	var total int
	// We need to execute count query with same args
	if err := r.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2. Get Data with Pagination
	baseQuery += " ORDER BY e.start_time ASC LIMIT $" + strconv.Itoa(idx) + " OFFSET $" + strconv.Itoa(idx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.VenueID, &e.Title, &e.Description, &e.Category, pq.Array(&e.Images), &e.StartTime, &e.EndTime, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		events = append(events, &e)
	}

	return events, total, nil
}

func (r *EventRepository) CreateTicketType(tt *domain.TicketType) error {
	tt.ID = uuid.New().String()
	tt.CreatedAt = time.Now()
	tt.UpdatedAt = time.Now()

	query := `INSERT INTO ticket_types (id, event_id, name, description, price_paisa, total_quantity, available_quantity, max_per_user, sales_start_time, sales_end_time, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// Handle optional times (can be null)
	var salesStart, salesEnd interface{}
	if !tt.SalesStartTime.IsZero() {
		salesStart = tt.SalesStartTime
	}
	if !tt.SalesEndTime.IsZero() {
		salesEnd = tt.SalesEndTime
	}

	_, err := r.DB.Exec(query, tt.ID, tt.EventID, tt.Name, tt.Description, tt.PricePaisa, tt.TotalQuantity, tt.AvailableQuantity, tt.MaxPerUser, salesStart, salesEnd, tt.CreatedAt, tt.UpdatedAt)
	return err
}

func (r *EventRepository) ListTicketTypes(eventID string) ([]*domain.TicketType, error) {
	query := `SELECT id, event_id, name, description, price_paisa, total_quantity, available_quantity, max_per_user, sales_start_time, sales_end_time, created_at, updated_at FROM ticket_types WHERE event_id = $1`
	rows, err := r.DB.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*domain.TicketType
	for rows.Next() {
		var t domain.TicketType
		var start, end sql.NullTime
		if err := rows.Scan(&t.ID, &t.EventID, &t.Name, &t.Description, &t.PricePaisa, &t.TotalQuantity, &t.AvailableQuantity, &t.MaxPerUser, &start, &end, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if start.Valid {
			t.SalesStartTime = start.Time
		}
		if end.Valid {
			t.SalesEndTime = end.Time
		}
		types = append(types, &t)
	}
	return types, nil
}

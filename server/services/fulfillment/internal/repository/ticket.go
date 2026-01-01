package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/domain"
	"github.com/google/uuid"
)

var ErrTicketNotFound = errors.New("ticket not found")

type TicketRepository struct {
	DB *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{DB: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *domain.Ticket) error {
	ticket.ID = uuid.New().String()
	ticket.CreatedAt = time.Now()

	query := `INSERT INTO tickets (
		id, booking_id, order_id, organization_id, trip_id, route_name,
		from_station, to_station, departure_time, arrival_time,
		passenger_nid, passenger_name, seat_number, seat_class,
		price_paisa, currency, qr_code_data, qr_code_url,
		status, created_at, valid_until
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.DB.ExecContext(ctx, query,
		ticket.ID, ticket.BookingID, ticket.OrderID, ticket.OrganizationID,
		ticket.TripID, ticket.RouteName, ticket.FromStation, ticket.ToStation,
		ticket.DepartureTime, ticket.ArrivalTime, ticket.PassengerNID, ticket.PassengerName,
		ticket.SeatNumber, ticket.SeatClass, ticket.PricePaisa, ticket.Currency,
		ticket.QRCodeData, ticket.QRCodeURL, ticket.Status, ticket.CreatedAt, ticket.ValidUntil,
	)
	return err
}

func (r *TicketRepository) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	query := `SELECT id, booking_id, order_id, organization_id, trip_id, route_name,
		from_station, to_station, departure_time, arrival_time,
		passenger_nid, passenger_name, seat_number, seat_class,
		price_paisa, currency, qr_code_data, qr_code_url,
		status, created_at, valid_until, is_boarded, boarded_at, boarded_by
		FROM tickets WHERE id = $1`

	var t domain.Ticket
	var boardedAt sql.NullTime
	var boardedBy sql.NullString

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.BookingID, &t.OrderID, &t.OrganizationID, &t.TripID, &t.RouteName,
		&t.FromStation, &t.ToStation, &t.DepartureTime, &t.ArrivalTime,
		&t.PassengerNID, &t.PassengerName, &t.SeatNumber, &t.SeatClass,
		&t.PricePaisa, &t.Currency, &t.QRCodeData, &t.QRCodeURL,
		&t.Status, &t.CreatedAt, &t.ValidUntil, &t.IsBoarded, &boardedAt, &boardedBy,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	if boardedAt.Valid {
		t.BoardedAt = boardedAt.Time
	}
	if boardedBy.Valid {
		t.BoardedBy = boardedBy.String
	}

	return &t, nil
}

func (r *TicketRepository) ListByOrder(ctx context.Context, orderID string) ([]*domain.Ticket, error) {
	query := `SELECT id, booking_id, order_id, organization_id, trip_id, route_name,
		from_station, to_station, departure_time, arrival_time,
		passenger_nid, passenger_name, seat_number, seat_class,
		price_paisa, currency, qr_code_data, qr_code_url,
		status, created_at, valid_until, is_boarded
		FROM tickets WHERE order_id = $1 ORDER BY seat_number`

	rows, err := r.DB.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.BookingID, &t.OrderID, &t.OrganizationID, &t.TripID, &t.RouteName,
			&t.FromStation, &t.ToStation, &t.DepartureTime, &t.ArrivalTime,
			&t.PassengerNID, &t.PassengerName, &t.SeatNumber, &t.SeatClass,
			&t.PricePaisa, &t.Currency, &t.QRCodeData, &t.QRCodeURL,
			&t.Status, &t.CreatedAt, &t.ValidUntil, &t.IsBoarded,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id string, status domain.TicketStatus) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, status, id)
	return err
}

func (r *TicketRepository) MarkAsBoarded(ctx context.Context, id, boardedBy string) error {
	query := `UPDATE tickets SET is_boarded = true, boarded_at = $1, boarded_by = $2, status = $3 WHERE id = $4`
	_, err := r.DB.ExecContext(ctx, query, time.Now(), boardedBy, domain.TicketStatusUsed, id)
	return err
}

// CreateBatch creates multiple tickets in a transaction
func (r *TicketRepository) CreateBatch(ctx context.Context, tickets []*domain.Ticket) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO tickets (
		id, booking_id, order_id, organization_id, trip_id, route_name,
		from_station, to_station, departure_time, arrival_time,
		passenger_nid, passenger_name, seat_number, seat_class,
		price_paisa, currency, qr_code_data, qr_code_url,
		status, created_at, valid_until
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickets {
		t.ID = uuid.New().String()
		t.CreatedAt = time.Now()

		_, err := stmt.ExecContext(ctx,
			t.ID, t.BookingID, t.OrderID, t.OrganizationID,
			t.TripID, t.RouteName, t.FromStation, t.ToStation,
			t.DepartureTime, t.ArrivalTime, t.PassengerNID, t.PassengerName,
			t.SeatNumber, t.SeatClass, t.PricePaisa, t.Currency,
			t.QRCodeData, t.QRCodeURL, t.Status, t.CreatedAt, t.ValidUntil,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// For testing
func (r *TicketRepository) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

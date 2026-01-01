package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrDuplicateOrder = errors.New("duplicate idempotency key")
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	passengersJSON, _ := json.Marshal(order.Passengers)
	seatsJSON, _ := json.Marshal(order.Seats)

	query := `INSERT INTO orders (
		id, organization_id, user_id, trip_id, from_station_id, to_station_id,
		passengers, subtotal_paisa, tax_paisa, booking_fee_paisa, discount_paisa, total_paisa, currency,
		payment_id, payment_status, payment_method, booking_id, hold_id, seats,
		status, saga_id, contact_email, contact_phone, created_at, updated_at, expires_at, idempotency_key
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)`

	_, err := r.DB.ExecContext(ctx, query,
		order.ID, order.OrganizationID, order.UserID, order.TripID, order.FromStationID, order.ToStationID,
		passengersJSON, order.SubtotalPaisa, order.TaxPaisa, order.BookingFeePaisa, order.DiscountPaisa, order.TotalPaisa, order.Currency,
		order.PaymentID, order.PaymentStatus, order.PaymentMethod, order.BookingID, order.HoldID, seatsJSON,
		order.Status, order.SagaID, order.ContactEmail, order.ContactPhone, order.CreatedAt, order.UpdatedAt, order.ExpiresAt, order.IdempotencyKey,
	)

	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id, userID string) (*domain.Order, error) {
	query := `SELECT 
		id, organization_id, user_id, trip_id, from_station_id, to_station_id,
		passengers, subtotal_paisa, tax_paisa, booking_fee_paisa, discount_paisa, total_paisa, currency,
		payment_id, payment_status, payment_method, booking_id, hold_id, seats,
		status, saga_id, contact_email, contact_phone, created_at, updated_at, expires_at, idempotency_key
		FROM orders WHERE id = $1 AND user_id = $2`

	var order domain.Order
	var passengersJSON, seatsJSON []byte

	err := r.DB.QueryRowContext(ctx, query, id, userID).Scan(
		&order.ID, &order.OrganizationID, &order.UserID, &order.TripID, &order.FromStationID, &order.ToStationID,
		&passengersJSON, &order.SubtotalPaisa, &order.TaxPaisa, &order.BookingFeePaisa, &order.DiscountPaisa, &order.TotalPaisa, &order.Currency,
		&order.PaymentID, &order.PaymentStatus, &order.PaymentMethod, &order.BookingID, &order.HoldID, &seatsJSON,
		&order.Status, &order.SagaID, &order.ContactEmail, &order.ContactPhone, &order.CreatedAt, &order.UpdatedAt, &order.ExpiresAt, &order.IdempotencyKey,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	json.Unmarshal(passengersJSON, &order.Passengers)
	json.Unmarshal(seatsJSON, &order.Seats)

	return &order, nil
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	query := `SELECT id FROM orders WHERE idempotency_key = $1`

	var orderID string
	err := r.DB.QueryRowContext(ctx, query, key).Scan(&orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is OK for idempotency check
		}
		return nil, err
	}

	// Found existing order - return it
	return r.GetByID(ctx, orderID, "")
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	order.UpdatedAt = time.Now()

	passengersJSON, _ := json.Marshal(order.Passengers)
	seatsJSON, _ := json.Marshal(order.Seats)

	query := `UPDATE orders SET
		passengers = $1, payment_id = $2, payment_status = $3, booking_id = $4, seats = $5,
		status = $6, saga_id = $7, updated_at = $8
		WHERE id = $9`

	_, err := r.DB.ExecContext(ctx, query,
		passengersJSON, order.PaymentID, order.PaymentStatus, order.BookingID, seatsJSON,
		order.Status, order.SagaID, order.UpdatedAt, order.ID,
	)

	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.DB.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID string, status string, limit, offset int) ([]*domain.Order, int, error) {
	var args []interface{}
	whereClause := "WHERE user_id = $1"
	args = append(args, userID)

	if status != "" {
		whereClause += " AND status = $2"
		args = append(args, status)
	}

	// Count
	countQuery := "SELECT COUNT(*) FROM orders " + whereClause
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Query
	query := `SELECT 
		id, organization_id, user_id, trip_id, from_station_id, to_station_id,
		passengers, subtotal_paisa, tax_paisa, booking_fee_paisa, discount_paisa, total_paisa, currency,
		payment_id, payment_status, payment_method, booking_id, hold_id, seats,
		status, saga_id, contact_email, contact_phone, created_at, updated_at, expires_at
		FROM orders ` + whereClause + ` ORDER BY created_at DESC LIMIT $` +
		string(rune('0'+len(args)+1)) + ` OFFSET $` + string(rune('0'+len(args)+2))

	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		var passengersJSON, seatsJSON []byte

		if err := rows.Scan(
			&o.ID, &o.OrganizationID, &o.UserID, &o.TripID, &o.FromStationID, &o.ToStationID,
			&passengersJSON, &o.SubtotalPaisa, &o.TaxPaisa, &o.BookingFeePaisa, &o.DiscountPaisa, &o.TotalPaisa, &o.Currency,
			&o.PaymentID, &o.PaymentStatus, &o.PaymentMethod, &o.BookingID, &o.HoldID, &seatsJSON,
			&o.Status, &o.SagaID, &o.ContactEmail, &o.ContactPhone, &o.CreatedAt, &o.UpdatedAt, &o.ExpiresAt,
		); err != nil {
			return nil, 0, err
		}

		json.Unmarshal(passengersJSON, &o.Passengers)
		json.Unmarshal(seatsJSON, &o.Seats)
		orders = append(orders, &o)
	}

	return orders, total, nil
}

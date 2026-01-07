package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/google/uuid"
)

// TxOrderRepository provides transactional order operations
type TxOrderRepository struct {
	tx *sql.Tx
}

// NewTxOrderRepository creates a transactional repository
func NewTxOrderRepository(tx *sql.Tx) *TxOrderRepository {
	return &TxOrderRepository{tx: tx}
}

// BeginTx starts a new transaction
func (r *OrderRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.DB.BeginTx(ctx, nil)
}

// CreateTx creates an order within a transaction
func (r *TxOrderRepository) CreateTx(ctx context.Context, order *domain.Order) error {
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

	_, err := r.tx.ExecContext(ctx, query,
		order.ID, order.OrganizationID, order.UserID, order.TripID, order.FromStationID, order.ToStationID,
		passengersJSON, order.SubtotalPaisa, order.TaxPaisa, order.BookingFeePaisa, order.DiscountPaisa, order.TotalPaisa, order.Currency,
		order.PaymentID, order.PaymentStatus, order.PaymentMethod, order.BookingID, order.HoldID, seatsJSON,
		order.Status, order.SagaID, order.ContactEmail, order.ContactPhone, order.CreatedAt, order.UpdatedAt, order.ExpiresAt, order.IdempotencyKey,
	)

	return err
}

// UpdateTx updates an order within a transaction
func (r *TxOrderRepository) UpdateTx(ctx context.Context, order *domain.Order) error {
	order.UpdatedAt = time.Now()

	passengersJSON, _ := json.Marshal(order.Passengers)
	seatsJSON, _ := json.Marshal(order.Seats)

	query := `UPDATE orders SET
		passengers = $1, payment_id = $2, payment_status = $3, booking_id = $4, seats = $5,
		status = $6, saga_id = $7, updated_at = $8
		WHERE id = $9`

	_, err := r.tx.ExecContext(ctx, query,
		passengersJSON, order.PaymentID, order.PaymentStatus, order.BookingID, seatsJSON,
		order.Status, order.SagaID, order.UpdatedAt, order.ID,
	)

	return err
}

// UpdateStatusTx updates order status within a transaction
func (r *TxOrderRepository) UpdateStatusTx(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// Tx returns the underlying transaction
func (r *TxOrderRepository) Tx() *sql.Tx {
	return r.tx
}

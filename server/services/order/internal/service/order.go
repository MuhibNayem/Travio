package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/order/internal/events"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
)

const (
	DefaultCurrency = "BDT"
	TaxRate         = 0.05 // 5% VAT
	BookingFeePaisa = 2000 // 20 BDT per passenger
)

type OrderService struct {
	db           *sql.DB
	orderRepo    *repository.OrderRepository
	sagaDeps     *saga.BookingDependencies
	orchestrator *saga.Orchestrator
	publisher    *events.Publisher
}

func NewOrderService(
	db *sql.DB,
	orderRepo *repository.OrderRepository,
	sagaDeps *saga.BookingDependencies,
) *OrderService {
	return &OrderService{
		db:           db,
		orderRepo:    orderRepo,
		sagaDeps:     sagaDeps,
		orchestrator: saga.NewOrchestrator(),
		publisher:    events.NewPublisher(db),
	}
}

// CreateOrder initiates the booking saga with transactional outbox event
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*domain.Order, error) {
	// Idempotency check
	if req.IdempotencyKey != "" {
		existing, err := s.orderRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil // Return existing order
		}
	}

	// Start transaction
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create order record
	order := &domain.Order{
		OrganizationID: req.OrgID,
		UserID:         req.UserID,
		TripID:         req.TripID,
		FromStationID:  req.FromStation,
		ToStationID:    req.ToStation,
		HoldID:         req.HoldID,
		Passengers:     convertPassengers(req.Passengers),
		PaymentMethod:  req.PaymentMethod,
		PaymentStatus:  domain.PaymentStatusPending,
		Status:         domain.OrderStatusPending,
		ContactEmail:   req.Email,
		ContactPhone:   req.Phone,
		Currency:       DefaultCurrency,
		ExpiresAt:      time.Now().Add(15 * time.Minute),
		IdempotencyKey: req.IdempotencyKey,
	}

	// Calculate totals (in production, fetch prices from catalog)
	basePrices := make(map[string]int64)
	for _, p := range req.Passengers {
		basePrices[p.SeatID] = 80000 // 800 BDT placeholder
	}
	order.CalculateTotals(basePrices, TaxRate, BookingFeePaisa)

	// Create order in transaction
	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.CreateTx(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish OrderCreated event to outbox (same transaction)
	if err := s.publisher.PublishOrderCreated(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to publish order created event: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Create booking saga
	bookingReq := &saga.BookingRequest{
		OrderID:       order.ID,
		UserID:        order.UserID,
		OrgID:         order.OrganizationID,
		TripID:        order.TripID,
		HoldID:        order.HoldID,
		FromStation:   order.FromStationID,
		ToStation:     order.ToStationID,
		Passengers:    convertToSagaPassengers(req.Passengers),
		PaymentToken:  req.PaymentToken,
		PaymentMethod: req.PaymentMethod,
		TotalPaisa:    order.TotalPaisa,
		Email:         order.ContactEmail,
		Phone:         order.ContactPhone,
	}

	sagaInstance := saga.NewBookingSaga(s.sagaDeps, bookingReq)
	order.SagaID = sagaInstance.ID

	// Update order with saga ID
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order with saga ID: %w", err)
	}

	// Execute saga asynchronously with outbox event on completion
	go func() {
		execCtx := context.Background()
		if err := s.orchestrator.Execute(execCtx, sagaInstance); err != nil {
			// Update order status on failure and publish event
			s.handleOrderFailed(execCtx, order, err.Error(), fmt.Sprintf("%v", sagaInstance.Status))
		} else {
			// Update order status on success and publish event
			s.handleOrderConfirmed(execCtx, order, sagaInstance)
		}
	}()

	return order, nil
}

// handleOrderConfirmed updates order and publishes confirmation event
func (s *OrderService) handleOrderConfirmed(ctx context.Context, order *domain.Order, sagaInstance *saga.Saga) {
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	order.Status = domain.OrderStatusConfirmed
	order.PaymentStatus = domain.PaymentStatusCaptured
	order.BookingID = sagaInstance.Context.GetString("booking_id")
	order.PaymentID = sagaInstance.Context.GetString("payment_id")

	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.UpdateTx(ctx, order); err != nil {
		return
	}

	if err := s.publisher.PublishOrderConfirmed(ctx, tx, order); err != nil {
		return
	}

	tx.Commit()
}

// handleOrderFailed updates order and publishes failure event
func (s *OrderService) handleOrderFailed(ctx context.Context, order *domain.Order, reason, sagaState string) {
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	order.Status = domain.OrderStatusFailed

	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.UpdateStatusTx(ctx, order.ID, domain.OrderStatusFailed); err != nil {
		return
	}

	if err := s.publisher.PublishOrderFailed(ctx, tx, order, reason, sagaState); err != nil {
		return
	}

	tx.Commit()
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, orderID, userID string) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID, userID)
}

// ListOrders retrieves user's orders
func (s *OrderService) ListOrders(ctx context.Context, userID, status string, pageSize int, pageToken string) ([]*domain.Order, int, string, error) {
	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}

	orders, total, err := s.orderRepo.ListByUser(ctx, userID, status, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return orders, total, nextToken, nil
}

// CancelOrder initiates the cancellation saga with transactional outbox event
func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID, reason string) (*domain.Order, *RefundInfo, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check if cancellable
	if order.Status != domain.OrderStatusConfirmed {
		return nil, nil, fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	// Create cancellation saga
	cancellationSaga := saga.NewCancellationSaga(
		s.sagaDeps,
		order.ID,
		order.UserID,
		order.BookingID,
		order.PaymentID,
		order.ContactEmail,
		order.ContactPhone,
		order.TotalPaisa,
	)

	// Execute cancellation saga
	if err := s.orchestrator.Execute(ctx, cancellationSaga); err != nil {
		return nil, nil, fmt.Errorf("cancellation failed: %w", err)
	}

	// Start transaction for final update
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	refundID := cancellationSaga.Context.GetString("refund_id")

	// Update order status
	order.Status = domain.OrderStatusRefunded
	order.PaymentStatus = domain.PaymentStatusRefunded

	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.UpdateTx(ctx, order); err != nil {
		return nil, nil, err
	}

	// Publish cancellation event
	if err := s.publisher.PublishOrderCancelled(ctx, tx, order, refundID, order.TotalPaisa, reason); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	refund := &RefundInfo{
		RefundID:    refundID,
		AmountPaisa: order.TotalPaisa,
		Status:      "completed",
	}

	return order, refund, nil
}

// GetOrderStatus returns order and saga status
func (s *OrderService) GetOrderStatus(ctx context.Context, orderID, userID string) (*domain.Order, *saga.Saga, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, userID)
	if err != nil {
		return nil, nil, err
	}

	sagaInstance, _ := s.orchestrator.GetSaga(order.SagaID)
	return order, sagaInstance, nil
}

// --- DTOs ---

type CreateOrderRequest struct {
	OrgID          string
	UserID         string
	TripID         string
	FromStation    string
	ToStation      string
	HoldID         string
	Passengers     []PassengerRequest
	PaymentToken   string
	PaymentMethod  string
	Email          string
	Phone          string
	CouponCode     string
	IdempotencyKey string
}

type PassengerRequest struct {
	NID         string
	Name        string
	SeatID      string
	DateOfBirth string
	Gender      string
	Age         int
}

type RefundInfo struct {
	RefundID    string
	AmountPaisa int64
	Status      string
}

// --- Helpers ---

func convertPassengers(reqs []PassengerRequest) []domain.OrderPassenger {
	var passengers []domain.OrderPassenger
	for _, r := range reqs {
		passengers = append(passengers, domain.OrderPassenger{
			NID:    r.NID,
			Name:   r.Name,
			SeatID: r.SeatID,
			Gender: r.Gender,
			Age:    r.Age,
		})
	}
	return passengers
}

func convertToSagaPassengers(reqs []PassengerRequest) []saga.PassengerInfo {
	var passengers []saga.PassengerInfo
	for _, r := range reqs {
		passengers = append(passengers, saga.PassengerInfo{
			NID:         r.NID,
			Name:        r.Name,
			SeatID:      r.SeatID,
			DateOfBirth: r.DateOfBirth,
			Gender:      r.Gender,
		})
	}
	return passengers
}

func parsePageToken(token string) int {
	if token == "" {
		return 0
	}
	var offset int
	fmt.Sscanf(token, "%d", &offset)
	return offset
}

func generatePageToken(offset int) string {
	return fmt.Sprintf("%d", offset)
}

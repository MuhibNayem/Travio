package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/order/internal/events"
	"github.com/MuhibNayem/Travio/server/services/order/internal/messaging"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
	"gorm.io/gorm"
)

const (
	TaxRate         = 0.05 // 5% VAT
	BookingFeePaisa = 2000 // 20 BDT per passenger
)

type OrderService struct {
	db              *sql.DB
	gormDB          *gorm.DB
	orderRepo       *repository.OrderRepository
	checkoutRepo    *repository.CheckoutRepository
	sagaDeps        *saga.BookingDependencies
	orchestrator    *saga.Orchestrator
	publisher       *events.Publisher
	dlq             messaging.DLQProducer
	crmClient       CRMClient
}

// CRMClient interface for coupon validation
type CRMClient interface {
	ValidateCoupon(ctx context.Context, orgID, code string, cartTotal int64) (*domain.CouponValidation, error)
}

func NewOrderService(
	db *sql.DB,
	gormDB *gorm.DB,
	dlq messaging.DLQProducer,
	orderRepo *repository.OrderRepository,
	checkoutRepo *repository.CheckoutRepository,
	sagaDeps *saga.BookingDependencies,
	crmClient CRMClient,
) *OrderService {
	// Auto-migrate saga instances and checkout tables
	_ = gormDB.AutoMigrate(&saga.SagaInstance{})
	_ = checkoutRepo.AutoMigrate()

	return &OrderService{
		db:           db,
		gormDB:       gormDB,
		orderRepo:    orderRepo,
		checkoutRepo: checkoutRepo,
		sagaDeps:     sagaDeps,
		orchestrator: saga.NewOrchestrator(gormDB, dlq),
		publisher:    events.NewPublisher(db),
		dlq:          dlq,
		crmClient:    crmClient,
	}
}

// GetOrchestrator returns the saga orchestr for external use
func (s *OrderService) GetOrchestrator() *saga.Orchestrator {
	return s.orchestrator
}

// RecoverIncompleteSagas loads and resumes incomplete sagas on startup
func (s *OrderService) RecoverIncompleteSagas(ctx context.Context) error {
	var runningSagas []saga.SagaInstance
	if err := s.gormDB.WithContext(ctx).
		Where("status IN ?", []string{"running", "pending", "compensating"}).
		Find(&runningSagas).Error; err != nil {
		return fmt.Errorf("failed to query running sagas: %w", err)
	}

	logger.Info("Recovering incomplete sagas", "count", len(runningSagas))

	for _, sagaInst := range runningSagas {
		// Get the associated order
		order, err := s.orderRepo.GetByID(ctx, sagaInst.Name, "") // Name stores order ID
		if err != nil {
			logger.Warn("Failed to get order for saga recovery", "saga_id", sagaInst.ID, "error", err)
			continue
		}

		// Resume the saga
		go func(si saga.SagaInstance, o *domain.Order) {
			execCtx := context.Background()
			if err := s.orchestrator.Retry(execCtx, si.ID); err != nil {
				logger.Error("Saga recovery failed", "saga_id", si.ID, "error", err)
				s.handleOrderFailed(execCtx, o, "saga recovery failed: "+err.Error(), "recovery_failed")
			} else {
				logger.Info("Saga recovered successfully", "saga_id", si.ID)
			}
		}(sagaInst, order)
	}

	return nil
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
			return existing, nil
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
		DiscountPaisa:  req.DiscountPaisa,
	}

	// Calculate totals
	basePrices := make(map[string]int64)
	for _, p := range req.Passengers {
		basePrices[p.SeatID] = 50000 // 500 BDT placeholder
	}
	order.CalculateTotals(basePrices, TaxRate, BookingFeePaisa)

	// Validate and apply coupon if provided
	if req.CouponCode != "" && s.crmClient != nil {
		couponValidation, err := s.crmClient.ValidateCoupon(ctx, req.OrgID, req.CouponCode, order.TotalPaisa)
		if err != nil {
			logger.Warn("Coupon validation failed", "error", err, "code", req.CouponCode)
			// Don't fail order for invalid coupon, just log it
		} else if couponValidation.Valid {
			order.CouponCode = req.CouponCode
			order.CouponDiscount = couponValidation.DiscountPaisa
			order.DiscountPaisa = couponValidation.DiscountPaisa
			order.TotalPaisa -= couponValidation.DiscountPaisa
			if order.TotalPaisa < 0 {
				order.TotalPaisa = 0
			}
			logger.Info("Coupon applied successfully",
				"code", req.CouponCode,
				"discount_paisa", couponValidation.DiscountPaisa,
			)
		}
	}

	// Create order in transaction
	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.CreateTx(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish OrderCreated event to outbox
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

	// Execute saga asynchronously
	go func() {
		execCtx := context.Background()
		if err := s.orchestrator.Execute(execCtx, sagaInstance); err != nil {
			s.handleOrderFailed(execCtx, order, err.Error(), fmt.Sprintf("%v", sagaInstance.Status))
		} else {
			s.handleOrderConfirmed(execCtx, order, sagaInstance)
		}
	}()

	return order, nil
}

// CreateOrderFromCheckout creates an order from a confirmed checkout session
func (s *OrderService) CreateOrderFromCheckout(ctx context.Context, checkout *domain.CheckoutSession) (*domain.Order, error) {
	// Build order request from checkout session
	req := &CreateOrderRequest{
		OrgID:         checkout.OrganizationID,
		UserID:        checkout.UserID,
		TripID:        checkout.TripID,
		FromStation:   checkout.FromStationID,
		ToStation:     checkout.ToStationID,
		HoldID:        checkout.HoldID,
		PaymentMethod: checkout.PaymentMethod,
		Email:         checkout.Passengers[0].Email,
		Phone:         checkout.Passengers[0].Phone,
		DiscountPaisa: checkout.DiscountPaisa,
	}

	for _, p := range checkout.Passengers {
		req.Passengers = append(req.Passengers, PassengerRequest{
			NID:         p.NID,
			Name:        p.Name,
			SeatID:      p.SeatID,
			DateOfBirth: p.DateOfBirth,
			Gender:      p.Gender,
			Age:         p.Age,
		})
	}

	order, err := s.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Link checkout session to order
	if err := s.checkoutRepo.UpdateOrderID(ctx, checkout.ID, order.ID); err != nil {
		// Non-fatal, log and continue
	}

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

	if order.Status != domain.OrderStatusConfirmed {
		return nil, nil, fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

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

	if err := s.orchestrator.Execute(ctx, cancellationSaga); err != nil {
		return nil, nil, fmt.Errorf("cancellation failed: %w", err)
	}

	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	refundID := cancellationSaga.Context.GetString("refund_id")

	order.Status = domain.OrderStatusRefunded
	order.PaymentStatus = domain.PaymentStatusRefunded

	txRepo := repository.NewTxOrderRepository(tx)
	if err := txRepo.UpdateTx(ctx, order); err != nil {
		return nil, nil, err
	}

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

// GetSaga retrieves a saga by ID
func (s *OrderService) GetSaga(ctx context.Context, sagaID string) (*saga.Saga, error) {
	sagaInstance, ok := s.orchestrator.GetSaga(sagaID)
	if !ok {
		return nil, fmt.Errorf("saga not found: %s", sagaID)
	}
	return sagaInstance, nil
}

// RetrySaga attempts to resume a failed saga
func (s *OrderService) RetrySaga(ctx context.Context, sagaID string) error {
	return s.orchestrator.Retry(ctx, sagaID)
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
	DiscountPaisa  int64
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

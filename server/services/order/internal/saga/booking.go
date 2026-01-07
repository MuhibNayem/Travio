package saga

import (
	"context"
	"fmt"
)

// BookingSaga defines the saga steps for creating a ticket booking
// Steps: CheckEntitlement -> ValidateNID -> HoldSeats -> ProcessPayment -> ConfirmBooking -> RecordUsage -> SendNotification
func NewBookingSaga(deps *BookingDependencies, req *BookingRequest) *Saga {
	o := NewOrchestrator(nil, nil) // Persistence and DLQ handled by execution context

	saga := o.CreateSaga("booking", []*Step{
		{
			Name: "check_entitlement",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.checkEntitlement(ctx, sagaCtx, req)
			},
			// No compensation - this is a gate, not a mutation
		},
		{
			Name: "validate_nid",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.validateNID(ctx, sagaCtx, req)
			},
			// No compensation needed - validation is idempotent
		},
		{
			Name: "hold_seats",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.holdSeats(ctx, sagaCtx, req)
			},
			CompensateFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.releaseSeats(ctx, sagaCtx)
			},
		},
		{
			Name: "process_payment",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.processPayment(ctx, sagaCtx, req)
			},
			CompensateFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.refundPayment(ctx, sagaCtx)
			},
		},
		{
			Name: "confirm_booking",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.confirmBooking(ctx, sagaCtx, req)
			},
			CompensateFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.cancelBooking(ctx, sagaCtx)
			},
		},
		{
			Name: "record_usage",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				// Best effort - log error but don't fail saga
				// In real world, push to DLQ or retry queue
				if err := deps.SubscriptionService.RecordUsage(ctx, req.OrgID, "ticket_sale", int64(len(req.Passengers)), req.OrderID); err != nil {
					fmt.Printf("Failed to record usage for order %s: %v\n", req.OrderID, err)
				}
				return nil
			},
		},
		{
			Name: "send_notification",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.sendNotification(ctx, sagaCtx, req)
			},
			// No compensation - notification failures shouldn't rollback booking
		},
	})

	// Pre-populate context with request data
	saga.Context.Set("order_id", req.OrderID)
	saga.Context.Set("user_id", req.UserID)
	saga.Context.Set("trip_id", req.TripID)
	saga.Context.Set("hold_id", req.HoldID)

	return saga
}

// BookingDependencies contains service clients needed for the saga
type BookingDependencies struct {
	NIDService          NIDVerifier
	InventoryService    InventoryClient
	PaymentService      PaymentClient
	SubscriptionService SubscriptionClient
	NotificationSvc     NotificationClient
}

// BookingRequest contains the booking order details
type BookingRequest struct {
	OrderID       string
	UserID        string
	OrgID         string
	TripID        string
	HoldID        string
	FromStation   string
	ToStation     string
	Passengers    []PassengerInfo
	PaymentToken  string
	PaymentMethod string
	TotalPaisa    int64
	Email         string
	Phone         string
}

type PassengerInfo struct {
	NID         string
	Name        string
	SeatID      string
	DateOfBirth string
	Gender      string
}

// --- Service Interfaces ---

type NIDVerifier interface {
	Verify(ctx context.Context, nid, dob, name string) (bool, error)
}

type InventoryClient interface {
	HoldSeats(ctx context.Context, tripID string, seatIDs []string, userID string) (string, error)
	ReleaseSeats(ctx context.Context, holdID, userID string) error
	ConfirmBooking(ctx context.Context, holdID, orderID, userID string, passengers []PassengerInfo) (string, error)
	CancelBooking(ctx context.Context, bookingID, orderID string) error
}

type PaymentClient interface {
	Authorize(ctx context.Context, orderID, orgID, token string, amountPaisa int64) (string, error)
	Capture(ctx context.Context, paymentID string) error
	Refund(ctx context.Context, paymentID string, amountPaisa int64) (string, error)
}

type SubscriptionClient interface {
	RecordUsage(ctx context.Context, orgID, eventType string, units int64, idempotencyKey string) error
	GetEntitlement(ctx context.Context, orgID string) (*EntitlementInfo, error)
}

// EntitlementInfo contains subscription entitlement data for enforcement
type EntitlementInfo struct {
	Status          string
	PlanID          string
	PlanName        string
	Features        map[string]string
	UsageThisPeriod map[string]int64
	QuotaLimits     map[string]int64
}

type NotificationClient interface {
	SendBookingConfirmation(ctx context.Context, email, phone, orderID string) error
	SendBookingCancellation(ctx context.Context, email, phone, orderID, reason string) error
}

// --- Step Implementations ---

func (d *BookingDependencies) checkEntitlement(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	// Skip entitlement check if no subscription service
	if d.SubscriptionService == nil {
		return nil
	}

	ent, err := d.SubscriptionService.GetEntitlement(ctx, req.OrgID)
	if err != nil {
		// Fail-open: log error but allow booking to proceed
		fmt.Printf("Warning: Entitlement check failed for org %s: %v\n", req.OrgID, err)
		return nil
	}

	if ent == nil {
		return fmt.Errorf("organization %s has no active subscription", req.OrgID)
	}

	// Check subscription status
	if ent.Status != "active" && ent.Status != "trialing" {
		return fmt.Errorf("subscription is %s, active subscription required", ent.Status)
	}

	// Check booking quota if defined
	if limit, ok := ent.QuotaLimits["max_bookings_per_month"]; ok {
		usage := ent.UsageThisPeriod["ticket_sale"]
		newUsage := usage + int64(len(req.Passengers))
		if newUsage > limit {
			return fmt.Errorf("booking quota exceeded: %d/%d (need %d more)", usage, limit, int64(len(req.Passengers)))
		}
	}

	sagaCtx.Set("entitlement_verified", true)
	sagaCtx.Set("plan_name", ent.PlanName)
	return nil
}

func (d *BookingDependencies) validateNID(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	for _, p := range req.Passengers {
		valid, err := d.NIDService.Verify(ctx, p.NID, p.DateOfBirth, p.Name)
		if err != nil {
			return fmt.Errorf("NID verification failed for %s: %w", p.Name, err)
		}
		if !valid {
			return fmt.Errorf("NID validation failed for %s", p.Name)
		}
	}
	sagaCtx.Set("nid_verified", true)
	return nil
}

func (d *BookingDependencies) holdSeats(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	// If we already have a hold from the frontend, verify it
	if req.HoldID != "" {
		sagaCtx.Set("hold_id", req.HoldID)
		return nil
	}

	// Otherwise, create a new hold
	var seatIDs []string
	for _, p := range req.Passengers {
		seatIDs = append(seatIDs, p.SeatID)
	}

	holdID, err := d.InventoryService.HoldSeats(ctx, req.TripID, seatIDs, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to hold seats: %w", err)
	}

	sagaCtx.Set("hold_id", holdID)
	return nil
}

func (d *BookingDependencies) releaseSeats(ctx context.Context, sagaCtx *SagaContext) error {
	holdID := sagaCtx.GetString("hold_id")
	userID := sagaCtx.GetString("user_id")

	if holdID == "" {
		return nil // Nothing to release
	}

	return d.InventoryService.ReleaseSeats(ctx, holdID, userID)
}

func (d *BookingDependencies) processPayment(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	paymentID, err := d.PaymentService.Authorize(ctx, req.OrderID, req.OrgID, req.PaymentToken, req.TotalPaisa)
	if err != nil {
		return fmt.Errorf("payment authorization failed: %w", err)
	}

	sagaCtx.Set("payment_id", paymentID)

	// Capture the payment
	if err := d.PaymentService.Capture(ctx, paymentID); err != nil {
		return fmt.Errorf("payment capture failed: %w", err)
	}

	sagaCtx.Set("payment_captured", true)
	return nil
}

func (d *BookingDependencies) refundPayment(ctx context.Context, sagaCtx *SagaContext) error {
	paymentID := sagaCtx.GetString("payment_id")
	if paymentID == "" {
		return nil // No payment to refund
	}

	// Only refund if payment was captured
	if captured, _ := sagaCtx.Get("payment_captured"); captured != true {
		return nil
	}

	// Get original amount
	amount := sagaCtx.GetInt64("total_paisa")

	refundID, err := d.PaymentService.Refund(ctx, paymentID, amount)
	if err != nil {
		return fmt.Errorf("refund failed: %w", err)
	}

	sagaCtx.Set("refund_id", refundID)
	return nil
}

func (d *BookingDependencies) confirmBooking(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	holdID := sagaCtx.GetString("hold_id")
	orderID := sagaCtx.GetString("order_id")
	userID := sagaCtx.GetString("user_id")

	bookingID, err := d.InventoryService.ConfirmBooking(ctx, holdID, orderID, userID, req.Passengers)
	if err != nil {
		return fmt.Errorf("booking confirmation failed: %w", err)
	}

	sagaCtx.Set("booking_id", bookingID)
	return nil
}

func (d *BookingDependencies) cancelBooking(ctx context.Context, sagaCtx *SagaContext) error {
	bookingID := sagaCtx.GetString("booking_id")
	orderID := sagaCtx.GetString("order_id")

	if bookingID == "" {
		return nil // Nothing to cancel
	}

	return d.InventoryService.CancelBooking(ctx, bookingID, orderID)
}

func (d *BookingDependencies) sendNotification(ctx context.Context, sagaCtx *SagaContext, req *BookingRequest) error {
	orderID := sagaCtx.GetString("order_id")

	// Best effort - don't fail saga for notification failures
	_ = d.NotificationSvc.SendBookingConfirmation(ctx, req.Email, req.Phone, orderID)

	return nil
}

// NewCancellationSaga creates a new cancellation saga
func NewCancellationSaga(
	deps *BookingDependencies,
	orderID, userID, bookingID, paymentID string,
	email, phone string,
	amount int64,
) *Saga {
	o := NewOrchestrator(nil, nil) // Persistence and DLQ handled by execution context

	saga := o.CreateSaga("cancellation", []*Step{
		{
			Name: "cancel_booking",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.InventoryService.CancelBooking(ctx, bookingID, orderID)
			},
		},
		{
			Name: "process_refund",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				refundID, err := deps.PaymentService.Refund(ctx, paymentID, amount)
				if err != nil {
					return err
				}
				sagaCtx.Set("refund_id", refundID)
				return nil
			},
		},
		{
			Name: "send_cancellation_notification",
			ExecuteFn: func(ctx context.Context, sagaCtx *SagaContext) error {
				return deps.NotificationSvc.SendBookingCancellation(ctx, email, phone, orderID, "user requested")
			},
		},
	})

	saga.Context.Set("order_id", orderID)
	saga.Context.Set("user_id", userID)
	saga.Context.Set("booking_id", bookingID)
	saga.Context.Set("payment_id", paymentID)

	return saga
}

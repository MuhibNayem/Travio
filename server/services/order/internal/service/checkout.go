package service

import (
	"context"
	"fmt"
	"time"

	paymentpb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	pricingpb "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CheckoutTTL           = 15 * time.Minute
	DefaultTaxRate        = 0.05 // 5%
	DefaultBookingFeePaisa = 2000 // 20 BDT per passenger
	DefaultCurrency       = "BDT"
)

// CheckoutService handles checkout session management
type CheckoutService struct {
	checkoutRepo    *repository.CheckoutRepository
	pricingClient   pricingpb.PricingServiceClient
	paymentClient   paymentpb.PaymentServiceClient
	crmClient       CouponValidator
}

// CouponValidator interface for coupon validation
type CouponValidator interface {
	ValidateCoupon(ctx context.Context, orgID, code string, cartTotal int64) (*domain.CouponValidation, error)
}

// NewCheckoutService creates a new checkout service
func NewCheckoutService(
	checkoutRepo *repository.CheckoutRepository,
	pricingClient pricingpb.PricingServiceClient,
	paymentClient paymentpb.PaymentServiceClient,
	crmClient CRMClient,
) *CheckoutService {
	return &CheckoutService{
		checkoutRepo:  checkoutRepo,
		pricingClient: pricingClient,
		paymentClient: paymentClient,
		crmClient:     crmClient,
	}
}

// CreateCheckoutRequest represents a request to create a checkout session
type CreateCheckoutRequest struct {
	OrganizationID string
	UserID         string
	TripID         string
	HoldID         string
	FromStationID  string
	ToStationID    string
	Passengers     []CheckoutPassengerInput
	CouponCode     string
}

// CheckoutPassengerInput represents passenger input data
type CheckoutPassengerInput struct {
	NID                   string
	Name                  string
	SeatID                string
	SeatNumber            string
	SeatClass             string
	SeatType              string
	DateOfBirth           string
	Gender                string
	Age                   int
	Phone                 string
	Email                 string
	EmergencyContactName  string
	EmergencyContactPhone string
}

// CreateCheckout creates a new checkout session from a seat hold
func (s *CheckoutService) CreateCheckout(ctx context.Context, req *CreateCheckoutRequest) (*domain.CheckoutSession, error) {
	// Validate seat count matches passenger count
	if len(req.Passengers) == 0 {
		return nil, fmt.Errorf("at least one passenger is required")
	}

	// Create checkout session
	session := &domain.CheckoutSession{
		OrganizationID: req.OrganizationID,
		UserID:         req.UserID,
		TripID:         req.TripID,
		HoldID:         req.HoldID,
		FromStationID:  req.FromStationID,
		ToStationID:    req.ToStationID,
		Status:         domain.CheckoutStatusPending,
		Currency:       DefaultCurrency,
		ExpiresAt:      time.Now().Add(CheckoutTTL),
	}

	// Map passengers
	for _, p := range req.Passengers {
		session.Passengers = append(session.Passengers, domain.CheckoutPassenger{
			NID:                   p.NID,
			Name:                  p.Name,
			SeatID:                p.SeatID,
			SeatNumber:            p.SeatNumber,
			SeatClass:             p.SeatClass,
			SeatType:              p.SeatType,
			DateOfBirth:           p.DateOfBirth,
			Gender:                p.Gender,
			Age:                   p.Age,
			Phone:                 p.Phone,
			Email:                 p.Email,
			EmergencyContactName:  p.EmergencyContactName,
			EmergencyContactPhone: p.EmergencyContactPhone,
		})
	}

	// Calculate pricing
	pricing, err := s.calculatePricing(ctx, req, session.Passengers)
	if err != nil {
		logger.Error("Failed to calculate pricing", "error", err, "trip_id", req.TripID)
		return nil, fmt.Errorf("failed to calculate pricing: %w", err)
	}

	session.BasePricePaisa = pricing.BasePricePaisa
	session.TaxPaisa = pricing.TaxPaisa
	session.BookingFeePaisa = pricing.BookingFeePaisa
	session.DiscountPaisa = pricing.DiscountPaisa
	session.TotalPaisa = pricing.TotalPaisa

	// Apply coupon if provided
	if req.CouponCode != "" {
		couponValidation, err := s.validateCoupon(ctx, req.OrganizationID, req.CouponCode, session.TotalPaisa)
		if err != nil {
			logger.Warn("Coupon validation failed", "error", err, "code", req.CouponCode)
			// Don't fail checkout for invalid coupon, just log it
		} else if couponValidation.Valid {
			session.CouponCode = req.CouponCode
			session.CouponDiscount = couponValidation.DiscountPaisa
			session.CouponValidated = true
			session.DiscountPaisa += couponValidation.DiscountPaisa
			session.TotalPaisa -= couponValidation.DiscountPaisa
			if session.TotalPaisa < 0 {
				session.TotalPaisa = 0
			}
		}
	}

	// Persist session
	if err := s.checkoutRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	logger.Info("Checkout session created",
		"session_id", session.ID,
		"hold_id", session.HoldID,
		"total_paisa", session.TotalPaisa,
		"passenger_count", len(session.Passengers),
	)

	return session, nil
}

// GetCheckout retrieves a checkout session by ID
func (s *CheckoutService) GetCheckout(ctx context.Context, sessionID, userID string) (*domain.CheckoutSession, error) {
	session, err := s.checkoutRepo.GetByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	// Check if session has expired
	if session.Status == domain.CheckoutStatusPending && time.Now().After(session.ExpiresAt) {
		_ = s.checkoutRepo.UpdateStatus(ctx, sessionID, domain.CheckoutStatusExpired)
		session.Status = domain.CheckoutStatusExpired
		return session, fmt.Errorf("checkout session has expired")
	}

	return session, nil
}

// GetCheckoutByHoldID retrieves a checkout session by hold ID
func (s *CheckoutService) GetCheckoutByHoldID(ctx context.Context, holdID, userID string) (*domain.CheckoutSession, error) {
	return s.checkoutRepo.GetByHoldID(ctx, holdID, userID)
}

// UpdateCheckout updates passenger details and payment method
func (s *CheckoutService) UpdateCheckout(ctx context.Context, sessionID, userID string, req *UpdateCheckoutRequest) (*domain.CheckoutSession, error) {
	session, err := s.checkoutRepo.GetByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	// Check if session is still pending
	if session.Status != domain.CheckoutStatusPending {
		return nil, fmt.Errorf("cannot update checkout session in status: %s", session.Status)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		_ = s.checkoutRepo.UpdateStatus(ctx, sessionID, domain.CheckoutStatusExpired)
		return nil, fmt.Errorf("checkout session has expired")
	}

	// Update passengers if provided
	if len(req.Passengers) > 0 {
		// Delete existing passengers
		if err := s.checkoutRepo.DeletePassengers(ctx, sessionID); err != nil {
			return nil, fmt.Errorf("failed to delete existing passengers: %w", err)
		}

		// Add new passengers
		for _, p := range req.Passengers {
			session.Passengers = append(session.Passengers, domain.CheckoutPassenger{
				NID:                   p.NID,
				Name:                  p.Name,
				SeatID:                p.SeatID,
				SeatNumber:            p.SeatNumber,
				SeatClass:             p.SeatClass,
				SeatType:              p.SeatType,
				DateOfBirth:           p.DateOfBirth,
				Gender:                p.Gender,
				Age:                   p.Age,
				Phone:                 p.Phone,
				Email:                 p.Email,
				EmergencyContactName:  p.EmergencyContactName,
				EmergencyContactPhone: p.EmergencyContactPhone,
			})
		}
	}

	// Update payment method
	if req.PaymentMethod != "" {
		session.PaymentMethod = req.PaymentMethod
	}
	if req.PaymentGateway != "" {
		session.PaymentGateway = req.PaymentGateway
	}

	// Update coupon if provided
	if req.CouponCode != "" && req.CouponCode != session.CouponCode {
		couponValidation, err := s.validateCoupon(ctx, session.OrganizationID, req.CouponCode, session.TotalPaisa)
		if err != nil {
			return nil, fmt.Errorf("coupon validation failed: %w", err)
		}
		if couponValidation.Valid {
			// Reset previous discount
			session.TotalPaisa += session.DiscountPaisa
			session.DiscountPaisa = 0

			// Apply new discount
			session.CouponCode = req.CouponCode
			session.CouponDiscount = couponValidation.DiscountPaisa
			session.CouponValidated = true
			session.DiscountPaisa = couponValidation.DiscountPaisa
			session.TotalPaisa -= couponValidation.DiscountPaisa
			if session.TotalPaisa < 0 {
				session.TotalPaisa = 0
			}
		}
	}

	session.Status = domain.CheckoutStatusInProgress
	session.UpdatedAt = time.Now()

	if err := s.checkoutRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update checkout session: %w", err)
	}

	return session, nil
}

// ConfirmCheckout confirms a checkout session and returns the order details
func (s *CheckoutService) ConfirmCheckout(ctx context.Context, sessionID, userID string) (*domain.CheckoutSession, error) {
	session, err := s.checkoutRepo.GetByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	// Validate session
	if session.Status != domain.CheckoutStatusInProgress && session.Status != domain.CheckoutStatusPending {
		return nil, fmt.Errorf("cannot confirm checkout session in status: %s", session.Status)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.checkoutRepo.UpdateStatus(ctx, sessionID, domain.CheckoutStatusExpired)
		return nil, fmt.Errorf("checkout session has expired")
	}

	// Validate all passengers have required info
	for _, p := range session.Passengers {
		if p.NID == "" || p.Name == "" {
			return nil, fmt.Errorf("passenger NID and name are required for all passengers")
		}
	}

	// Validate payment method is set
	if session.PaymentMethod == "" {
		return nil, fmt.Errorf("payment method must be set before confirming checkout")
	}

	// Mark as confirmed
	session.Status = domain.CheckoutStatusConfirmed
	now := time.Now()
	session.CompletedAt = &now
	session.UpdatedAt = now

	if err := s.checkoutRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to confirm checkout session: %w", err)
	}

	logger.Info("Checkout session confirmed",
		"session_id", session.ID,
		"hold_id", session.HoldID,
		"total_paisa", session.TotalPaisa,
	)

	return session, nil
}

// ListCheckouts retrieves checkout sessions for a user
func (s *CheckoutService) ListCheckouts(ctx context.Context, userID, status string, limit, offset int) ([]*domain.CheckoutSession, int64, error) {
	return s.checkoutRepo.ListByUser(ctx, userID, status, limit, offset)
}

// ExpireSessions runs a cleanup job to expire old checkout sessions
func (s *CheckoutService) ExpireSessions(ctx context.Context) error {
	return s.checkoutRepo.ExpireOldSessions(ctx)
}

// --- Internal Methods ---

func (s *CheckoutService) calculatePricing(
	ctx context.Context,
	req *CreateCheckoutRequest,
	passengers []domain.CheckoutPassenger,
) (*domain.PricingBreakdown, error) {
	// Calculate base price from passengers
	var basePrice int64
	for i := 0; i < len(passengers); i++ {
		// In production, fetch from pricing service or catalog
		// For now, use a placeholder or get from passenger seat info
		basePrice += 50000 // 500 BDT default per seat
	}

	// Try dynamic pricing from Pricing Service
	if s.pricingClient != nil {
		pricingResp, err := s.pricingClient.CalculatePrice(ctx, &pricingpb.CalculatePriceRequest{
			TripId:         req.TripID,
			SeatClass:      passengers[0].SeatClass,
			Date:           time.Now().Format("2006-01-02"),
			Quantity:       int32(len(passengers)),
			BasePricePaisa: basePrice,
			OccupancyRate:  0.5, // Default 50% occupancy
		})
		if err != nil {
			logger.Warn("Failed to get dynamic pricing, using base price", "error", err)
		} else {
			basePrice = pricingResp.FinalPricePaisa
		}
	}

	bookingFee := int64(DefaultBookingFeePaisa) * int64(len(passengers))
	tax := int64(float64(basePrice) * DefaultTaxRate)
	total := basePrice + tax + bookingFee

	return &domain.PricingBreakdown{
		BasePricePaisa:  basePrice,
		TaxPaisa:        tax,
		BookingFeePaisa: bookingFee,
		DiscountPaisa:   0,
		TotalPaisa:      total,
		Currency:        DefaultCurrency,
		CouponApplied:   false,
	}, nil
}

func (s *CheckoutService) validateCoupon(ctx context.Context, orgID, code string, cartTotal int64) (*domain.CouponValidation, error) {
	if s.crmClient == nil {
		return nil, fmt.Errorf("CRM service not configured")
	}
	return s.crmClient.ValidateCoupon(ctx, orgID, code, cartTotal)
}

// UpdateCheckoutRequest represents a request to update a checkout session
type UpdateCheckoutRequest struct {
	Passengers     []CheckoutPassengerInput
	PaymentMethod  string
	PaymentGateway string
	CouponCode     string
}

// NewCheckoutServiceWithGRPC creates a checkout service with gRPC clients
func NewCheckoutServiceWithGRPC(
	checkoutRepo *repository.CheckoutRepository,
	pricingAddr string,
	paymentAddr string,
	crmClient CRMClient,
) (*CheckoutService, error) {
	var pricingClient pricingpb.PricingServiceClient
	var paymentClient paymentpb.PaymentServiceClient

	if pricingAddr != "" {
		conn, err := grpc.Dial(pricingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to pricing service: %w", err)
		}
		pricingClient = pricingpb.NewPricingServiceClient(conn)
	}

	if paymentAddr != "" {
		conn, err := grpc.Dial(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to payment service: %w", err)
		}
		paymentClient = paymentpb.NewPaymentServiceClient(conn)
	}

	return NewCheckoutService(checkoutRepo, pricingClient, paymentClient, crmClient), nil
}

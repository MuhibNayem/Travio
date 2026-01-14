package clients

import (
	"context"
	"time"

	catalogpb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	inventorypb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	nidpb "github.com/MuhibNayem/Travio/server/api/proto/nid/v1"
	paymentpb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	pricingpb "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	subscriptionpb "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NIDClient implements saga.NIDVerifier via gRPC
type NIDClient struct {
	client nidpb.NIDServiceClient
}

func NewNIDClient(addr string) (*NIDClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &NIDClient{client: nidpb.NewNIDServiceClient(conn)}, nil
}

func (c *NIDClient) Verify(ctx context.Context, nid, dob, name string) (bool, error) {
	resp, err := c.client.Verify(ctx, &nidpb.VerifyNIDRequest{
		Nid:         nid,
		Country:     "BD",
		DateOfBirth: dob,
		Name:        name,
	})
	if err != nil {
		return false, err
	}
	return resp.IsValid, nil
}

// InventoryClient implements saga.InventoryClient via gRPC
type InventoryClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &InventoryClient{client: inventorypb.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryClient) HoldSeats(ctx context.Context, orgID, tripID string, seatIDs []string, userID string) (string, error) {
	resp, err := c.client.HoldSeats(ctx, &inventorypb.HoldSeatsRequest{
		OrganizationId:      orgID,
		TripId:              tripID,
		SeatIds:             seatIDs,
		UserId:              userID,
		HoldDurationSeconds: 600,
	})
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", &saga.SagaError{Message: resp.FailureReason}
	}
	return resp.HoldId, nil
}

func (c *InventoryClient) ReleaseSeats(ctx context.Context, orgID, holdID, userID string) error {
	_, err := c.client.ReleaseSeats(ctx, &inventorypb.ReleaseSeatsRequest{
		OrganizationId: orgID,
		HoldId:         holdID,
		UserId:         userID,
	})
	return err
}

func (c *InventoryClient) ConfirmBooking(ctx context.Context, orgID, holdID, orderID, userID string, passengers []saga.PassengerInfo) (string, error) {
	var pbPassengers []*inventorypb.PassengerSeat
	for _, p := range passengers {
		pbPassengers = append(pbPassengers, &inventorypb.PassengerSeat{
			SeatId:        p.SeatID,
			PassengerNid:  p.NID,
			PassengerName: p.Name,
		})
	}

	resp, err := c.client.ConfirmBooking(ctx, &inventorypb.ConfirmBookingRequest{
		OrganizationId: orgID,
		HoldId:         holdID,
		OrderId:        orderID,
		UserId:         userID,
		Passengers:     pbPassengers,
	})
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", &saga.SagaError{Message: resp.FailureReason}
	}
	return resp.BookingId, nil
}

func (c *InventoryClient) CancelBooking(ctx context.Context, bookingID, orderID string) error {
	// Inventory service would need a CancelBooking RPC - for now use release
	return nil
}

func (c *InventoryClient) GetSeatMap(ctx context.Context, orgID, tripID, fromStationID, toStationID string) (*inventorypb.GetSeatMapResponse, error) {
	return c.client.GetSeatMap(ctx, &inventorypb.GetSeatMapRequest{
		OrganizationId: orgID,
		TripId:         tripID,
		FromStationId:  fromStationID,
		ToStationId:    toStationID,
	})
}

// PaymentClient implements saga.PaymentClient via gRPC
type PaymentClient struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PaymentClient{client: paymentpb.NewPaymentServiceClient(conn)}, nil
}

func (c *PaymentClient) Authorize(ctx context.Context, orderID, orgID, token string, amountPaisa int64) (string, error) {
	resp, err := c.client.CreatePayment(ctx, &paymentpb.CreatePaymentRequest{
		OrderId:        orderID,
		OrganizationId: orgID,
		AmountPaisa:    amountPaisa,
		Currency:       "BDT",
		PaymentMethod:  "sslcommerz",
	})
	if err != nil {
		return "", err
	}
	return resp.PaymentId, nil
}

func (c *PaymentClient) Capture(ctx context.Context, paymentID string) error {
	_, err := c.client.CapturePayment(ctx, &paymentpb.CapturePaymentRequest{
		Gateway:       "sslcommerz",
		TransactionId: paymentID,
	})
	return err
}

func (c *PaymentClient) Refund(ctx context.Context, paymentID string, amountPaisa int64) (string, error) {
	resp, err := c.client.RefundPayment(ctx, &paymentpb.RefundPaymentRequest{
		Gateway:       "sslcommerz",
		TransactionId: paymentID,
		AmountPaisa:   amountPaisa,
		Reason:        "order_cancelled",
	})
	if err != nil {
		return "", err
	}
	return resp.RefundId, nil
}

// SubscriptionClient implements saga.SubscriptionClient via gRPC
type SubscriptionClient struct {
	client subscriptionpb.SubscriptionServiceClient
}

func NewSubscriptionClient(addr string) (*SubscriptionClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &SubscriptionClient{client: subscriptionpb.NewSubscriptionServiceClient(conn)}, nil
}

func (c *SubscriptionClient) RecordUsage(ctx context.Context, orgID, eventType string, units int64, idempotencyKey string) error {
	_, err := c.client.RecordUsage(ctx, &subscriptionpb.RecordUsageRequest{
		OrganizationId: orgID,
		EventType:      eventType,
		Units:          units,
		IdempotencyKey: idempotencyKey,
	})
	return err
}

func (c *SubscriptionClient) GetEntitlement(ctx context.Context, orgID string) (*saga.EntitlementInfo, error) {
	resp, err := c.client.GetEntitlement(ctx, &subscriptionpb.GetEntitlementRequest{
		OrganizationId: orgID,
	})
	if err != nil {
		return nil, err
	}

	return &saga.EntitlementInfo{
		Status:          resp.Status,
		PlanID:          resp.PlanId,
		PlanName:        resp.PlanName,
		Features:        resp.Features,
		UsageThisPeriod: resp.UsageThisPeriod,
		QuotaLimits:     resp.QuotaLimits,
	}, nil
}

// CatalogClient fetches trip information for pricing.
type CatalogClient struct {
	client catalogpb.CatalogServiceClient
}

func NewCatalogClient(addr string) (*CatalogClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CatalogClient{client: catalogpb.NewCatalogServiceClient(conn)}, nil
}

func (c *CatalogClient) GetTrip(ctx context.Context, orgID, tripID string) (*catalogpb.Trip, error) {
	return c.client.GetTrip(ctx, &catalogpb.GetTripRequest{
		Id:             tripID,
		OrganizationId: orgID,
	})
}

// PricingClient calculates dynamic pricing rules.
type PricingClient struct {
	client pricingpb.PricingServiceClient
}

func NewPricingClient(addr string) (*PricingClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PricingClient{client: pricingpb.NewPricingServiceClient(conn)}, nil
}

func (c *PricingClient) CalculatePrice(ctx context.Context, req *pricingpb.CalculatePriceRequest) (*pricingpb.CalculatePriceResponse, error) {
	return c.client.CalculatePrice(ctx, req)
}

// NotificationClient implements saga.NotificationClient using structured logging.
// Once notification.proto is defined, this should be replaced with gRPC client.
type NotificationClient struct{}

func NewNotificationClient() *NotificationClient {
	return &NotificationClient{}
}

func (c *NotificationClient) SendBookingConfirmation(ctx context.Context, email, phone, orderID string) error {
	logger.Info("Sending booking confirmation",
		"order_id", orderID,
		"email", email,
		"phone", phone,
	)
	return nil
}

func (c *NotificationClient) SendBookingCancellation(ctx context.Context, email, phone, orderID, reason string) error {
	logger.Info("Sending booking cancellation",
		"order_id", orderID,
		"email", email,
		"phone", phone,
		"reason", reason,
	)
	return nil
}

// Helper for timeouts
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 30*time.Second)
}

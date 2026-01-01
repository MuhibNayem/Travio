package clients

import (
	"context"
	"time"

	inventorypb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	nidpb "github.com/MuhibNayem/Travio/server/api/proto/nid/v1"
	paymentpb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
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

func (c *InventoryClient) HoldSeats(ctx context.Context, tripID string, seatIDs []string, userID string) (string, error) {
	resp, err := c.client.HoldSeats(ctx, &inventorypb.HoldSeatsRequest{
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

func (c *InventoryClient) ReleaseSeats(ctx context.Context, holdID, userID string) error {
	_, err := c.client.ReleaseSeats(ctx, &inventorypb.ReleaseSeatsRequest{
		HoldId: holdID,
		UserId: userID,
	})
	return err
}

func (c *InventoryClient) ConfirmBooking(ctx context.Context, holdID, orderID, userID string, passengers []saga.PassengerInfo) (string, error) {
	var pbPassengers []*inventorypb.PassengerSeat
	for _, p := range passengers {
		pbPassengers = append(pbPassengers, &inventorypb.PassengerSeat{
			SeatId:        p.SeatID,
			PassengerNid:  p.NID,
			PassengerName: p.Name,
		})
	}

	resp, err := c.client.ConfirmBooking(ctx, &inventorypb.ConfirmBookingRequest{
		HoldId:     holdID,
		OrderId:    orderID,
		UserId:     userID,
		Passengers: pbPassengers,
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

func (c *PaymentClient) Authorize(ctx context.Context, orderID, token string, amountPaisa int64) (string, error) {
	resp, err := c.client.CreatePayment(ctx, &paymentpb.CreatePaymentRequest{
		OrderId:       orderID,
		AmountPaisa:   amountPaisa,
		Currency:      "BDT",
		PaymentMethod: "sslcommerz",
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

// NotificationClient implements saga.NotificationClient
type NotificationClient struct {
	// In production, this would be a gRPC client to notification service
}

func NewNotificationClient() *NotificationClient {
	return &NotificationClient{}
}

func (c *NotificationClient) SendBookingConfirmation(ctx context.Context, email, phone, orderID string) error {
	// TODO: Implement via notification service gRPC
	return nil
}

func (c *NotificationClient) SendBookingCancellation(ctx context.Context, email, phone, orderID, reason string) error {
	// TODO: Implement via notification service gRPC
	return nil
}

// Helper for timeouts
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 30*time.Second)
}

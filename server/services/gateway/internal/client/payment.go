package client

import (
	"context"
	"time"

	paymentv1 "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// PaymentClient wraps the gRPC payment client
type PaymentClient struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

// NewPaymentClient creates a new gRPC client for the payment service
// Uses mTLS if TLS config is provided
func NewPaymentClient(address string, tlsCfg TLSConfig) (*PaymentClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to payment service", "address", address, "tls", tlsCfg.CertFile != "")
	return &PaymentClient{
		conn:   conn,
		client: paymentv1.NewPaymentServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *PaymentClient) Close() error {
	return c.conn.Close()
}

// CreatePayment creates a new payment
func (c *PaymentClient) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	return c.client.CreatePayment(ctx, req)
}

// VerifyPayment verifies a payment status
func (c *PaymentClient) VerifyPayment(ctx context.Context, req *paymentv1.VerifyPaymentRequest) (*paymentv1.PaymentStatusResponse, error) {
	return c.client.VerifyPayment(ctx, req)
}

// RefundPayment processes a refund
func (c *PaymentClient) RefundPayment(ctx context.Context, req *paymentv1.RefundPaymentRequest) (*paymentv1.RefundResponse, error) {
	return c.client.RefundPayment(ctx, req)
}

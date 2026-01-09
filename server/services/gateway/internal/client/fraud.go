package client

import (
	"context"
	"time"

	fraudv1 "github.com/MuhibNayem/Travio/server/api/proto/fraud/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// FraudClient wraps the gRPC fraud client
type FraudClient struct {
	conn   *grpc.ClientConn
	client fraudv1.FraudServiceClient
}

// NewFraudClient creates a new gRPC client for the fraud service
func NewFraudClient(address string, tlsCfg TLSConfig) (*FraudClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to fraud service", "address", address, "tls", tlsCfg.CertFile != "")
	return &FraudClient{
		conn:   conn,
		client: fraudv1.NewFraudServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *FraudClient) Close() error {
	return c.conn.Close()
}

// AnalyzeBooking performs fraud analysis on a booking
func (c *FraudClient) AnalyzeBooking(ctx context.Context, req *fraudv1.AnalyzeBookingRequest) (*fraudv1.AnalyzeBookingResponse, error) {
	return c.client.AnalyzeBooking(ctx, req)
}

// VerifyDocument performs document verification
func (c *FraudClient) VerifyDocument(ctx context.Context, req *fraudv1.VerifyDocumentRequest) (*fraudv1.VerifyDocumentResponse, error) {
	return c.client.VerifyDocument(ctx, req)
}

// Health checks the fraud service health
func (c *FraudClient) Health(ctx context.Context) (*fraudv1.HealthResponse, error) {
	return c.client.Health(ctx, &fraudv1.HealthRequest{})
}

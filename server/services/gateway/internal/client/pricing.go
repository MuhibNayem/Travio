package client

import (
	"context"
	"time"

	pricingv1 "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// PricingClient wraps the gRPC pricing client
type PricingClient struct {
	conn   *grpc.ClientConn
	client pricingv1.PricingServiceClient
}

// NewPricingClient creates a new gRPC client for the pricing service
// Uses mTLS if TLS config is provided
func NewPricingClient(address string, tlsCfg TLSConfig) (*PricingClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to pricing service", "address", address, "tls", tlsCfg.CertFile != "")
	return &PricingClient{
		conn:   conn,
		client: pricingv1.NewPricingServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *PricingClient) Close() error {
	return c.conn.Close()
}

// CalculatePrice calculates dynamic price via gRPC
func (c *PricingClient) CalculatePrice(ctx context.Context, req *pricingv1.CalculatePriceRequest) (*pricingv1.CalculatePriceResponse, error) {
	return c.client.CalculatePrice(ctx, req)
}

// GetRules retrieves all pricing rules via gRPC
func (c *PricingClient) GetRules(ctx context.Context, includeInactive bool) (*pricingv1.GetRulesResponse, error) {
	return c.client.GetRules(ctx, &pricingv1.GetRulesRequest{IncludeInactive: includeInactive})
}

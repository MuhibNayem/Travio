package client

import (
	"context"
	"time"

	operatorv1 "github.com/MuhibNayem/Travio/server/api/proto/operator/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// OperatorClient wraps the gRPC operator client
type OperatorClient struct {
	conn   *grpc.ClientConn
	client operatorv1.VendorServiceClient
}

// NewOperatorClient creates a new gRPC client for the operator service
// Uses mTLS if TLS config is provided
func NewOperatorClient(address string, tlsCfg TLSConfig) (*OperatorClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to operator service", "address", address, "tls", tlsCfg.CertFile != "")
	return &OperatorClient{
		conn:   conn,
		client: operatorv1.NewVendorServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *OperatorClient) Close() error {
	return c.conn.Close()
}

// CreateVendor creates a new vendor
func (c *OperatorClient) CreateVendor(ctx context.Context, req *operatorv1.CreateVendorRequest) (*operatorv1.CreateVendorResponse, error) {
	return c.client.CreateVendor(ctx, req)
}

// GetVendor retrieves a vendor by ID
func (c *OperatorClient) GetVendor(ctx context.Context, id string) (*operatorv1.GetVendorResponse, error) {
	return c.client.GetVendor(ctx, &operatorv1.GetVendorRequest{Id: id})
}

// UpdateVendor updates an existing vendor
func (c *OperatorClient) UpdateVendor(ctx context.Context, req *operatorv1.UpdateVendorRequest) (*operatorv1.UpdateVendorResponse, error) {
	return c.client.UpdateVendor(ctx, req)
}

// ListVendors lists vendors with pagination
func (c *OperatorClient) ListVendors(ctx context.Context, page, limit int32) (*operatorv1.ListVendorsResponse, error) {
	return c.client.ListVendors(ctx, &operatorv1.ListVendorsRequest{Page: page, Limit: limit})
}

// DeleteVendor deletes a vendor by ID
func (c *OperatorClient) DeleteVendor(ctx context.Context, id string) (*operatorv1.DeleteVendorResponse, error) {
	return c.client.DeleteVendor(ctx, &operatorv1.DeleteVendorRequest{Id: id})
}

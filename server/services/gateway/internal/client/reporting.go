package client

import (
	"context"
	"time"

	reportingv1 "github.com/MuhibNayem/Travio/server/api/proto/reporting/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// ReportingClient wraps the gRPC reporting client
type ReportingClient struct {
	conn   *grpc.ClientConn
	client reportingv1.ReportingServiceClient
}

// NewReportingClient creates a new gRPC client for the reporting service
func NewReportingClient(address string, tlsCfg TLSConfig) (*ReportingClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to reporting service", "address", address, "tls", tlsCfg.CertFile != "")
	return &ReportingClient{
		conn:   conn,
		client: reportingv1.NewReportingServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *ReportingClient) Close() error {
	return c.conn.Close()
}

// GetRevenueReport retrieves revenue data
func (c *ReportingClient) GetRevenueReport(ctx context.Context, req *reportingv1.RevenueReportRequest) (*reportingv1.RevenueReportResponse, error) {
	return c.client.GetRevenueReport(ctx, req)
}

// GetBookingTrends retrieves booking trends
func (c *ReportingClient) GetBookingTrends(ctx context.Context, req *reportingv1.BookingTrendsRequest) (*reportingv1.BookingTrendsResponse, error) {
	return c.client.GetBookingTrends(ctx, req)
}

// GetTopRoutes retrieves top routes
func (c *ReportingClient) GetTopRoutes(ctx context.Context, req *reportingv1.TopRoutesRequest) (*reportingv1.TopRoutesResponse, error) {
	return c.client.GetTopRoutes(ctx, req)
}

// GetOrganizationMetrics retrieves organization metrics
func (c *ReportingClient) GetOrganizationMetrics(ctx context.Context, req *reportingv1.OrganizationMetricsRequest) (*reportingv1.OrganizationMetricsResponse, error) {
	return c.client.GetOrganizationMetrics(ctx, req)
}

// ExportReport exports report data
func (c *ReportingClient) ExportReport(ctx context.Context, req *reportingv1.ExportReportRequest) (*reportingv1.ExportReportResponse, error) {
	return c.client.ExportReport(ctx, req)
}

// Health checks the reporting service health
func (c *ReportingClient) Health(ctx context.Context) (*reportingv1.HealthResponse, error) {
	return c.client.Health(ctx, &reportingv1.HealthRequest{})
}

package client

import (
	"context"
	"time"

	crmv1 "github.com/MuhibNayem/Travio/server/api/proto/crm/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

type CRMClient struct {
	conn   *grpc.ClientConn
	client crmv1.CRMServiceClient
}

func NewCRMClient(address string, tlsCfg TLSConfig) (*CRMClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to crm service", "address", address, "tls", tlsCfg.CertFile != "")
	return &CRMClient{
		conn:   conn,
		client: crmv1.NewCRMServiceClient(conn),
	}, nil
}

func (c *CRMClient) Close() error {
	return c.conn.Close()
}

// --- Coupons ---

func (c *CRMClient) CreateCoupon(ctx context.Context, req *crmv1.CreateCouponRequest) (*crmv1.Coupon, error) {
	return c.client.CreateCoupon(ctx, req)
}

func (c *CRMClient) GetCoupon(ctx context.Context, req *crmv1.GetCouponRequest) (*crmv1.Coupon, error) {
	return c.client.GetCoupon(ctx, req)
}

func (c *CRMClient) ListCoupons(ctx context.Context, req *crmv1.ListCouponsRequest) (*crmv1.ListCouponsResponse, error) {
	return c.client.ListCoupons(ctx, req)
}

func (c *CRMClient) ValidateCoupon(ctx context.Context, req *crmv1.ValidateCouponRequest) (*crmv1.ValidateCouponResponse, error) {
	return c.client.ValidateCoupon(ctx, req)
}

// --- Support ---

func (c *CRMClient) CreateTicket(ctx context.Context, req *crmv1.CreateTicketRequest) (*crmv1.SupportTicket, error) {
	return c.client.CreateTicket(ctx, req)
}

func (c *CRMClient) ListTickets(ctx context.Context, req *crmv1.ListTicketsRequest) (*crmv1.ListTicketsResponse, error) {
	return c.client.ListTickets(ctx, req)
}

func (c *CRMClient) AddTicketMessage(ctx context.Context, req *crmv1.AddTicketMessageRequest) (*crmv1.TicketMessage, error) {
	return c.client.AddTicketMessage(ctx, req)
}

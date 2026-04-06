package clients

import (
	"context"
	"fmt"

	crmv1 "github.com/MuhibNayem/Travio/server/api/proto/crm/v1"
	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CRMClient implements coupon validation via gRPC
type CRMClient struct {
	client crmv1.CRMServiceClient
}

// NewCRMClient creates a new CRM gRPC client
func NewCRMClient(addr string) (*CRMClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CRM service: %w", err)
	}
	return &CRMClient{client: crmv1.NewCRMServiceClient(conn)}, nil
}

// NewCRMClientFromConn creates a CRM client from an existing connection
func NewCRMClientFromConn(conn *grpc.ClientConn) *CRMClient {
	return &CRMClient{client: crmv1.NewCRMServiceClient(conn)}
}

// ValidateCoupon validates a coupon code for an organization
func (c *CRMClient) ValidateCoupon(ctx context.Context, orgID, code string, cartTotal int64) (*domain.CouponValidation, error) {
	resp, err := c.client.ValidateCoupon(ctx, &crmv1.ValidateCouponRequest{
		OrganizationId: orgID,
		Code:           code,
	})
	if err != nil {
		return nil, fmt.Errorf("coupon validation failed: %w", err)
	}

	return &domain.CouponValidation{
		Valid:         resp.Valid,
		DiscountPaisa: resp.DiscountAmount,
		Message:       resp.Message,
		CouponCode:    code,
	}, nil
}

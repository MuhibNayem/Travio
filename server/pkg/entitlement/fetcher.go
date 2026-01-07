package entitlement

import (
	"context"
	"strconv"

	subscriptionv1 "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SubscriptionFetcher implements EntitlementFetcher using the Subscription gRPC service.
type SubscriptionFetcher struct {
	client subscriptionv1.SubscriptionServiceClient
	conn   *grpc.ClientConn
}

// NewSubscriptionFetcher creates a new fetcher connected to the subscription service.
func NewSubscriptionFetcher(addr string) (*SubscriptionFetcher, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &SubscriptionFetcher{
		client: subscriptionv1.NewSubscriptionServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection.
func (f *SubscriptionFetcher) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}

// FetchEntitlement retrieves entitlements from the Subscription service using GetEntitlement RPC.
func (f *SubscriptionFetcher) FetchEntitlement(ctx context.Context, orgID string) (*Entitlements, error) {
	resp, err := f.client.GetEntitlement(ctx, &subscriptionv1.GetEntitlementRequest{
		OrganizationId: orgID,
	})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil // No subscription
	}

	// Build quota limits from features (numeric values)
	quotaLimits := make(map[string]int64)
	for key, value := range resp.Features {
		if num, err := strconv.ParseInt(value, 10, 64); err == nil {
			quotaLimits[key] = num
		}
	}

	// Merge with explicit quota limits from response
	for key, value := range resp.QuotaLimits {
		quotaLimits[key] = value
	}

	return &Entitlements{
		OrganizationID:  resp.OrganizationId,
		PlanID:          resp.PlanId,
		PlanName:        resp.PlanName,
		Status:          resp.Status,
		Features:        resp.Features,
		QuotaLimits:     quotaLimits,
		UsageThisPeriod: resp.UsageThisPeriod,
	}, nil
}

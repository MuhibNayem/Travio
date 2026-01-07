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

// FetchEntitlement retrieves entitlements from the Subscription service.
func (f *SubscriptionFetcher) FetchEntitlement(ctx context.Context, orgID string) (*Entitlements, error) {
	// Get subscription
	sub, err := f.client.GetSubscription(ctx, &subscriptionv1.GetSubscriptionRequest{
		OrganizationId: orgID,
	})
	if err != nil {
		return nil, err
	}

	if sub == nil {
		return nil, nil // No subscription
	}

	// Get the plan to fetch features
	plan, err := f.client.GetPlan(ctx, &subscriptionv1.GetPlanRequest{
		PlanId: sub.PlanId,
	})
	if err != nil {
		// Return subscription without plan features
		return &Entitlements{
			OrganizationID: orgID,
			PlanID:         sub.PlanId,
			Status:         sub.Status,
			Features:       map[string]string{},
			QuotaLimits:    map[string]int64{},
		}, nil
	}

	// Build quota limits from features
	quotaLimits := make(map[string]int64)
	for key, value := range plan.Features {
		if num, err := strconv.ParseInt(value, 10, 64); err == nil {
			quotaLimits[key] = num
		}
	}

	return &Entitlements{
		OrganizationID:  orgID,
		PlanID:          sub.PlanId,
		PlanName:        plan.Name,
		Status:          sub.Status,
		Features:        plan.Features,
		QuotaLimits:     quotaLimits,
		UsageThisPeriod: map[string]int64{}, // TODO: Fetch from usage tracking
	}, nil
}

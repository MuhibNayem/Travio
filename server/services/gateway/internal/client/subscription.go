package client

import (
	"context"
	"time"

	subscriptionv1 "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// SubscriptionClient wraps the gRPC subscription client
type SubscriptionClient struct {
	conn   *grpc.ClientConn
	client subscriptionv1.SubscriptionServiceClient
}

// NewSubscriptionClient creates a new gRPC client for the subscription service
func NewSubscriptionClient(address string, tlsCfg TLSConfig) (*SubscriptionClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to subscription service", "address", address, "tls", tlsCfg.CertFile != "")
	return &SubscriptionClient{
		conn:   conn,
		client: subscriptionv1.NewSubscriptionServiceClient(conn),
	}, nil
}

func (c *SubscriptionClient) Close() error {
	return c.conn.Close()
}

// Plans
func (c *SubscriptionClient) CreatePlan(ctx context.Context, req *subscriptionv1.CreatePlanRequest) (*subscriptionv1.Plan, error) {
	return c.client.CreatePlan(ctx, req)
}

func (c *SubscriptionClient) ListPlans(ctx context.Context, includeInactive bool) (*subscriptionv1.ListPlansResponse, error) {
	return c.client.ListPlans(ctx, &subscriptionv1.ListPlansRequest{IncludeInactive: includeInactive})
}

func (c *SubscriptionClient) GetPlan(ctx context.Context, id string) (*subscriptionv1.Plan, error) {
	return c.client.GetPlan(ctx, &subscriptionv1.GetPlanRequest{PlanId: id})
}

// Subscriptions
func (c *SubscriptionClient) CreateSubscription(ctx context.Context, req *subscriptionv1.CreateSubscriptionRequest) (*subscriptionv1.Subscription, error) {
	return c.client.CreateSubscription(ctx, req)
}

func (c *SubscriptionClient) GetSubscription(ctx context.Context, orgID string) (*subscriptionv1.Subscription, error) {
	return c.client.GetSubscription(ctx, &subscriptionv1.GetSubscriptionRequest{OrganizationId: orgID})
}

func (c *SubscriptionClient) CancelSubscription(ctx context.Context, orgID string) (*subscriptionv1.Subscription, error) {
	return c.client.CancelSubscription(ctx, &subscriptionv1.CancelSubscriptionRequest{OrganizationId: orgID})
}

// Billing
func (c *SubscriptionClient) ListInvoices(ctx context.Context, subscriptionID string) (*subscriptionv1.ListInvoicesResponse, error) {
	return c.client.ListInvoices(ctx, &subscriptionv1.ListInvoicesRequest{SubscriptionId: subscriptionID})
}

// Admin
func (c *SubscriptionClient) ListSubscriptions(ctx context.Context, planID, status string) (*subscriptionv1.ListSubscriptionsResponse, error) {
	return c.client.ListSubscriptions(ctx, &subscriptionv1.ListSubscriptionsRequest{PlanId: planID, Status: status})
}

func (c *SubscriptionClient) UpdatePlan(ctx context.Context, req *subscriptionv1.UpdatePlanRequest) (*subscriptionv1.Plan, error) {
	return c.client.UpdatePlan(ctx, req)
}

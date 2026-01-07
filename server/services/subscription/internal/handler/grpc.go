package handler

import (
	"context"
	"net"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/subscription/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedSubscriptionServiceServer
	svc *service.SubscriptionService
}

func NewGRPCHandler(svc *service.SubscriptionService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.Plan, error) {
	plan, err := h.svc.CreatePlan(ctx, req.Name, req.Description, req.PricePaisa, req.Interval, req.Features)
	if err != nil {
		logger.Error("Failed to create plan", "error", err)
		return nil, status.Error(codes.Internal, "failed to create plan")
	}
	return toProtoPlan(plan), nil
}

func (h *GRPCHandler) ListPlans(ctx context.Context, req *pb.ListPlansRequest) (*pb.ListPlansResponse, error) {
	plans, err := h.svc.ListPlans(ctx, req.IncludeInactive)
	if err != nil {
		logger.Error("Failed to list plans", "error", err)
		return nil, status.Error(codes.Internal, "failed to list plans")
	}

	var pbPlans []*pb.Plan
	for _, p := range plans {
		pbPlans = append(pbPlans, toProtoPlan(p))
	}
	return &pb.ListPlansResponse{Plans: pbPlans}, nil
}

func (h *GRPCHandler) GetPlan(ctx context.Context, req *pb.GetPlanRequest) (*pb.Plan, error) {
	plan, err := h.svc.GetPlan(ctx, req.PlanId)
	if err != nil {
		logger.Error("Failed to get plan", "error", err)
		return nil, status.Error(codes.Internal, "failed to get plan")
	}
	if plan == nil {
		return nil, status.Error(codes.NotFound, "plan not found")
	}
	return toProtoPlan(plan), nil
}

func (h *GRPCHandler) UpdatePlan(ctx context.Context, req *pb.UpdatePlanRequest) (*pb.Plan, error) {
	plan, err := h.svc.UpdatePlan(ctx, req.Id, req.Name, req.Description, req.PricePaisa, req.IsActive, req.Features)
	if err != nil {
		logger.Error("Failed to update plan", "error", err)
		return nil, status.Error(codes.Internal, "failed to update plan")
	}
	return toProtoPlan(plan), nil
}

func (h *GRPCHandler) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (*pb.Subscription, error) {
	sub, err := h.svc.CreateSubscription(ctx, req.OrganizationId, req.PlanId)
	if err != nil {
		logger.Error("Failed to create subscription", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoSubscription(sub), nil
}

func (h *GRPCHandler) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.Subscription, error) {
	sub, err := h.svc.GetSubscription(ctx, req.OrganizationId)
	if err != nil {
		logger.Error("Failed to get subscription", "error", err)
		return nil, status.Error(codes.Internal, "failed to get subscription")
	}
	if sub == nil {
		return nil, status.Error(codes.NotFound, "active subscription not found")
	}
	return toProtoSubscription(sub), nil
}

func (h *GRPCHandler) CancelSubscription(ctx context.Context, req *pb.CancelSubscriptionRequest) (*pb.Subscription, error) {
	_, err := h.svc.CancelSubscription(ctx, req.OrganizationId)
	if err != nil {
		logger.Error("Failed to cancel subscription", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Subscription{Status: "canceled"}, nil
}

func (h *GRPCHandler) ListSubscriptions(ctx context.Context, req *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	subs, err := h.svc.ListSubscriptions(ctx, req.PlanId, req.Status)
	if err != nil {
		logger.Error("Failed to list subscriptions", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pbSubs []*pb.Subscription
	for _, s := range subs {
		pbSubs = append(pbSubs, toProtoSubscription(s))
	}
	return &pb.ListSubscriptionsResponse{Subscriptions: pbSubs}, nil
}

func (h *GRPCHandler) ListInvoices(ctx context.Context, req *pb.ListInvoicesRequest) (*pb.ListInvoicesResponse, error) {
	invoices, err := h.svc.ListInvoices(ctx, req.SubscriptionId)
	if err != nil {
		logger.Error("Failed to list invoices", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pbInvoices []*pb.Invoice
	for _, i := range invoices {
		pbInvoices = append(pbInvoices, toProtoInvoice(i))
	}
	return &pb.ListInvoicesResponse{Invoices: pbInvoices}, nil
}

// Helpers

func toProtoPlan(p *repository.Plan) *pb.Plan {
	if p == nil {
		return nil
	}
	return &pb.Plan{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PricePaisa:  p.PricePaisa,
		Interval:    p.Interval,
		Features:    p.Features,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoSubscription(s *repository.Subscription) *pb.Subscription {
	if s == nil {
		return nil
	}
	return &pb.Subscription{
		Id:                 s.ID,
		OrganizationId:     s.OrganizationID,
		PlanId:             s.PlanID,
		Status:             s.Status,
		CurrentPeriodStart: s.CurrentPeriodStart.Format(time.RFC3339),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format(time.RFC3339),
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          s.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoInvoice(i *repository.Invoice) *pb.Invoice {
	if i == nil {
		return nil
	}
	paidAt := ""
	if i.PaidAt != nil {
		paidAt = i.PaidAt.Format(time.RFC3339)
	}
	return &pb.Invoice{
		Id:             i.ID,
		SubscriptionId: i.SubscriptionID,
		AmountPaisa:    i.AmountPaisa,
		Status:         i.Status,
		IssuedAt:       i.IssuedAt.Format(time.RFC3339),
		DueDate:        i.DueDate.Format(time.RFC3339),
		PaidAt:         paidAt,
	}
}

func StartGRPCServer(port string, svc *service.SubscriptionService) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	// Register the service
	pb.RegisterSubscriptionServiceServer(server, NewGRPCHandler(svc))

	logger.Info("Subscription gRPC server starting", "port", port)
	return server.Serve(listener)
}

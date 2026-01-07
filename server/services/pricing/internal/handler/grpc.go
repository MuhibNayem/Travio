package handler

import (
	"context"
	"net"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCHandler implements the Pricing gRPC service
type GRPCHandler struct {
	svc *service.PricingService
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(svc *service.PricingService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

// CalculatePriceRequest for JSON/HTTP compatibility
type CalculatePriceRequest struct {
	TripID         string  `json:"trip_id"`
	SeatClass      string  `json:"seat_class"`
	Date           string  `json:"date"`
	Quantity       int32   `json:"quantity"`
	BasePricePaisa int64   `json:"base_price_paisa"`
	OccupancyRate  float64 `json:"occupancy_rate"`
}

// CalculatePriceResponse for JSON/HTTP compatibility
type CalculatePriceResponse struct {
	FinalPricePaisa int64         `json:"final_price_paisa"`
	BasePricePaisa  int64         `json:"base_price_paisa"`
	AppliedRules    []AppliedRule `json:"applied_rules"`
}

// AppliedRule for JSON/HTTP compatibility
type AppliedRule struct {
	RuleID     string  `json:"rule_id"`
	RuleName   string  `json:"rule_name"`
	Multiplier float64 `json:"multiplier"`
}

// CalculatePrice handles price calculation via gRPC
func (h *GRPCHandler) CalculatePrice(ctx context.Context, req *CalculatePriceRequest) (*CalculatePriceResponse, error) {
	result, err := h.svc.CalculatePrice(ctx, &service.CalculatePriceRequest{
		TripID:         req.TripID,
		SeatClass:      req.SeatClass,
		Date:           req.Date,
		Quantity:       int(req.Quantity),
		BasePricePaisa: req.BasePricePaisa,
		OccupancyRate:  req.OccupancyRate,
	})
	if err != nil {
		return nil, err
	}

	var appliedRules []AppliedRule
	for _, r := range result.AppliedRules {
		appliedRules = append(appliedRules, AppliedRule{
			RuleID:     r.RuleID,
			RuleName:   r.RuleName,
			Multiplier: r.Multiplier,
		})
	}

	return &CalculatePriceResponse{
		FinalPricePaisa: result.FinalPricePaisa,
		BasePricePaisa:  result.BasePricePaisa,
		AppliedRules:    appliedRules,
	}, nil
}

// GetRulesResponse for JSON/HTTP compatibility
type GetRulesResponse struct {
	Rules []*repository.PricingRule `json:"rules"`
}

// GetRules returns all pricing rules
func (h *GRPCHandler) GetRules(ctx context.Context, includeInactive bool) (*GetRulesResponse, error) {
	rules, err := h.svc.GetRules(ctx, includeInactive)
	if err != nil {
		return nil, err
	}
	return &GetRulesResponse{Rules: rules}, nil
}

// CreateRule creates a new pricing rule
func (h *GRPCHandler) CreateRule(ctx context.Context, rule *repository.PricingRule) (*repository.PricingRule, error) {
	if err := h.svc.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// StartGRPCServer starts the gRPC server (simple implementation without generated proto)
func StartGRPCServer(port string, svc *service.PricingService) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	reflection.Register(server)

	// Note: Without generated proto code, we can't register the service properly.
	// For now, the service is accessible via the HTTP handler or direct Go calls.

	logger.Info("Pricing gRPC server starting", "port", port)
	return server.Serve(listener)
}

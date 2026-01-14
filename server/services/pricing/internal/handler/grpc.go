package handler

import (
	"context"
	"net"
	"time"

	pricingv1 "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCHandler implements the PricingService gRPC server.
type GRPCHandler struct {
	pricingv1.UnimplementedPricingServiceServer
	svc *service.PricingService
}

// NewGRPCHandler creates a new gRPC handler.
func NewGRPCHandler(svc *service.PricingService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CalculatePrice(ctx context.Context, req *pricingv1.CalculatePriceRequest) (*pricingv1.CalculatePriceResponse, error) {
	result, err := h.svc.CalculatePrice(ctx, &service.CalculatePriceRequest{
		TripID:         req.TripId,
		SeatClass:      req.SeatClass,
		SeatCategory:   req.SeatCategory,
		Date:           req.Date,
		Quantity:       int(req.Quantity),
		BasePricePaisa: req.BasePricePaisa,
		OccupancyRate:  req.OccupancyRate,
		OrganizationID: req.OrganizationId,
		DepartureTime:  req.DepartureTime,
		RouteID:        req.RouteId,
		ScheduleID:     req.ScheduleId,
		FromStationID:  req.FromStationId,
		ToStationID:    req.ToStationId,
		VehicleType:    req.VehicleType,
		VehicleClass:   req.VehicleClass,
		PromoCode:      req.PromoCode,
	})
	if err != nil {
		return nil, err
	}

	var appliedRules []*pricingv1.AppliedRule
	for _, rule := range result.AppliedRules {
		appliedRules = append(appliedRules, &pricingv1.AppliedRule{
			RuleId:     rule.RuleID,
			RuleName:   rule.RuleName,
			Multiplier: rule.Multiplier,
		})
	}

	var promoApplied *pricingv1.PromotionApplied
	if result.PromotionApplied != nil {
		promoApplied = &pricingv1.PromotionApplied{
			PromoCode:           result.PromotionApplied.Code,
			DiscountAmountPaisa: result.PromotionApplied.DiscountAmountPaisa,
		}
	}

	return &pricingv1.CalculatePriceResponse{
		FinalPricePaisa:  result.FinalPricePaisa,
		BasePricePaisa:   result.BasePricePaisa,
		AppliedRules:     appliedRules,
		PromotionApplied: promoApplied,
	}, nil
}

func (h *GRPCHandler) GetRules(ctx context.Context, req *pricingv1.GetRulesRequest) (*pricingv1.GetRulesResponse, error) {
	rules, err := h.svc.GetRules(ctx, req.IncludeInactive, req.OrganizationId)
	if err != nil {
		return nil, err
	}
	return &pricingv1.GetRulesResponse{Rules: pricingRulesToProto(rules)}, nil
}

func (h *GRPCHandler) CreateRule(ctx context.Context, req *pricingv1.CreateRuleRequest) (*pricingv1.CreateRuleResponse, error) {
	rule := protoToPricingRule(req.OrganizationId, req.Name, req.Description, req.Condition, req.Multiplier, req.AdjustmentType, req.AdjustmentValue, req.Priority)
	if err := h.svc.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return &pricingv1.CreateRuleResponse{Rule: pricingRuleToProto(rule)}, nil
}

func (h *GRPCHandler) UpdateRule(ctx context.Context, req *pricingv1.UpdateRuleRequest) (*pricingv1.UpdateRuleResponse, error) {
	rule := protoToPricingRule("", req.Name, req.Description, req.Condition, req.Multiplier, req.AdjustmentType, req.AdjustmentValue, req.Priority)
	rule.ID = req.Id
	rule.IsActive = req.IsActive
	if err := h.svc.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}
	return &pricingv1.UpdateRuleResponse{Rule: pricingRuleToProto(rule)}, nil
}

func (h *GRPCHandler) DeleteRule(ctx context.Context, req *pricingv1.DeleteRuleRequest) (*pricingv1.DeleteRuleResponse, error) {
	if err := h.svc.DeleteRule(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pricingv1.DeleteRuleResponse{Success: true}, nil
}

// StartGRPCServer starts the gRPC server.
func StartGRPCServer(port string, svc *service.PricingService) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pricingv1.RegisterPricingServiceServer(server, NewGRPCHandler(svc))
	reflection.Register(server)

	logger.Info("Pricing gRPC server starting", "port", port)
	return server.Serve(listener)
}

func pricingRulesToProto(rules []*repository.PricingRule) []*pricingv1.PricingRule {
	out := make([]*pricingv1.PricingRule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, pricingRuleToProto(rule))
	}
	return out
}

func pricingRuleToProto(rule *repository.PricingRule) *pricingv1.PricingRule {
	if rule == nil {
		return nil
	}
	orgID := ""
	if rule.OrganizationID != nil {
		orgID = *rule.OrganizationID
	}
	return &pricingv1.PricingRule{
		Id:              rule.ID,
		OrganizationId:  orgID,
		Name:            rule.Name,
		Description:     rule.Description,
		Condition:       rule.Condition,
		Multiplier:      rule.Multiplier,
		AdjustmentType:  rule.AdjustmentType,
		AdjustmentValue: rule.AdjustmentValue,
		Priority:        int32(rule.Priority),
		IsActive:        rule.IsActive,
		CreatedAt:       rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       rule.UpdatedAt.Format(time.RFC3339),
	}
}

func protoToPricingRule(orgID, name, description, condition string, multiplier float64, adjustmentType string, adjustmentValue float64, priority int32) *repository.PricingRule {
	var orgPtr *string
	if orgID != "" {
		orgPtr = &orgID
	}
	return &repository.PricingRule{
		OrganizationID:  orgPtr,
		Name:            name,
		Description:     description,
		Condition:       condition,
		Multiplier:      multiplier,
		AdjustmentType:  adjustmentType,
		AdjustmentValue: adjustmentValue,
		Priority:        int(priority),
		IsActive:        true,
	}
}

// --- Promotion Handlers ---

func (h *GRPCHandler) CreatePromotion(ctx context.Context, req *pricingv1.CreatePromotionRequest) (*pricingv1.CreatePromotionResponse, error) {
	var validFrom, validUntil *time.Time
	if req.ValidFrom != "" {
		t, err := time.Parse("2006-01-02", req.ValidFrom)
		if err == nil {
			validFrom = &t
		}
	}
	if req.ValidUntil != "" {
		t, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = &t
		}
	}

	var orgPtr *string
	if req.OrganizationId != "" {
		org := req.OrganizationId
		orgPtr = &org
	}

	promo := &repository.Promotion{
		Code:                req.Code,
		Description:         req.Description,
		DiscountType:        req.DiscountType,
		DiscountValue:       req.DiscountValue,
		MaxUsage:            req.MaxUsage,
		OrganizationID:      orgPtr,
		ValidFrom:           validFrom,
		ValidUntil:          validUntil,
		MinOrderAmountPaisa: req.MinOrderAmountPaisa,
		IsActive:            true,
	}

	if err := h.svc.CreatePromotion(ctx, promo); err != nil {
		return nil, err
	}

	return &pricingv1.CreatePromotionResponse{
		Promotion: promotionToProto(promo),
	}, nil
}

func (h *GRPCHandler) GetPromotions(ctx context.Context, req *pricingv1.GetPromotionsRequest) (*pricingv1.GetPromotionsResponse, error) {
	promos, err := h.svc.GetPromotions(ctx, req.OrganizationId, req.ActiveOnly)
	if err != nil {
		return nil, err
	}

	var pbPromos []*pricingv1.Promotion
	for _, p := range promos {
		pbPromos = append(pbPromos, promotionToProto(p))
	}

	return &pricingv1.GetPromotionsResponse{
		Promotions: pbPromos,
	}, nil
}

func promotionToProto(p *repository.Promotion) *pricingv1.Promotion {
	if p == nil {
		return nil
	}
	var validFrom, validUntil string
	if p.ValidFrom != nil {
		validFrom = p.ValidFrom.Format("2006-01-02")
	}
	if p.ValidUntil != nil {
		validUntil = p.ValidUntil.Format("2006-01-02")
	}
	orgID := ""
	if p.OrganizationID != nil {
		orgID = *p.OrganizationID
	}
	return &pricingv1.Promotion{
		Id:                  p.ID,
		Code:                p.Code,
		Description:         p.Description,
		DiscountType:        p.DiscountType,
		DiscountValue:       p.DiscountValue,
		MaxUsage:            p.MaxUsage,
		CurrentUsage:        p.CurrentUsage,
		ValidFrom:           validFrom,
		ValidUntil:          validUntil,
		MinOrderAmountPaisa: p.MinOrderAmountPaisa,
		IsActive:            p.IsActive,
		OrganizationId:      orgID,
	}
}

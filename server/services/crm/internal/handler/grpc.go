package handler

import (
	"context"
	"time"

	crmv1 "github.com/MuhibNayem/Travio/server/api/proto/crm/v1"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	crmv1.UnimplementedCRMServiceServer
	service *service.CRMService
}

func NewGRPCHandler(svc *service.CRMService) *GRPCHandler {
	return &GRPCHandler{service: svc}
}

// --- Coupons ---

func (h *GRPCHandler) CreateCoupon(ctx context.Context, req *crmv1.CreateCouponRequest) (*crmv1.Coupon, error) {
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		// Default to 1 year if invalid
		endDate = time.Now().AddDate(1, 0, 0)
	}

	dType := domain.DiscountType(req.DiscountType)

	coupon, err := h.service.CreateCoupon(ctx, req.OrganizationId, req.Code, dType, req.DiscountValue, endDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return mapCouponToProto(coupon), nil
}

func (h *GRPCHandler) GetCoupon(ctx context.Context, req *crmv1.GetCouponRequest) (*crmv1.Coupon, error) {
	coupon, err := h.service.GetCoupon(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return mapCouponToProto(coupon), nil
}

func (h *GRPCHandler) ListCoupons(ctx context.Context, req *crmv1.ListCouponsRequest) (*crmv1.ListCouponsResponse, error) {
	coupons, err := h.service.ListCoupons(ctx, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoCoupons []*crmv1.Coupon
	for _, c := range coupons {
		protoCoupons = append(protoCoupons, mapCouponToProto(c))
	}
	return &crmv1.ListCouponsResponse{Coupons: protoCoupons}, nil
}

func (h *GRPCHandler) UpdateCoupon(ctx context.Context, req *crmv1.UpdateCouponRequest) (*crmv1.Coupon, error) {
	coupon, err := h.service.UpdateCoupon(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapCouponToProto(coupon), nil
}

func (h *GRPCHandler) ValidateCoupon(ctx context.Context, req *crmv1.ValidateCouponRequest) (*crmv1.ValidateCouponResponse, error) {
	valid, amount, msg, err := h.service.ValidateCoupon(ctx, req.Code, req.OrganizationId, req.CartTotalAmount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &crmv1.ValidateCouponResponse{
		Valid:          valid,
		DiscountAmount: amount,
		Message:        msg,
	}, nil
}

// --- Support ---

func (h *GRPCHandler) CreateTicket(ctx context.Context, req *crmv1.CreateTicketRequest) (*crmv1.SupportTicket, error) {
	ticket, err := h.service.CreateTicket(ctx, req.UserId, req.Subject, req.Message)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapTicketToProto(ticket), nil
}

func (h *GRPCHandler) ListTickets(ctx context.Context, req *crmv1.ListTicketsRequest) (*crmv1.ListTicketsResponse, error) {
	tickets, err := h.service.ListTickets(ctx, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoTickets []*crmv1.SupportTicket
	for _, t := range tickets {
		protoTickets = append(protoTickets, mapTicketToProto(t))
	}
	return &crmv1.ListTicketsResponse{Tickets: protoTickets}, nil
}

func (h *GRPCHandler) AddTicketMessage(ctx context.Context, req *crmv1.AddTicketMessageRequest) (*crmv1.TicketMessage, error) {
	msg, err := h.service.AddTicketMessage(ctx, req.TicketId, req.SenderId, req.Message)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &crmv1.TicketMessage{
		Id:        msg.ID,
		TicketId:  msg.TicketID,
		SenderId:  msg.SenderID,
		Message:   msg.Message,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}, nil
}

// --- Helpers ---

func mapCouponToProto(c *domain.Coupon) *crmv1.Coupon {
	return &crmv1.Coupon{
		Id:                c.ID,
		OrganizationId:    c.OrganizationID,
		Code:              c.Code,
		DiscountType:      crmv1.DiscountType(c.DiscountType),
		DiscountValue:     c.DiscountValue,
		MinPurchaseAmount: c.MinPurchaseAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		StartDate:         c.StartDate.Format(time.RFC3339),
		EndDate:           c.EndDate.Format(time.RFC3339),
		UsageLimit:        c.UsageLimit,
		UsageCount:        c.UsageCount,
		IsActive:          c.IsActive,
	}
}

func mapTicketToProto(t *domain.SupportTicket) *crmv1.SupportTicket {
	return &crmv1.SupportTicket{
		Id:             t.ID,
		OrganizationId: t.OrganizationID,
		UserId:         t.UserID,
		Subject:        t.Subject,
		Status:         t.Status,
		Priority:       t.Priority,
		CreatedAt:      t.CreatedAt.Format(time.RFC3339),
	}
}

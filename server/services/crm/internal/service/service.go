package service

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/crm/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/repository"
)

type CRMService struct {
	repo *repository.CRMRepository
}

func NewCRMService(repo *repository.CRMRepository) *CRMService {
	return &CRMService{repo: repo}
}

// --- Coupons ---

func (s *CRMService) CreateCoupon(ctx context.Context, orgID, code string, dType domain.DiscountType, dValue float64, endDate time.Time) (*domain.Coupon, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}

	coupon := &domain.Coupon{
		OrganizationID:    orgID,
		Code:              code,
		DiscountType:      dType,
		DiscountValue:     dValue,
		MinPurchaseAmount: 0,
		MaxDiscountAmount: 0,
		StartDate:         time.Now(),
		EndDate:           endDate,
		UsageLimit:        1000, // Default limit
		UsageCount:        0,
		IsActive:          true,
	}

	if err := s.repo.CreateCoupon(coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (s *CRMService) ValidateCoupon(ctx context.Context, code, orgID string, cartTotal int64) (bool, int64, string, error) {
	coupon, err := s.repo.GetCouponByCode(code, orgID)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return false, 0, "Invalid coupon code", nil
		}
		return false, 0, err.Error(), err
	}

	if !coupon.IsActive {
		return false, 0, "Coupon is inactive", nil
	}

	if time.Now().After(coupon.EndDate) {
		return false, 0, "Coupon expired", nil
	}

	if coupon.UsageLimit > 0 && coupon.UsageCount >= coupon.UsageLimit {
		return false, 0, "Coupon usage limit reached", nil
	}

	if cartTotal < coupon.MinPurchaseAmount {
		return false, 0, "Minimum purchase amount not met", nil
	}

	// Calculate Discount
	var discount int64
	if coupon.DiscountType == domain.DiscountTypeFixedAmount {
		discount = int64(coupon.DiscountValue)
		if discount > cartTotal {
			discount = cartTotal
		}
	} else if coupon.DiscountType == domain.DiscountTypePercentage {
		discount = int64(float64(cartTotal) * (coupon.DiscountValue / 100.0))
		if coupon.MaxDiscountAmount > 0 && discount > coupon.MaxDiscountAmount {
			discount = coupon.MaxDiscountAmount
		}
	}

	return true, discount, "Coupon applied successfully", nil
}

func (s *CRMService) ListCoupons(ctx context.Context, orgID string) ([]*domain.Coupon, error) {
	return s.repo.ListCoupons(orgID)
}

func (s *CRMService) GetCoupon(ctx context.Context, id string) (*domain.Coupon, error) {
	return s.repo.GetCoupon(id)
}

func (s *CRMService) UpdateCoupon(ctx context.Context, id string) (*domain.Coupon, error) {
	// Placeholder: implementation for full update not requested in immediate flow,
	// but proto has it. For now, we return existing.
	return s.repo.GetCoupon(id)
}

// --- Support ---

func (s *CRMService) CreateTicket(ctx context.Context, userID, subject, message string) (*domain.SupportTicket, error) {
	ticket := &domain.SupportTicket{
		OrganizationID: "generic-org", // In multi-tenant, this comes from context or input
		UserID:         userID,
		Subject:        subject,
		Priority:       "MEDIUM",
	}

	if err := s.repo.CreateTicket(ticket); err != nil {
		return nil, err
	}

	// Add initial message
	msg := &domain.TicketMessage{
		TicketID: ticket.ID,
		SenderID: userID,
		Message:  message,
	}
	if err := s.repo.CreateTicketMessage(msg); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *CRMService) ListTickets(ctx context.Context, orgID string) ([]*domain.SupportTicket, error) {
	return s.repo.ListTickets(orgID)
}

func (s *CRMService) AddTicketMessage(ctx context.Context, ticketID, senderID, message string) (*domain.TicketMessage, error) {
	msg := &domain.TicketMessage{
		TicketID: ticketID,
		SenderID: senderID,
		Message:  message,
	}
	if err := s.repo.CreateTicketMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

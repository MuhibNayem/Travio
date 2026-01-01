package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/google/uuid"
)

type PaymentService struct {
	registry *gateway.Registry
}

func NewPaymentService(registry *gateway.Registry) *PaymentService {
	return &PaymentService{registry: registry}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentReq) (*PaymentResult, error) {
	gw, err := s.registry.SelectByMethod(req.PaymentMethod)
	if err != nil {
		return nil, err
	}

	gwReq := &gateway.CreatePaymentRequest{
		OrderID:       req.OrderID,
		Amount:        gateway.Money{AmountPaisa: req.AmountPaisa, Currency: req.Currency},
		Currency:      req.Currency,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Description:   req.Description,
		ReturnURL:     req.ReturnURL,
		CancelURL:     req.CancelURL,
		IPNURL:        req.IPNURL,
	}

	resp, err := gw.CreatePayment(ctx, gwReq)
	if err != nil {
		return nil, fmt.Errorf("gateway error: %w", err)
	}

	return &PaymentResult{
		PaymentID:   uuid.New().String(),
		OrderID:     req.OrderID,
		Gateway:     gw.Name(),
		SessionID:   resp.SessionID,
		RedirectURL: resp.RedirectURL,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, gatewayName, transactionID string) (*gateway.PaymentStatus, error) {
	gw, err := s.registry.Get(gatewayName)
	if err != nil {
		return nil, err
	}
	return gw.VerifyPayment(ctx, transactionID)
}

func (s *PaymentService) CapturePayment(ctx context.Context, gatewayName, transactionID string) (*gateway.PaymentStatus, error) {
	gw, err := s.registry.Get(gatewayName)
	if err != nil {
		return nil, err
	}
	return gw.CapturePayment(ctx, transactionID)
}

func (s *PaymentService) RefundPayment(ctx context.Context, gatewayName, transactionID string, amount int64, reason string) (*gateway.RefundResponse, error) {
	gw, err := s.registry.Get(gatewayName)
	if err != nil {
		return nil, err
	}
	return gw.RefundPayment(ctx, transactionID, amount, reason)
}

type CreatePaymentReq struct {
	OrderID       string
	AmountPaisa   int64
	Currency      string
	PaymentMethod string
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	Description   string
	ReturnURL     string
	CancelURL     string
	IPNURL        string
}

type PaymentResult struct {
	PaymentID   string
	OrderID     string
	Gateway     string
	SessionID   string
	RedirectURL string
	Status      string
	CreatedAt   time.Time
}

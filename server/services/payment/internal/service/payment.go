package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/model"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
	"github.com/google/uuid"
)

type PaymentService struct {
	registry   *gateway.Registry
	repo       *repository.TransactionRepository
	configRepo *repository.PaymentConfigRepository
}

func NewPaymentService(registry *gateway.Registry, repo *repository.TransactionRepository, configRepo *repository.PaymentConfigRepository) *PaymentService {
	return &PaymentService{
		registry:   registry,
		repo:       repo,
		configRepo: configRepo,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentReq) (*PaymentResult, error) {
	// 1. Idempotency Check
	// Assuming Attempt 1 for new requests. In real world, client sends idempotency-key.
	// Here we derive it deterministicly from OrderID.
	idempotencyKey := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%d", req.OrderID, 1))).String()

	tx := &model.Transaction{
		OrderID:        req.OrderID,
		OrganizationID: req.OrganizationID,
		Attempt:        1,
		Amount:         req.AmountPaisa,
		Currency:       req.Currency,
		Gateway:        req.PaymentMethod,
		Status:         "PENDING",
		IdempotencyKey: idempotencyKey,
	}

	savedTx, exists, err := s.repo.CreateIdempotent(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 2. Return cached result if exists
	if exists {
		// If Success or Pending, return existing transaction info to avoid double charging.
		// If Failed, we might allow retry, but strictly speaking idempotency means same result.
		// For now, let's enforce strict idempotency for PENDING/SUCCESS.
		if savedTx.Status != "FAILED" {
			return &PaymentResult{
				PaymentID: savedTx.ID,
				OrderID:   savedTx.OrderID,
				Gateway:   savedTx.Gateway,
				SessionID: savedTx.GatewayTxID, // Stored as GatewayTxID
				Status:    savedTx.Status,
				CreatedAt: savedTx.CreatedAt,
				// Note: RedirectURL is lost here unless we persist it.
				// For PENDING, RedirectURL is needed for the user to pay.
				// We should ideally store RedirectURL in DB.
			}, nil
		}
		// If FAILED, we proceed to create a new attempt?
		// Our IdempotencyKey is based on Attempt=1. So we can't really create a NEW attempt with same key.
		// We should probably return the FAILED status too. Strict Idempotency.
		return &PaymentResult{
			PaymentID: savedTx.ID,
			OrderID:   savedTx.OrderID,
			Gateway:   savedTx.Gateway,
			Status:    savedTx.Status,
			CreatedAt: savedTx.CreatedAt,
		}, nil
	}

	// 3. Resolve Gateway Factory and Credentials
	providerName := s.registry.ResolveProvider(req.PaymentMethod)
	factory, err := s.registry.GetFactory(providerName)
	if err != nil {
		return nil, err
	}

	// Fetch Organization Keys
	if req.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id required for direct payment")
	}
	payConfig, err := s.configRepo.GetConfig(ctx, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment config for organization %s: %w", err)
	}

	// Instantiate Gateway client
	gw, err := factory.Create(payConfig.Credentials, false) // TODO: Handle Sandbox flag properly
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway instance: %w", err)
	}

	gwReq := &gateway.CreatePaymentRequest{
		OrderID:       req.OrderID,
		Amount:        gateway.Money{AmountPaisa: req.AmountPaisa, Currency: req.Currency},
		Currency:      req.Currency,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Description:   req.Description,
		ReturnURL:     req.ReturnURL + "&org=" + req.OrganizationID, // Pass org context if needed
		CancelURL:     req.CancelURL + "&org=" + req.OrganizationID,
		IPNURL:        req.IPNURL,
	}

	resp, err := gw.CreatePayment(ctx, gwReq)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, savedTx.ID, "FAILED", "")
		return nil, fmt.Errorf("gateway error: %w", err)
	}

	// 4. Update Status
	_ = s.repo.UpdateStatus(ctx, savedTx.ID, "PENDING", resp.SessionID)

	return &PaymentResult{
		PaymentID:   savedTx.ID,
		OrderID:     req.OrderID,
		Gateway:     gw.Name(),
		SessionID:   resp.SessionID,
		RedirectURL: resp.RedirectURL,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, gatewayName, transactionID string) (*gateway.PaymentStatus, error) {
	// 1. Load Transaction to get OrgID
	tx, err := s.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// 2. Load Config
	payConfig, err := s.configRepo.GetConfig(ctx, tx.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}

	// 3. Create Gateway
	providerName := s.registry.ResolveProvider(gatewayName)
	factory, err := s.registry.GetFactory(providerName)
	if err != nil {
		return nil, err
	}
	gw, err := factory.Create(payConfig.Credentials, false)
	if err != nil {
		return nil, err
	}

	return gw.VerifyPayment(ctx, transactionID)
}

func (s *PaymentService) CapturePayment(ctx context.Context, gatewayName, transactionID string) (*gateway.PaymentStatus, error) {
	// 1. Load Transaction to get OrgID
	tx, err := s.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// 2. Load Config
	payConfig, err := s.configRepo.GetConfig(ctx, tx.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}

	// 3. Create Gateway
	providerName := s.registry.ResolveProvider(gatewayName)
	factory, err := s.registry.GetFactory(providerName)
	if err != nil {
		return nil, err
	}
	gw, err := factory.Create(payConfig.Credentials, false)
	if err != nil {
		return nil, err
	}

	return gw.CapturePayment(ctx, transactionID)
}

func (s *PaymentService) RefundPayment(ctx context.Context, gatewayName, transactionID string, amount int64, reason string) (*gateway.RefundResponse, error) {
	// 1. Load Transaction to get OrgID
	tx, err := s.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// 2. Load Config
	payConfig, err := s.configRepo.GetConfig(ctx, tx.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}

	// 3. Create Gateway
	providerName := s.registry.ResolveProvider(gatewayName)
	factory, err := s.registry.GetFactory(providerName)
	if err != nil {
		return nil, err
	}
	gw, err := factory.Create(payConfig.Credentials, false)
	if err != nil {
		return nil, err
	}

	return gw.RefundPayment(ctx, transactionID, amount, reason)
}

type CreatePaymentReq struct {
	OrderID        string
	OrganizationID string
	AmountPaisa    int64
	Currency       string
	PaymentMethod  string
	CustomerName   string
	CustomerEmail  string
	CustomerPhone  string
	Description    string
	ReturnURL      string
	CancelURL      string
	IPNURL         string
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

func (s *PaymentService) UpdatePaymentConfig(ctx context.Context, organizationID, gatewayName string, credentials map[string]string, isActive bool) error {
	// Validate Gateway Name
	providerName := s.registry.ResolveProvider(gatewayName)
	if _, err := s.registry.GetFactory(providerName); err != nil {
		return fmt.Errorf("invalid gateway: %s", gatewayName)
	}

	// Marshal Credentials
	credsJSON, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	config := &repository.PaymentConfig{
		OrganizationID: organizationID,
		Gateway:        gatewayName,
		Credentials:    credsJSON,
		IsActive:       isActive,
		UpdatedAt:      time.Now(),
	}
	// Repo Save logic: "r.db.WithContext(ctx).Save(config)" -> GORM Save performs Insert or Update
	if err := s.configRepo.SaveConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func (s *PaymentService) GetPaymentConfig(ctx context.Context, organizationID string) (*repository.PaymentConfig, error) {
	return s.configRepo.GetConfig(ctx, organizationID)
}

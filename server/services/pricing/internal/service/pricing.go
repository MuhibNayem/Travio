package service

import (
	"context"
	"sync"

	"github.com/MuhibNayem/Travio/server/services/pricing/internal/engine"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
)

// PricingService handles pricing calculations
type PricingService struct {
	repo   *repository.PostgresRepository
	engine *engine.RulesEngine
	mu     sync.RWMutex
}

// NewPricingService creates a new pricing service
func NewPricingService(repo *repository.PostgresRepository) (*PricingService, error) {
	svc := &PricingService{repo: repo}
	if err := svc.RefreshRules(context.Background()); err != nil {
		return nil, err
	}
	return svc, nil
}

// RefreshRules reloads rules from database
func (s *PricingService) RefreshRules(ctx context.Context) error {
	rules, err := s.repo.GetActiveRules(ctx)
	if err != nil {
		return err
	}

	engineRules := repository.ToEngineRules(rules)
	newEngine, err := engine.NewRulesEngine(engineRules)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.engine = newEngine
	s.mu.Unlock()

	return nil
}

// CalculatePriceRequest represents a pricing calculation request
type CalculatePriceRequest struct {
	TripID         string
	SeatClass      string
	Date           string
	Quantity       int
	BasePricePaisa int64
	OccupancyRate  float64
}

// CalculatePriceResponse represents a pricing calculation response
type CalculatePriceResponse struct {
	FinalPricePaisa int64
	BasePricePaisa  int64
	AppliedRules    []engine.AppliedRule
}

// CalculatePrice calculates the final price by applying all matching rules
func (s *PricingService) CalculatePrice(ctx context.Context, req *CalculatePriceRequest) (*CalculatePriceResponse, error) {
	s.mu.RLock()
	eng := s.engine
	s.mu.RUnlock()

	if eng == nil {
		return &CalculatePriceResponse{
			FinalPricePaisa: req.BasePricePaisa,
			BasePricePaisa:  req.BasePricePaisa,
		}, nil
	}

	env := engine.CreateEnvironment(req.SeatClass, req.Date, req.Quantity, req.OccupancyRate)
	finalPrice, appliedRules, err := eng.Evaluate(ctx, req.BasePricePaisa, env)
	if err != nil {
		return nil, err
	}

	return &CalculatePriceResponse{
		FinalPricePaisa: finalPrice,
		BasePricePaisa:  req.BasePricePaisa,
		AppliedRules:    appliedRules,
	}, nil
}

// GetRules returns all pricing rules
func (s *PricingService) GetRules(ctx context.Context, includeInactive bool) ([]*repository.PricingRule, error) {
	return s.repo.GetAllRules(ctx, includeInactive)
}

// CreateRule creates a new pricing rule
func (s *PricingService) CreateRule(ctx context.Context, rule *repository.PricingRule) error {
	rule.IsActive = true
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return err
	}
	return s.RefreshRules(ctx)
}

// UpdateRule updates a pricing rule
func (s *PricingService) UpdateRule(ctx context.Context, rule *repository.PricingRule) error {
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return err
	}
	return s.RefreshRules(ctx)
}

// DeleteRule deletes a pricing rule
func (s *PricingService) DeleteRule(ctx context.Context, id string) error {
	if err := s.repo.DeleteRule(ctx, id); err != nil {
		return err
	}
	return s.RefreshRules(ctx)
}

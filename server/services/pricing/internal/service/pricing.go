package service

import (
	"context"
	"sync"

	"github.com/MuhibNayem/Travio/server/services/pricing/internal/engine"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
)

// PricingService handles pricing calculations
type PricingService struct {
	repo         *repository.PostgresRepository
	globalEngine *engine.RulesEngine
	orgEngines   map[string]*engine.RulesEngine
	mu           sync.RWMutex
}

// NewPricingService creates a new pricing service
func NewPricingService(repo *repository.PostgresRepository) (*PricingService, error) {
	svc := &PricingService{
		repo:       repo,
		orgEngines: make(map[string]*engine.RulesEngine),
	}
	if err := svc.RefreshRules(context.Background()); err != nil {
		return nil, err
	}
	return svc, nil
}

// RefreshRules reloads rules from database and partitions them into engines
func (s *PricingService) RefreshRules(ctx context.Context) error {
	rules, err := s.repo.GetActiveRules(ctx)
	if err != nil {
		return err
	}

	globalRules := make([]*repository.PricingRule, 0)
	orgRulesMap := make(map[string][]*repository.PricingRule)

	// Partition rules
	for _, r := range rules {
		if r.OrganizationID == nil || *r.OrganizationID == "" {
			globalRules = append(globalRules, r)
		} else {
			orgID := *r.OrganizationID
			orgRulesMap[orgID] = append(orgRulesMap[orgID], r)
		}
	}

	// Build Global Engine
	newGlobalEngine, err := engine.NewRulesEngine(repository.ToEngineRules(globalRules))
	if err != nil {
		return err
	}

	// Build Org Engines (with merged overrides)
	newOrgEngines := make(map[string]*engine.RulesEngine)

	for orgID, specificRules := range orgRulesMap {
		// Start with global rules
		mergedRules := make([]*repository.PricingRule, len(globalRules))
		copy(mergedRules, globalRules)

		// Create map of specific rules by name for O(1) override check
		specificMap := make(map[string]*repository.PricingRule)
		for _, r := range specificRules {
			specificMap[r.Name] = r
		}

		// Apply Overrides: Replace global rule if name matches specific rule
		for i, gRule := range mergedRules {
			if override, exists := specificMap[gRule.Name]; exists {
				mergedRules[i] = override       // Replace global with override
				delete(specificMap, gRule.Name) // Mark as used
			}
		}

		// Append remaining unique operator rules (additions)
		for _, r := range specificMap {
			mergedRules = append(mergedRules, r)
		}

		eng, err := engine.NewRulesEngine(repository.ToEngineRules(mergedRules))
		if err != nil {
			// Log error but continue for other orgs? For now return error strict.
			return err
		}
		newOrgEngines[orgID] = eng
	}

	s.mu.Lock()
	s.globalEngine = newGlobalEngine
	s.orgEngines = newOrgEngines
	s.mu.Unlock()

	return nil
}

// CalculatePriceRequest represents a pricing calculation request
type CalculatePriceRequest struct {
	TripID         string
	SeatClass      string
	SeatCategory   string
	Date           string
	Quantity       int
	BasePricePaisa int64
	OccupancyRate  float64
	OrganizationID string // New field
	DepartureTime  int64
	RouteID        string
	ScheduleID     string
	FromStationID  string
	ToStationID    string
	VehicleType    string
	VehicleClass   string
	PromoCode      string
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
	// Use Org-Specific Engine if available, otherwise fallback to Global
	eng := s.globalEngine
	if req.OrganizationID != "" {
		if orgEng, exists := s.orgEngines[req.OrganizationID]; exists {
			eng = orgEng
		}
		// Note: If OrganizationID is provided but no Specific rules exist, we use Global Engine.
		// Since RefreshRules builds OrgEngines ONLY if they have specific rules,
		// we safely fallback to Global here.
		// Wait - logic check: If Org has NO specific rules, they adhere to Global rules.
		// Correct.
	}
	s.mu.RUnlock()

	if eng == nil {
		return &CalculatePriceResponse{
			FinalPricePaisa: req.BasePricePaisa,
			BasePricePaisa:  req.BasePricePaisa,
		}, nil
	}

	env := engine.CreateEnvironment(engine.EnvironmentParams{
		SeatClass:     req.SeatClass,
		SeatCategory:  req.SeatCategory,
		Date:          req.Date,
		Quantity:      req.Quantity,
		OccupancyRate: req.OccupancyRate,
		TripID:        req.TripID,
		RouteID:       req.RouteID,
		ScheduleID:    req.ScheduleID,
		FromStationID: req.FromStationID,
		ToStationID:   req.ToStationID,
		VehicleType:   req.VehicleType,
		VehicleClass:  req.VehicleClass,
		PromoCode:     req.PromoCode,
		DepartureTime: req.DepartureTime,
	})
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
func (s *PricingService) GetRules(ctx context.Context, includeInactive bool, organizationID string) ([]*repository.PricingRule, error) {
	return s.repo.GetAllRules(ctx, includeInactive, organizationID)
}

// CreateRule creates a new pricing rule
func (s *PricingService) CreateRule(ctx context.Context, rule *repository.PricingRule) error {
	rule.IsActive = true
	if rule.AdjustmentType == "" {
		rule.AdjustmentType = "multiplier"
	}
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return err
	}
	return s.RefreshRules(ctx)
}

// UpdateRule updates a pricing rule
func (s *PricingService) UpdateRule(ctx context.Context, rule *repository.PricingRule) error {
	if rule.AdjustmentType == "" {
		rule.AdjustmentType = "multiplier"
	}
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

package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MuhibNayem/Travio/server/services/pricing/internal/engine"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
	"github.com/google/uuid"
)

// PricingService handles pricing calculations
type PricingService struct {
	repo         *repository.PostgresRepository
	redisRepo    *repository.RedisRepository
	globalEngine *engine.RulesEngine
	orgEngines   map[string]*engine.RulesEngine
	mu           sync.RWMutex
}

// NewPricingService creates a new pricing service
func NewPricingService(repo *repository.PostgresRepository, redisRepo *repository.RedisRepository) (*PricingService, error) {
	svc := &PricingService{
		repo:       repo,
		redisRepo:  redisRepo,
		orgEngines: make(map[string]*engine.RulesEngine),
	}
	if err := svc.RefreshRules(context.Background()); err != nil {
		return nil, err // TODO: Should we fail if DB is down? Yes.
	}
	return svc, nil
}

// RefreshRules reloads rules from database and partitions them into engines
func (s *PricingService) RefreshRules(ctx context.Context) error {
	// Try Redis Cache first for active rules
	rules, err := s.redisRepo.GetCachedRules(ctx, "active:all")
	if err != nil || rules == nil {
		// Cache Miss or Error -> DB Fallback
		rules, err = s.repo.GetActiveRules(ctx)
		if err != nil {
			return err
		}
		// Populate Cache
		go func() {
			_ = s.redisRepo.CacheRules(context.Background(), "active:all", rules, 1*time.Hour)
		}()
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
	FinalPricePaisa  int64
	BasePricePaisa   int64
	AppliedRules     []engine.AppliedRule
	PromotionApplied *PromotionApplied
}

type PromotionApplied struct {
	Code                string
	DiscountAmountPaisa int64
}

// CalculatePrice calculates the final price by applying all matching rules
func (s *PricingService) CalculatePrice(ctx context.Context, req *CalculatePriceRequest) (*CalculatePriceResponse, error) {
	s.mu.RLock()
	eng := s.globalEngine
	if req.OrganizationID != "" {
		if orgEng, exists := s.orgEngines[req.OrganizationID]; exists {
			eng = orgEng
		}
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

	evaluatedPrice, appliedRules, err := eng.Evaluate(ctx, req.BasePricePaisa, env)
	if err != nil {
		return nil, err
	}

	finalPrice := evaluatedPrice
	var promoApplied *PromotionApplied

	// Apply Promotion logic
	if req.PromoCode != "" {
		promo, err := s.repo.GetPromotionByCode(ctx, req.PromoCode, req.OrganizationID)
		// Check validity
		if err == nil && promo != nil && promo.IsActive {
			valid := true
			now := time.Now()

			if promo.ValidFrom != nil && now.Before(*promo.ValidFrom) {
				valid = false
			}
			if promo.ValidUntil != nil && now.After(*promo.ValidUntil) {
				valid = false
			}
			if promo.MaxUsage > 0 && promo.CurrentUsage >= promo.MaxUsage {
				valid = false
			}
			if promo.MinOrderAmountPaisa > 0 && finalPrice < promo.MinOrderAmountPaisa {
				valid = false
			}

			if valid {
				var discountAmount int64
				if promo.DiscountType == "PERCENT" {
					discountAmount = int64(float64(finalPrice) * (promo.DiscountValue / 100.0))
				} else if promo.DiscountType == "FIXED" {
					discountAmount = int64(promo.DiscountValue)
				}

				if discountAmount > finalPrice {
					discountAmount = finalPrice
				}
				finalPrice -= discountAmount

				promoApplied = &PromotionApplied{
					Code:                promo.Code,
					DiscountAmountPaisa: discountAmount,
				}
			}
		}
	}

	return &CalculatePriceResponse{
		FinalPricePaisa:  finalPrice,
		BasePricePaisa:   req.BasePricePaisa,
		AppliedRules:     appliedRules,
		PromotionApplied: promoApplied,
	}, nil
}

// GetRules returns all pricing rules
func (s *PricingService) GetRules(ctx context.Context, includeInactive bool, organizationID string) ([]*repository.PricingRule, error) {
	cacheKey := fmt.Sprintf("%s:%t", organizationID, includeInactive)
	rules, err := s.redisRepo.GetCachedRules(ctx, cacheKey)
	if err == nil && rules != nil {
		return rules, nil
	}

	rules, err = s.repo.GetAllRules(ctx, includeInactive, organizationID)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = s.redisRepo.CacheRules(context.Background(), cacheKey, rules, 1*time.Hour)
	}()
	return rules, nil
}

// CreateRule creates a new pricing rule
func (s *PricingService) CreateRule(ctx context.Context, rule *repository.PricingRule) error {
	if err := s.ValidateRule(ctx, rule); err != nil {
		return err
	}
	rule.IsActive = true
	if rule.AdjustmentType == "" {
		rule.AdjustmentType = "multiplier"
	}
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return err
	}
	s.invalidateCache(ctx, rule)
	return s.RefreshRules(ctx)
}

// UpdateRule updates a pricing rule
func (s *PricingService) UpdateRule(ctx context.Context, rule *repository.PricingRule) error {
	if err := s.ValidateRule(ctx, rule); err != nil {
		return err
	}
	if rule.AdjustmentType == "" {
		rule.AdjustmentType = "multiplier"
	}
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return err
	}
	s.invalidateCache(ctx, rule)
	return s.RefreshRules(ctx)
}

// DeleteRule deletes a pricing rule
func (s *PricingService) DeleteRule(ctx context.Context, id string) error {
	// Need to fetch rule to know OrgID for invalidation?
	// Or just invalidate global/all?
	// For efficiency, mostly just Global Active matters.
	// But let's try to do it right.
	// Allow partial invalidation if we don't have rule?
	// Let's just invalidate active:all for now, as that's critical.
	// To do org-specific, we'd need to read before delete.

	if err := s.repo.DeleteRule(ctx, id); err != nil {
		return err
	}
	_ = s.redisRepo.InvalidateRules(ctx, "active:all")
	return s.RefreshRules(ctx)
}

func (s *PricingService) invalidateCache(ctx context.Context, rule *repository.PricingRule) {
	_ = s.redisRepo.InvalidateRules(ctx, "active:all")
	orgID := ""
	if rule.OrganizationID != nil {
		orgID = *rule.OrganizationID
	}
	_ = s.redisRepo.InvalidateRules(ctx, fmt.Sprintf("%s:true", orgID))
	_ = s.redisRepo.InvalidateRules(ctx, fmt.Sprintf("%s:false", orgID))
}

func (s *PricingService) ValidateRule(ctx context.Context, rule *repository.PricingRule) error {
	// 1. Date Validation
	if rule.ValidFrom != nil && rule.ValidTo != nil {
		if rule.ValidFrom.After(*rule.ValidTo) {
			return errors.New("valid_from cannot be after valid_to")
		}
	}

	// 2. Overlap/Duplicate Detection (Block identical conditions for same Org)
	// Fetch all rules for this org (or global)
	orgID := ""
	if rule.OrganizationID != nil {
		orgID = *rule.OrganizationID
	}

	existingRules, err := s.repo.GetAllRules(ctx, true, orgID)
	if err != nil {
		return err
	}

	for _, existing := range existingRules {
		// Skip self (for Update)
		if existing.ID == rule.ID {
			continue
		}

		// Check for identical condition
		if existing.Condition == rule.Condition {
			return fmt.Errorf("duplicate rule condition detected: same condition already exists in rule %s", existing.Name)
		}
	}

	return nil
}

// --- Promotions ---

func (s *PricingService) CreatePromotion(ctx context.Context, promo *repository.Promotion) error {
	// 1. Basic Validation
	if promo.ID == "" {
		promo.ID = uuid.New().String()
	}
	if promo.Code == "" {
		return errors.New("promotion code is required")
	}
	if promo.DiscountValue <= 0 {
		return errors.New("discount value must be positive")
	}

	if err := s.repo.CreatePromotion(ctx, promo); err != nil {
		return err
	}
	return nil
}

func (s *PricingService) GetPromotions(ctx context.Context, orgID string, activeOnly bool) ([]*repository.Promotion, error) {
	return s.repo.GetPromotions(ctx, orgID, activeOnly)
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

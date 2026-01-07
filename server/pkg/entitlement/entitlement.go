// Package entitlement provides subscription enforcement for Travio services.
// It implements a cache-first entitlement checking system with Redis caching
// and gRPC interceptors for service-to-service enforcement.
package entitlement

import (
	"context"
	"time"
)

// Entitlements represents the subscription state and feature limits for an organization.
type Entitlements struct {
	OrganizationID  string            `json:"organization_id"`
	PlanID          string            `json:"plan_id"`
	PlanName        string            `json:"plan_name"`
	Status          string            `json:"status"` // active, past_due, canceled, trialing
	Features        map[string]string `json:"features"`
	UsageThisPeriod map[string]int64  `json:"usage_this_period"`
	QuotaLimits     map[string]int64  `json:"quota_limits"`
	PeriodStart     time.Time         `json:"period_start"`
	PeriodEnd       time.Time         `json:"period_end"`
	CachedAt        time.Time         `json:"cached_at"`
}

// FeatureKey constants for standard features
const (
	FeatureMaxTripsPerMonth    = "max_trips_per_month"
	FeatureMaxBookingsPerMonth = "max_bookings_per_month"
	FeatureMaxAdmins           = "max_admins"
	FeaturePrioritySupport     = "priority_support"
	FeatureAPIRateLimit        = "api_rate_limit"
	FeatureCustomBranding      = "custom_branding"
	FeatureAdvancedAnalytics   = "advanced_analytics"
)

// EntitlementChecker is the interface for checking organization entitlements.
type EntitlementChecker interface {
	// CheckEntitlement retrieves the entitlements for an organization.
	// Returns nil if the organization has no active subscription.
	CheckEntitlement(ctx context.Context, orgID string) (*Entitlements, error)

	// HasFeature checks if the organization has access to a specific feature.
	HasFeature(ctx context.Context, orgID, featureKey string) (bool, error)

	// CheckQuota verifies if the organization is within quota for a given metric.
	// Returns (allowed, remaining, error).
	CheckQuota(ctx context.Context, orgID, quotaKey string) (bool, int64, error)

	// InvalidateCache removes the cached entitlement for an organization.
	InvalidateCache(ctx context.Context, orgID string) error
}

// CheckResult represents the result of an entitlement check.
type CheckResult struct {
	Allowed   bool   `json:"allowed"`
	Reason    string `json:"reason,omitempty"`
	Remaining int64  `json:"remaining,omitempty"`
}

// IsActive returns true if the subscription status allows service usage.
func (e *Entitlements) IsActive() bool {
	return e.Status == "active" || e.Status == "trialing"
}

// GetFeature returns the value of a feature, or empty string if not found.
func (e *Entitlements) GetFeature(key string) string {
	if e.Features == nil {
		return ""
	}
	return e.Features[key]
}

// GetQuotaLimit returns the quota limit for a key, or 0 if not set.
func (e *Entitlements) GetQuotaLimit(key string) int64 {
	if e.QuotaLimits == nil {
		return 0
	}
	return e.QuotaLimits[key]
}

// GetUsage returns the current usage for a key, or 0 if not tracked.
func (e *Entitlements) GetUsage(key string) int64 {
	if e.UsageThisPeriod == nil {
		return 0
	}
	return e.UsageThisPeriod[key]
}

// IsWithinQuota checks if current usage is within the quota limit.
// Returns (allowed, remaining).
func (e *Entitlements) IsWithinQuota(key string) (bool, int64) {
	limit := e.GetQuotaLimit(key)
	if limit == 0 {
		// No limit set, allow unlimited
		return true, -1
	}
	usage := e.GetUsage(key)
	remaining := limit - usage
	return remaining > 0, remaining
}

// Config holds configuration for the entitlement checker.
type Config struct {
	// Enabled is the master toggle for enforcement.
	Enabled bool `json:"enabled"`

	// FailOpen determines behavior when entitlement check fails.
	// If true, requests are allowed on failure (availability > correctness).
	// If false, requests are denied on failure (correctness > availability).
	FailOpen bool `json:"fail_open"`

	// CacheTTL is the duration to cache entitlements in Redis.
	CacheTTL time.Duration `json:"cache_ttl"`

	// SubscriptionServiceAddr is the gRPC address of the subscription service.
	SubscriptionServiceAddr string `json:"subscription_service_addr"`

	// RedisAddr is the Redis connection address.
	RedisAddr string `json:"redis_addr"`
}

// DefaultConfig returns sensible defaults for production.
func DefaultConfig() Config {
	return Config{
		Enabled:                 true,
		FailOpen:                true,
		CacheTTL:                5 * time.Minute,
		SubscriptionServiceAddr: "localhost:50060",
		RedisAddr:               "localhost:6379",
	}
}

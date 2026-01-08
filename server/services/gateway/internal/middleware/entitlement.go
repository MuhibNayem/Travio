package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// EntitlementMiddleware enforces subscription entitlements for HTTP requests
type EntitlementMiddleware struct {
	checker     entitlement.EntitlementChecker
	enabled     bool
	failOpen    bool
	skipPaths   []string
	requireAuth bool // If true, require active subscription
}

// NewEntitlementMiddleware creates a new entitlement enforcement middleware
func NewEntitlementMiddleware(redisAddr string) *EntitlementMiddleware {
	enabled := os.Getenv("ENTITLEMENT_ENABLED") != "false"
	failOpen := os.Getenv("ENTITLEMENT_FAIL_OPEN") != "false" // Default true

	subscriptionAddr := os.Getenv("SUBSCRIPTION_URL")
	if subscriptionAddr == "" {
		subscriptionAddr = "localhost:50060"
	}

	// Skip paths that don't require entitlement checks
	skipPaths := []string{
		"/health", "/ready", "/v1/auth", "/v1/plans", "/v1/public",
	}

	if !enabled {
		logger.Info("Entitlement enforcement disabled")
		return &EntitlementMiddleware{enabled: false}
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Connect to Subscription service
	fetcher, err := entitlement.NewSubscriptionFetcher(subscriptionAddr)
	if err != nil {
		logger.Warn("Failed to connect to subscription service for entitlement", "error", err)
		return &EntitlementMiddleware{enabled: false, failOpen: true}
	}

	cfg := entitlement.Config{
		Enabled:                 true,
		FailOpen:                failOpen,
		CacheTTL:                5 * time.Minute,
		SubscriptionServiceAddr: subscriptionAddr,
		RedisAddr:               redisAddr,
	}

	checker := entitlement.NewCachedChecker(redisClient, fetcher, cfg)

	logger.Info("Entitlement enforcement enabled for Gateway")

	return &EntitlementMiddleware{
		checker:     checker,
		enabled:     true,
		failOpen:    failOpen,
		skipPaths:   skipPaths,
		requireAuth: true,
	}
}

// Middleware returns an HTTP middleware that enforces subscription entitlements
func (e *EntitlementMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if enforcement disabled
		if !e.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip certain paths
		for _, path := range e.skipPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Extract organization ID from request context (set by auth middleware)
		orgID := r.Header.Get("X-Organization-ID")
		if orgID == "" {
			// Try to get from JWT claims in context
			if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
				if id, ok := claims["organization_id"].(string); ok {
					orgID = id
				}
			}
		}

		// If no org ID, skip check (might be public endpoint or user without org)
		if orgID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check entitlement
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		ent, err := e.checker.CheckEntitlement(ctx, orgID)
		if err != nil {
			if e.failOpen {
				logger.Warn("Entitlement check failed, allowing request (fail-open)",
					"org_id", orgID, "error", err)
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Subscription verification failed", http.StatusInternalServerError)
			return
		}

		if ent == nil {
			http.Error(w, "No active subscription found", http.StatusPaymentRequired)
			return
		}

		if !ent.IsActive() {
			http.Error(w, "Subscription is not active", http.StatusPaymentRequired)
			return
		}

		// Attach entitlement info to context for downstream handlers
		ctx = context.WithValue(r.Context(), "entitlement", ent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

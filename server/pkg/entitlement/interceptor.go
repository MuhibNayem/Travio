package entitlement

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// MetadataKeyOrgID is the metadata key for organization ID.
	MetadataKeyOrgID = "x-organization-id"
)

// InterceptorConfig holds configuration for the entitlement interceptor.
type InterceptorConfig struct {
	// Checker is the entitlement checker to use.
	Checker EntitlementChecker

	// SkipMethods is a list of method names to skip entitlement checks.
	// Format: "/package.Service/Method" or just "Method"
	SkipMethods []string

	// RequireActiveSubscription determines if an active subscription is required.
	RequireActiveSubscription bool

	// QuotaKey is the quota key to check for each request (optional).
	// If set, quota will be checked and requests will be rejected if quota exceeded.
	QuotaKey string
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for entitlement checks.
func UnaryServerInterceptor(cfg InterceptorConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if method should be skipped
		if shouldSkipMethod(info.FullMethod, cfg.SkipMethods) {
			return handler(ctx, req)
		}

		// Extract organization ID from metadata
		orgID, err := extractOrgID(ctx)
		if err != nil {
			// No org ID in context - this might be a public endpoint
			return handler(ctx, req)
		}

		// Check entitlement
		if err := checkEntitlementForRequest(ctx, cfg, orgID); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor for entitlement checks.
func StreamServerInterceptor(cfg InterceptorConfig) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Check if method should be skipped
		if shouldSkipMethod(info.FullMethod, cfg.SkipMethods) {
			return handler(srv, ss)
		}

		// Extract organization ID from metadata
		ctx := ss.Context()
		orgID, err := extractOrgID(ctx)
		if err != nil {
			return handler(srv, ss)
		}

		// Check entitlement
		if err := checkEntitlementForRequest(ctx, cfg, orgID); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// shouldSkipMethod checks if the method should bypass entitlement checks.
func shouldSkipMethod(fullMethod string, skipMethods []string) bool {
	for _, skip := range skipMethods {
		if strings.HasSuffix(fullMethod, skip) || fullMethod == skip {
			return true
		}
	}
	return false
}

// extractOrgID extracts the organization ID from gRPC metadata.
func extractOrgID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "missing metadata")
	}

	values := md.Get(MetadataKeyOrgID)
	if len(values) == 0 {
		return "", status.Error(codes.InvalidArgument, "missing organization ID")
	}

	return values[0], nil
}

// checkEntitlementForRequest performs the entitlement check and returns an error if denied.
func checkEntitlementForRequest(ctx context.Context, cfg InterceptorConfig, orgID string) error {
	if cfg.Checker == nil {
		return nil // No checker configured, allow
	}

	ent, err := cfg.Checker.CheckEntitlement(ctx, orgID)
	if err != nil {
		// Error during check - behavior depends on fail-open config
		// The CachedChecker handles fail-open internally
		return status.Errorf(codes.Internal, "entitlement check failed: %v", err)
	}

	if ent == nil {
		return status.Error(codes.PermissionDenied, "no subscription found")
	}

	// Check subscription status
	if cfg.RequireActiveSubscription && !ent.IsActive() {
		return status.Errorf(codes.PermissionDenied,
			"subscription is %s, active subscription required", ent.Status)
	}

	// Check quota if configured
	if cfg.QuotaKey != "" {
		allowed, remaining, err := cfg.Checker.CheckQuota(ctx, orgID, cfg.QuotaKey)
		if err != nil {
			return status.Errorf(codes.Internal, "quota check failed: %v", err)
		}
		if !allowed {
			return status.Errorf(codes.ResourceExhausted,
				"quota exceeded for %s (remaining: %d)", cfg.QuotaKey, remaining)
		}
	}

	return nil
}

// WithOrgID adds the organization ID to the outgoing gRPC context.
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetadataKeyOrgID, orgID)
}

// GetOrgIDFromContext extracts the organization ID from the context.
// This is useful for handlers that need to know the org ID.
func GetOrgIDFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(MetadataKeyOrgID)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Metadata keys for auth propagation
const (
	UserIDKey    = "x-user-id"
	OrgIDKey     = "x-org-id"
	UserRoleKey  = "x-user-role"
	RequestIDKey = "x-request-id"
)

// UserContext holds user information extracted from JWT
type UserContext struct {
	UserID    string
	OrgID     string
	Role      string
	RequestID string
}

// UnaryClientInterceptor propagates user context via gRPC metadata
func UnaryClientInterceptor(userCtxKey interface{}) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Extract user context from Go context
		if uc, ok := ctx.Value(userCtxKey).(*UserContext); ok && uc != nil {
			md := metadata.Pairs(
				UserIDKey, uc.UserID,
				OrgIDKey, uc.OrgID,
				UserRoleKey, uc.Role,
				RequestIDKey, uc.RequestID,
			)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInterceptor extracts user context from gRPC metadata
func UnaryServerInterceptor(userCtxKey interface{}) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			uc := &UserContext{}
			if vals := md.Get(UserIDKey); len(vals) > 0 {
				uc.UserID = vals[0]
			}
			if vals := md.Get(OrgIDKey); len(vals) > 0 {
				uc.OrgID = vals[0]
			}
			if vals := md.Get(UserRoleKey); len(vals) > 0 {
				uc.Role = vals[0]
			}
			if vals := md.Get(RequestIDKey); len(vals) > 0 {
				uc.RequestID = vals[0]
			}
			ctx = context.WithValue(ctx, userCtxKey, uc)
		}
		return handler(ctx, req)
	}
}

// StreamClientInterceptor propagates user context for streaming RPCs
func StreamClientInterceptor(userCtxKey interface{}) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if uc, ok := ctx.Value(userCtxKey).(*UserContext); ok && uc != nil {
			md := metadata.Pairs(
				UserIDKey, uc.UserID,
				OrgIDKey, uc.OrgID,
				UserRoleKey, uc.Role,
				RequestIDKey, uc.RequestID,
			)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// GetUserFromContext extracts UserContext from context
func GetUserFromContext(ctx context.Context, key interface{}) *UserContext {
	if uc, ok := ctx.Value(key).(*UserContext); ok {
		return uc
	}
	return nil
}

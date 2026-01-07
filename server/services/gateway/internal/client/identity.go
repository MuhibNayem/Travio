package client

import (
	"context"
	"time"

	identityv1 "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// IdentityClient wraps the gRPC identity client
type IdentityClient struct {
	conn   *grpc.ClientConn
	client identityv1.IdentityServiceClient
}

// NewIdentityClient creates a new gRPC client for the identity service
// Uses mTLS if TLS config is provided
func NewIdentityClient(address string, tlsCfg TLSConfig) (*IdentityClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to identity service", "address", address, "tls", tlsCfg.CertFile != "")
	return &IdentityClient{
		conn:   conn,
		client: identityv1.NewIdentityServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *IdentityClient) Close() error {
	return c.conn.Close()
}

// Register creates a new user account
func (c *IdentityClient) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	return c.client.Register(ctx, req)
}

// Login authenticates a user
func (c *IdentityClient) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	return c.client.Login(ctx, req)
}

// RefreshToken refreshes an access token
func (c *IdentityClient) RefreshToken(ctx context.Context, req *identityv1.RefreshTokenRequest) (*identityv1.RefreshTokenResponse, error) {
	return c.client.RefreshToken(ctx, req)
}

// Logout invalidates a refresh token
func (c *IdentityClient) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	return c.client.Logout(ctx, req)
}

// CreateOrganization creates a new organization
func (c *IdentityClient) CreateOrganization(ctx context.Context, req *identityv1.CreateOrgRequest) (*identityv1.CreateOrgResponse, error) {
	return c.client.CreateOrganization(ctx, req)
}

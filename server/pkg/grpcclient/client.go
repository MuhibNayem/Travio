package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address string
	Timeout time.Duration
}

// Dial creates a new gRPC client connection with standard timeouts and options
func Dial(cfg Config) (*grpc.ClientConn, error) {
	logger.Info("Dialing gRPC service", "address", cfg.Address)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// In a real env, we would add TLS credentials here, Interceptors for Trace Propagation, etc.
	conn, err := grpc.DialContext(ctx, cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait for connection to be active
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", cfg.Address, err)
	}

	return conn, nil
}

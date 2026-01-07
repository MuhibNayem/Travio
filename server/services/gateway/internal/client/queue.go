package client

import (
	"context"
	"time"

	queuev1 "github.com/MuhibNayem/Travio/server/api/proto/queue/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// QueueClient wraps the gRPC queue client
type QueueClient struct {
	conn   *grpc.ClientConn
	client queuev1.QueueServiceClient
}

// NewQueueClient creates a new gRPC client for the queue service
// Uses mTLS if TLS config is provided
func NewQueueClient(address string, tlsCfg TLSConfig) (*QueueClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to queue service", "address", address, "tls", tlsCfg.CertFile != "")
	return &QueueClient{
		conn:   conn,
		client: queuev1.NewQueueServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *QueueClient) Close() error {
	return c.conn.Close()
}

// JoinQueue adds a user to the queue
func (c *QueueClient) JoinQueue(ctx context.Context, eventID, userID, sessionID string) (*queuev1.QueuePosition, error) {
	return c.client.JoinQueue(ctx, &queuev1.JoinQueueRequest{
		EventId:   eventID,
		UserId:    userID,
		SessionId: sessionID,
	})
}

// GetPosition returns user's current queue position
func (c *QueueClient) GetPosition(ctx context.Context, eventID, userID string) (*queuev1.QueuePosition, error) {
	return c.client.GetPosition(ctx, &queuev1.GetPositionRequest{
		EventId: eventID,
		UserId:  userID,
	})
}

// ValidateToken validates an admission token
func (c *QueueClient) ValidateToken(ctx context.Context, token string) (*queuev1.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, &queuev1.ValidateTokenRequest{Token: token})
}

// GetQueueStats returns queue statistics
func (c *QueueClient) GetQueueStats(ctx context.Context, eventID string) (*queuev1.QueueStats, error) {
	return c.client.GetQueueStats(ctx, &queuev1.GetQueueStatsRequest{EventId: eventID})
}

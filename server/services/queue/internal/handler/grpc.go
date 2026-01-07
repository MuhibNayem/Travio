package handler

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/queue/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/service"
)

// GrpcHandler handles gRPC requests for queue service
// Note: This will be fully wired once queue.proto is generated with protoc
type GrpcHandler struct {
	svc *service.QueueService
}

// NewGrpcHandler creates a new gRPC handler
func NewGrpcHandler(svc *service.QueueService) *GrpcHandler {
	return &GrpcHandler{svc: svc}
}

// GrpcJoinRequest is the request type for JoinQueue
type GrpcJoinRequest struct {
	EventID   string
	UserID    string
	SessionID string
}

// GrpcPositionResponse is the response type for queue position
type GrpcPositionResponse struct {
	Position      int32
	EstimatedWait int32
	Token         string
	Status        string
}

// JoinQueue adds a user to the virtual waiting room
func (h *GrpcHandler) JoinQueue(ctx context.Context, req *GrpcJoinRequest) (*GrpcPositionResponse, error) {
	entry, err := h.svc.JoinQueue(ctx, req.EventID, req.UserID, req.SessionID)
	if err != nil {
		return nil, err
	}

	return &GrpcPositionResponse{
		Position:      int32(entry.Position),
		EstimatedWait: int32(entry.EstimatedWait.Seconds()),
		Token:         entry.Token,
		Status:        string(entry.Status),
	}, nil
}

// GrpcGetPositionRequest is the request type for GetPosition
type GrpcGetPositionRequest struct {
	EventID string
	UserID  string
}

// GetPosition returns user's current queue position
func (h *GrpcHandler) GetPosition(ctx context.Context, req *GrpcGetPositionRequest) (*GrpcPositionResponse, error) {
	entry, err := h.svc.GetPosition(ctx, req.EventID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &GrpcPositionResponse{
		Position:      int32(entry.Position),
		EstimatedWait: int32(entry.EstimatedWait.Seconds()),
		Token:         entry.Token,
		Status:        string(entry.Status),
	}, nil
}

// GrpcLeaveRequest is the request type for LeaveQueue
type GrpcLeaveRequest struct {
	EventID string
	UserID  string
}

// LeaveQueue removes a user from the queue
func (h *GrpcHandler) LeaveQueue(ctx context.Context, req *GrpcLeaveRequest) (bool, error) {
	err := h.svc.LeaveQueue(ctx, req.EventID, req.UserID)
	return err == nil, err
}

// GrpcValidateTokenRequest is the request type for ValidateToken
type GrpcValidateTokenRequest struct {
	Token string
}

// GrpcValidateTokenResponse is the response for token validation
type GrpcValidateTokenResponse struct {
	Valid   bool
	UserID  string
	EventID string
}

// ValidateToken checks if an admission token is valid
func (h *GrpcHandler) ValidateToken(ctx context.Context, req *GrpcValidateTokenRequest) (*GrpcValidateTokenResponse, error) {
	valid, userID, eventID, err := h.svc.ValidateAdmission(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &GrpcValidateTokenResponse{
		Valid:   valid,
		UserID:  userID,
		EventID: eventID,
	}, nil
}

// ConsumeToken marks a token as used
func (h *GrpcHandler) ConsumeToken(ctx context.Context, token string) (bool, error) {
	err := h.svc.CompleteAdmission(ctx, token)
	return err == nil, err
}

// QueueStatsResponse is the response for queue stats
type QueueStatsResponse struct {
	EventID       string
	TotalWaiting  int32
	TotalAdmitted int32
	AvgWaitSecs   int32
	AdmissionRate int32
}

// GetQueueStats returns queue statistics
func (h *GrpcHandler) GetQueueStats(ctx context.Context, eventID string) (*QueueStatsResponse, error) {
	stats, err := h.svc.GetStats(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return &QueueStatsResponse{
		EventID:       stats.EventID,
		TotalWaiting:  int32(stats.TotalWaiting),
		TotalAdmitted: int32(stats.TotalAdmitted),
		AvgWaitSecs:   int32(stats.AvgWaitTime.Seconds()),
		AdmissionRate: int32(stats.AdmissionRate),
	}, nil
}

// GrpcConfigureRequest is the request for queue configuration
type GrpcConfigureRequest struct {
	EventID       string
	MaxConcurrent int32
	BatchSize     int32
	TokenTTLSecs  int32
	Enabled       bool
}

// ConfigureQueue sets queue configuration
func (h *GrpcHandler) ConfigureQueue(ctx context.Context, req *GrpcConfigureRequest) (bool, error) {
	config := &domain.AdmissionConfig{
		EventID:            req.EventID,
		MaxConcurrent:      int(req.MaxConcurrent),
		AdmissionBatchSize: int(req.BatchSize),
		TokenTTL:           time.Duration(req.TokenTTLSecs) * time.Second,
		QueueEnabled:       req.Enabled,
	}

	err := h.svc.ConfigureQueue(ctx, config)
	return err == nil, err
}

package handler

import (
	"context"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/queue/v1"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcHandler handles gRPC requests for queue service
type GrpcHandler struct {
	pb.UnimplementedQueueServiceServer
	svc *service.QueueService
}

// NewGrpcHandler creates a new gRPC handler
func NewGrpcHandler(svc *service.QueueService) *GrpcHandler {
	return &GrpcHandler{svc: svc}
}

// JoinQueue adds a user to the virtual waiting room
func (h *GrpcHandler) JoinQueue(ctx context.Context, req *pb.JoinQueueRequest) (*pb.QueuePosition, error) {
	entry, err := h.svc.JoinQueue(ctx, req.EventId, req.UserId, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.QueuePosition{
		Position:      int32(entry.Position),
		EstimatedWait: int32(entry.EstimatedWait.Seconds()),
		Token:         entry.Token,
		Status:        mapQueueStatus(entry.Status),
	}, nil
}

// GetPosition returns user's current queue position
func (h *GrpcHandler) GetPosition(ctx context.Context, req *pb.GetPositionRequest) (*pb.QueuePosition, error) {
	entry, err := h.svc.GetPosition(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not in queue")
	}

	return &pb.QueuePosition{
		Position:      int32(entry.Position),
		EstimatedWait: int32(entry.EstimatedWait.Seconds()),
		Token:         entry.Token,
		Status:        mapQueueStatus(entry.Status),
	}, nil
}

// LeaveQueue removes a user from the queue
func (h *GrpcHandler) LeaveQueue(ctx context.Context, req *pb.LeaveQueueRequest) (*pb.LeaveQueueResponse, error) {
	err := h.svc.LeaveQueue(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LeaveQueueResponse{Success: true}, nil
}

// ValidateToken checks if an admission token is valid
func (h *GrpcHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	valid, userID, eventID, err := h.svc.ValidateAdmission(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ValidateTokenResponse{
		Valid:   valid,
		UserId:  userID,
		EventId: eventID,
	}, nil
}

// ConsumeToken marks a token as used
func (h *GrpcHandler) ConsumeToken(ctx context.Context, req *pb.ConsumeTokenRequest) (*pb.ConsumeTokenResponse, error) {
	err := h.svc.CompleteAdmission(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ConsumeTokenResponse{Success: true}, nil
}

// GetQueueStats returns queue statistics
func (h *GrpcHandler) GetQueueStats(ctx context.Context, req *pb.GetQueueStatsRequest) (*pb.QueueStats, error) {
	stats, err := h.svc.GetStats(ctx, req.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.QueueStats{
		EventId:       stats.EventID,
		TotalWaiting:  int32(stats.TotalWaiting),
		TotalAdmitted: int32(stats.TotalAdmitted),
		AvgWaitSecs:   int32(stats.AvgWaitTime.Seconds()),
		AdmissionRate: int32(stats.AdmissionRate),
	}, nil
}

// ConfigureQueue sets queue configuration
func (h *GrpcHandler) ConfigureQueue(ctx context.Context, req *pb.ConfigureQueueRequest) (*pb.ConfigureQueueResponse, error) {
	config := &domain.AdmissionConfig{
		EventID:            req.EventId,
		MaxConcurrent:      int(req.MaxConcurrent),
		AdmissionBatchSize: int(req.BatchSize),
		TokenTTL:           time.Duration(req.TokenTtlSecs) * time.Second,
		QueueEnabled:       req.Enabled,
	}

	if err := h.svc.ConfigureQueue(ctx, config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ConfigureQueueResponse{Success: true}, nil
}

func mapQueueStatus(s domain.QueueStatus) pb.QueueStatus {
	switch s {
	case domain.QueueStatusWaiting:
		return pb.QueueStatus_QUEUE_STATUS_WAITING
	case domain.QueueStatusReady:
		return pb.QueueStatus_QUEUE_STATUS_READY
	case domain.QueueStatusExpired:
		return pb.QueueStatus_QUEUE_STATUS_EXPIRED
	case domain.QueueStatusCompleted:
		return pb.QueueStatus_QUEUE_STATUS_COMPLETED
	default:
		return pb.QueueStatus_QUEUE_STATUS_UNSPECIFIED
	}
}

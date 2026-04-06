package handler

import (
	"context"
	"encoding/json"

	pb "github.com/MuhibNayem/Travio/server/api/proto/audit/v1"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GrpcHandler handles gRPC requests for audit operations
type GrpcHandler struct {
	pb.UnimplementedAuditServiceServer
	auditService *service.AuditService
}

// NewGrpcHandler creates a new audit gRPC handler
func NewGrpcHandler(auditService *service.AuditService) *GrpcHandler {
	return &GrpcHandler{
		auditService: auditService,
	}
}

func (h *GrpcHandler) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	log := &domain.AuditLog{
		AggregateID:    req.GetAggregateId(),
		EventType:      req.GetEventType(),
		ResourceType:   req.GetResourceType(),
		Action:         req.GetAction(),
		ActorID:        req.GetActorId(),
		ActorType:      req.GetActorType(),
		OrganizationID: req.GetOrganizationId(),
		IPAddress:      req.GetIpAddress(),
		UserAgent:      req.GetUserAgent(),
		Status:         req.GetStatus(),
		FailureReason:  req.GetFailureReason(),
	}

	// Parse details JSON
	if req.GetDetails() != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(req.GetDetails()), &details); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid details JSON")
		}
		// Marshal back to json.RawMessage for storage
		detailsJSON, _ := json.Marshal(details)
		log.Details = detailsJSON
	}

	if err := h.auditService.Log(ctx, log); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LogResponse{
		Id:      log.ID,
		Success: true,
	}, nil
}

func (h *GrpcHandler) GetByActorID(ctx context.Context, req *pb.GetByActorIDRequest) (*pb.GetAuditLogsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.GetOffset())

	logs, total, err := h.auditService.GetByActorID(ctx, req.GetActorId(), limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAuditLogsResponse{
		Logs:  toProtoLogs(logs),
		Total: total,
	}, nil
}

func (h *GrpcHandler) GetByResourceID(ctx context.Context, req *pb.GetByResourceIDRequest) (*pb.GetAuditLogsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.GetOffset())

	logs, total, err := h.auditService.GetByResourceID(ctx, req.GetResourceId(), limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAuditLogsResponse{
		Logs:  toProtoLogs(logs),
		Total: total,
	}, nil
}

func (h *GrpcHandler) GetByTimeRange(ctx context.Context, req *pb.GetByTimeRangeRequest) (*pb.GetAuditLogsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.GetOffset())

	from := req.GetFrom().AsTime()
	to := req.GetTo().AsTime()

	logs, total, err := h.auditService.GetByTimeRange(ctx, from, to, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAuditLogsResponse{
		Logs:  toProtoLogs(logs),
		Total: total,
	}, nil
}

func toProtoLogs(logs []*domain.AuditLog) []*pb.AuditLog {
	result := make([]*pb.AuditLog, len(logs))
	for i, log := range logs {
		detailsStr := "{}"
		if len(log.Details) > 0 {
			detailsStr = string(log.Details)
		}
		
		result[i] = &pb.AuditLog{
			Id:            log.ID,
			AggregateId:   log.AggregateID,
			EventType:     log.EventType,
			ResourceType:  log.ResourceType,
			Action:        log.Action,
			ActorId:       log.ActorID,
			ActorType:     log.ActorType,
			OrganizationId: log.OrganizationID,
			IpAddress:     log.IPAddress,
			UserAgent:     log.UserAgent,
			Details:       detailsStr,
			Status:        log.Status,
			FailureReason: log.FailureReason,
			CreatedAt:     timestamppb.New(log.CreatedAt),
			Archived:      log.Archived,
		}
	}
	return result
}

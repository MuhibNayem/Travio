package handler

import (
	"context"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fraud/v1"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/service"
)

// GrpcHandler implements the FraudService gRPC server.
type GrpcHandler struct {
	pb.UnimplementedFraudServiceServer
	svc *service.FraudService
}

// NewGrpcHandler creates a new gRPC handler.
func NewGrpcHandler(svc *service.FraudService) *GrpcHandler {
	return &GrpcHandler{svc: svc}
}

// Health returns the health status of the service.
func (h *GrpcHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

// AnalyzeBooking analyzes a booking for fraud indicators.
func (h *GrpcHandler) AnalyzeBooking(ctx context.Context, req *pb.AnalyzeBookingRequest) (*pb.AnalyzeBookingResponse, error) {
	domainReq := &domain.BookingAnalysisRequest{
		OrderID:             req.OrderId,
		OrganizationID:      req.OrganizationId,
		UserID:              req.UserId,
		TripID:              req.TripId,
		PassengerNIDs:       req.PassengerNids,
		PassengerNames:      req.PassengerNames,
		PassengerCount:      int(req.PassengerCount),
		BookingTimestamp:    time.Unix(req.BookingTimestamp, 0),
		IPAddress:           req.IpAddress,
		UserAgent:           req.UserAgent,
		PaymentMethod:       req.PaymentMethod,
		TotalAmountPaisa:    req.TotalAmountPaisa,
		BookingsLast24Hours: int(req.BookingsLast_24Hours),
		BookingsLastWeek:    int(req.BookingsLastWeek),
		PreviousFraudFlags:  int(req.PreviousFraudFlags),
	}

	result, err := h.svc.AnalyzeBooking(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	riskFactors := make([]*pb.RiskFactor, 0, len(result.RiskFactors))
	for _, rf := range result.RiskFactors {
		riskFactors = append(riskFactors, &pb.RiskFactor{
			Code:        rf.Code,
			Description: rf.Description,
			Severity:    rf.Severity,
			Score:       int32(rf.Score),
		})
	}

	return &pb.AnalyzeBookingResponse{
		RiskScore:   int32(result.RiskScore),
		RiskLevel:   string(result.RiskLevel),
		Confidence:  int32(result.Confidence),
		ShouldBlock: result.ShouldBlock,
		RiskFactors: riskFactors,
		Summary:     result.Summary,
		Model:       result.Model,
	}, nil
}

// VerifyDocument verifies a document image for authenticity.
func (h *GrpcHandler) VerifyDocument(ctx context.Context, req *pb.VerifyDocumentRequest) (*pb.VerifyDocumentResponse, error) {
	domainReq := &domain.DocumentVerificationRequest{
		DocumentType:   req.DocumentType,
		DocumentImage:  req.DocumentImage,
		ImageMimeType:  req.ImageMimeType,
		ExpectedNID:    req.ExpectedNid,
		ExpectedName:   req.ExpectedName,
		OrganizationID: req.OrganizationId,
	}

	result, err := h.svc.VerifyDocument(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	return &pb.VerifyDocumentResponse{
		IsAuthentic:    result.IsAuthentic,
		Confidence:     int32(result.Confidence),
		ExtractedNid:   result.ExtractedNID,
		ExtractedName:  result.ExtractedName,
		TamperingScore: int32(result.TamperingScore),
		Issues:         result.Issues,
		Summary:        result.Summary,
	}, nil
}

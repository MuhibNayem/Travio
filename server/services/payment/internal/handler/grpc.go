package handler

import (
	"context"

	pb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedPaymentServiceServer
	paymentService *service.PaymentService
	registry       *gateway.Registry
}

func NewGrpcHandler(svc *service.PaymentService, reg *gateway.Registry) *GrpcHandler {
	return &GrpcHandler{paymentService: svc, registry: reg}
}

func (h *GrpcHandler) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	result, err := h.paymentService.CreatePayment(ctx, &service.CreatePaymentReq{
		OrderID:       req.OrderId,
		AmountPaisa:   req.AmountPaisa,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Description:   req.Description,
		ReturnURL:     req.ReturnUrl,
		CancelURL:     req.CancelUrl,
		IPNURL:        req.IpnUrl,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreatePaymentResponse{
		PaymentId:   result.PaymentID,
		SessionId:   result.SessionID,
		RedirectUrl: result.RedirectURL,
		Gateway:     result.Gateway,
		Status:      result.Status,
	}, nil
}

func (h *GrpcHandler) VerifyPayment(ctx context.Context, req *pb.VerifyPaymentRequest) (*pb.PaymentStatusResponse, error) {
	result, err := h.paymentService.VerifyPayment(ctx, req.Gateway, req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.PaymentStatusResponse{
		TransactionId: result.TransactionID,
		GatewayRef:    result.GatewayRef,
		Status:        string(result.Status),
		AmountPaisa:   result.AmountPaisa,
		Currency:      result.Currency,
		CardBrand:     result.CardBrand,
		CardLast4:     result.CardLast4,
		FailureReason: result.FailureReason,
	}, nil
}

func (h *GrpcHandler) CapturePayment(ctx context.Context, req *pb.CapturePaymentRequest) (*pb.PaymentStatusResponse, error) {
	result, err := h.paymentService.CapturePayment(ctx, req.Gateway, req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.PaymentStatusResponse{
		TransactionId: result.TransactionID,
		GatewayRef:    result.GatewayRef,
		Status:        string(result.Status),
		AmountPaisa:   result.AmountPaisa,
		Currency:      result.Currency,
	}, nil
}

func (h *GrpcHandler) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundResponse, error) {
	result, err := h.paymentService.RefundPayment(ctx, req.Gateway, req.TransactionId, req.AmountPaisa, req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RefundResponse{
		RefundId:      result.RefundID,
		TransactionId: result.TransactionID,
		AmountPaisa:   result.AmountPaisa,
		Status:        result.Status,
	}, nil
}

func (h *GrpcHandler) HandleIPN(ctx context.Context, req *pb.IPNRequest) (*pb.IPNResponse, error) {
	gw, err := h.registry.Get(req.Gateway)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unknown gateway")
	}

	ipnData, err := gw.ValidateIPN(ctx, req.Payload)
	if err != nil {
		return &pb.IPNResponse{Valid: false}, nil
	}

	return &pb.IPNResponse{
		Valid:         ipnData.IsValid,
		TransactionId: ipnData.TransactionID,
		OrderId:       ipnData.OrderID,
		Status:        string(ipnData.Status),
	}, nil
}

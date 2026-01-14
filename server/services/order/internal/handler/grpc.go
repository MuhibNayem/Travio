package handler

import (
	"context"

	pb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/order/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedOrderServiceServer
	orderService *service.OrderService
}

func NewGrpcHandler(orderService *service.OrderService) *GrpcHandler {
	return &GrpcHandler{orderService: orderService}
}

func (h *GrpcHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	var passengers []service.PassengerRequest
	for _, p := range req.Passengers {
		passengers = append(passengers, service.PassengerRequest{
			NID:         p.Nid,
			Name:        p.Name,
			SeatID:      p.SeatId,
			DateOfBirth: p.DateOfBirth,
			Gender:      p.Gender,
			Age:         int(p.Age),
		})
	}

	order, err := h.orderService.CreateOrder(ctx, &service.CreateOrderRequest{
		OrgID:          req.OrganizationId,
		UserID:         req.UserId,
		TripID:         req.TripId,
		FromStation:    req.FromStationId,
		ToStation:      req.ToStationId,
		HoldID:         req.HoldId,
		Passengers:     passengers,
		PaymentToken:   req.PaymentMethod.Token,
		PaymentMethod:  req.PaymentMethod.Type,
		Email:          req.ContactEmail,
		Phone:          req.ContactPhone,
		CouponCode:     req.CouponCode,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateOrderResponse{
		Order: orderToProto(order),
	}, nil
}

func (h *GrpcHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	order, err := h.orderService.GetOrder(ctx, req.OrderId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return orderToProto(order), nil
}

func (h *GrpcHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, total, nextToken, err := h.orderService.ListOrders(ctx, req.UserId, req.Status.String(), int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbOrders []*pb.Order
	for _, o := range orders {
		pbOrders = append(pbOrders, orderToProto(o))
	}

	return &pb.ListOrdersResponse{
		Orders:        pbOrders,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
}

func (h *GrpcHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	order, refund, err := h.orderService.CancelOrder(ctx, req.OrderId, req.UserId, req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.CancelOrderResponse{
		Success: true,
		Order:   orderToProto(order),
	}

	if refund != nil {
		resp.Refund = &pb.RefundInfo{
			RefundId:    refund.RefundID,
			AmountPaisa: refund.AmountPaisa,
			Status:      refund.Status,
		}
	}

	return resp, nil
}

func (h *GrpcHandler) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.OrderStatusResponse, error) {
	order, sagaInst, err := h.orderService.GetOrderStatus(ctx, req.OrderId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	resp := &pb.OrderStatusResponse{
		Status: mapOrderStatus(order.Status),
	}

	if sagaInst != nil {
		resp.Saga = sagaToProto(sagaInst)
	}

	return resp, nil
}

func (h *GrpcHandler) RetryOrder(ctx context.Context, req *pb.RetryOrderRequest) (*pb.RetryOrderResponse, error) {
	// Not implemented - would require saga retry integration
	return &pb.RetryOrderResponse{Success: false}, nil
}

// Converters
func orderToProto(o *domain.Order) *pb.Order {
	if o == nil {
		return nil
	}

	var passengers []*pb.Passenger
	for _, p := range o.Passengers {
		passengers = append(passengers, &pb.Passenger{
			Nid:         p.NID,
			Name:        p.Name,
			SeatId:      p.SeatID,
			SeatNumber:  p.SeatNumber,
			SeatClass:   p.SeatClass,
			Gender:      p.Gender,
			Age:         int32(p.Age),
			NidVerified: p.NIDVerified,
		})
	}

	var seats []*pb.BookedSeat
	for _, s := range o.Seats {
		seats = append(seats, &pb.BookedSeat{
			SeatId:     s.SeatID,
			SeatNumber: s.SeatNumber,
			SeatClass:  s.SeatClass,
			TicketId:   s.TicketID,
			PricePaisa: s.PricePaisa,
		})
	}

	return &pb.Order{
		Id:              o.ID,
		OrganizationId:  o.OrganizationID,
		UserId:          o.UserID,
		TripId:          o.TripID,
		FromStationId:   o.FromStationID,
		ToStationId:     o.ToStationID,
		Passengers:      passengers,
		SubtotalPaisa:   o.SubtotalPaisa,
		TaxPaisa:        o.TaxPaisa,
		BookingFeePaisa: o.BookingFeePaisa,
		DiscountPaisa:   o.DiscountPaisa,
		TotalPaisa:      o.TotalPaisa,
		Currency:        o.Currency,
		PaymentId:       o.PaymentID,
		PaymentStatus:   mapPaymentStatus(o.PaymentStatus),
		BookingId:       o.BookingID,
		Seats:           seats,
		Status:          mapOrderStatus(o.Status),
		ContactEmail:    o.ContactEmail,
		ContactPhone:    o.ContactPhone,
		CreatedAt:       o.CreatedAt.Unix(),
		UpdatedAt:       o.UpdatedAt.Unix(),
		ExpiresAt:       o.ExpiresAt.Unix(),
	}
}

func sagaToProto(s interface{}) *pb.SagaState {
	// Simplified - would need proper saga state conversion
	return &pb.SagaState{
		Status: pb.SagaStatus_SAGA_STATUS_RUNNING,
	}
}

func mapOrderStatus(s domain.OrderStatus) pb.OrderStatus {
	switch s {
	case domain.OrderStatusPending:
		return pb.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusConfirmed:
		return pb.OrderStatus_ORDER_STATUS_CONFIRMED
	case domain.OrderStatusFailed:
		return pb.OrderStatus_ORDER_STATUS_FAILED
	case domain.OrderStatusCancelled:
		return pb.OrderStatus_ORDER_STATUS_CANCELLED
	case domain.OrderStatusRefunded:
		return pb.OrderStatus_ORDER_STATUS_REFUNDED
	default:
		return pb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapPaymentStatus(s string) pb.PaymentStatus {
	switch s {
	case domain.PaymentStatusPending:
		return pb.PaymentStatus_PAYMENT_STATUS_PENDING
	case domain.PaymentStatusAuthorized:
		return pb.PaymentStatus_PAYMENT_STATUS_AUTHORIZED
	case domain.PaymentStatusCaptured:
		return pb.PaymentStatus_PAYMENT_STATUS_CAPTURED
	case domain.PaymentStatusFailed:
		return pb.PaymentStatus_PAYMENT_STATUS_FAILED
	case domain.PaymentStatusRefunded:
		return pb.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

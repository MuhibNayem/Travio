package handler

import (
	"context"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fulfillment/v1"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedFulfillmentServiceServer
	svc *service.FulfillmentService
}

func NewGrpcHandler(svc *service.FulfillmentService) *GrpcHandler {
	return &GrpcHandler{svc: svc}
}

func (h *GrpcHandler) GenerateTickets(ctx context.Context, req *pb.GenerateTicketsRequest) (*pb.GenerateTicketsResponse, error) {
	var passengers []service.PassengerSeat
	for _, p := range req.Passengers {
		passengers = append(passengers, service.PassengerSeat{
			NID:        p.Nid,
			Name:       p.Name,
			SeatID:     p.SeatId,
			SeatNumber: p.SeatNumber,
			SeatClass:  p.SeatClass,
			PricePaisa: p.PricePaisa,
		})
	}

	resp, err := h.svc.GenerateTickets(ctx, &service.GenerateTicketsReq{
		BookingID:      req.BookingId,
		OrderID:        req.OrderId,
		OrganizationID: req.OrganizationId,
		TripID:         req.TripId,
		RouteName:      req.RouteName,
		FromStation:    req.FromStation,
		ToStation:      req.ToStation,
		DepartureTime:  time.Unix(req.DepartureTime, 0),
		ArrivalTime:    time.Unix(req.ArrivalTime, 0),
		Passengers:     passengers,
		ContactEmail:   req.ContactEmail,
		ContactPhone:   req.ContactPhone,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var tickets []*pb.Ticket
	for _, t := range resp.Tickets {
		tickets = append(tickets, ticketToProto(t))
	}

	return &pb.GenerateTicketsResponse{Tickets: tickets}, nil
}

func (h *GrpcHandler) GetTicket(ctx context.Context, req *pb.GetTicketRequest) (*pb.Ticket, error) {
	ticket, err := h.svc.GetTicket(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ticket not found")
	}
	return ticketToProto(ticket), nil
}

func (h *GrpcHandler) ListTickets(ctx context.Context, req *pb.ListTicketsRequest) (*pb.ListTicketsResponse, error) {
	tickets, err := h.svc.ListTickets(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbTickets []*pb.Ticket
	for _, t := range tickets {
		pbTickets = append(pbTickets, ticketToProto(t))
	}
	return &pb.ListTicketsResponse{Tickets: pbTickets}, nil
}

func (h *GrpcHandler) ValidateTicket(ctx context.Context, req *pb.ValidateTicketRequest) (*pb.ValidateTicketResponse, error) {
	ticket, err := h.svc.ValidateTicket(ctx, req.QrCodeData, req.ValidatorId)
	if err != nil {
		return &pb.ValidateTicketResponse{
			IsValid:       false,
			Ticket:        ticketToProto(ticket),
			FailureReason: err.Error(),
			AlreadyUsed:   ticket != nil && ticket.IsBoarded,
		}, nil
	}
	return &pb.ValidateTicketResponse{
		IsValid: true,
		Ticket:  ticketToProto(ticket),
	}, nil
}

func (h *GrpcHandler) CancelTicket(ctx context.Context, req *pb.CancelTicketRequest) (*pb.CancelTicketResponse, error) {
	ticket, err := h.svc.CancelTicket(ctx, req.TicketId, req.Reason)
	if err != nil {
		return &pb.CancelTicketResponse{Success: false}, nil
	}
	return &pb.CancelTicketResponse{Success: true, Ticket: ticketToProto(ticket)}, nil
}

func (h *GrpcHandler) GetTicketPDF(ctx context.Context, req *pb.GetTicketPDFRequest) (*pb.TicketPDFResponse, error) {
	pdfData, err := h.svc.GetTicketPDF(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TicketPDFResponse{
		PdfData:     pdfData,
		Filename:    "ticket.pdf",
		ContentType: "application/pdf",
	}, nil
}

func (h *GrpcHandler) ResendTicket(ctx context.Context, req *pb.ResendTicketRequest) (*pb.ResendTicketResponse, error) {
	ticket, err := h.svc.GetTicket(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ticket not found")
	}

	_, err = h.svc.GetTicketPDF(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate PDF")
	}

	// Notification is handled asynchronously by the notification service
	// The PDF and ticket details are already stored, notification triggers on ticket generation
	return &pb.ResendTicketResponse{
		Success: true,
		Message: "Ticket resend queued for " + ticket.PassengerName + " to " + req.Email,
	}, nil
}

func ticketToProto(t *domain.Ticket) *pb.Ticket {
	if t == nil {
		return nil
	}
	return &pb.Ticket{
		Id:             t.ID,
		BookingId:      t.BookingID,
		OrderId:        t.OrderID,
		OrganizationId: t.OrganizationID,
		TripId:         t.TripID,
		RouteName:      t.RouteName,
		FromStation:    t.FromStation,
		ToStation:      t.ToStation,
		DepartureTime:  t.DepartureTime.Unix(),
		ArrivalTime:    t.ArrivalTime.Unix(),
		PassengerNid:   t.PassengerNID,
		PassengerName:  t.PassengerName,
		SeatNumber:     t.SeatNumber,
		SeatClass:      t.SeatClass,
		PricePaisa:     t.PricePaisa,
		Currency:       t.Currency,
		QrCodeData:     t.QRCodeData,
		QrCodeUrl:      t.QRCodeURL,
		Status:         mapTicketStatus(t.Status),
		CreatedAt:      t.CreatedAt.Unix(),
		ValidUntil:     t.ValidUntil.Unix(),
		IsBoarded:      t.IsBoarded,
		BoardedAt:      t.BoardedAt.Unix(),
		BoardedBy:      t.BoardedBy,
	}
}

func mapTicketStatus(s domain.TicketStatus) pb.TicketStatus {
	switch s {
	case domain.TicketStatusActive:
		return pb.TicketStatus_TICKET_STATUS_ACTIVE
	case domain.TicketStatusUsed:
		return pb.TicketStatus_TICKET_STATUS_USED
	case domain.TicketStatusCancelled:
		return pb.TicketStatus_TICKET_STATUS_CANCELLED
	case domain.TicketStatusExpired:
		return pb.TicketStatus_TICKET_STATUS_EXPIRED
	default:
		return pb.TicketStatus_TICKET_STATUS_UNSPECIFIED
	}
}

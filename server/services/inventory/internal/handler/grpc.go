package handler

import (
	"context"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedInventoryServiceServer
	inventoryService *service.InventoryService
}

func NewGrpcHandler(inventoryService *service.InventoryService) *GrpcHandler {
	return &GrpcHandler{inventoryService: inventoryService}
}

func (h *GrpcHandler) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	result, err := h.inventoryService.CheckAvailability(ctx, req.OrganizationId, req.TripId, req.FromStationId, req.ToStationId, int(req.Passengers), req.SeatClass)
	if err != nil {
		if err == domain.ErrInvalidStationRange {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "availability check failed")
	}

	var seats []*pb.SeatAvailability
	for _, s := range result.Seats {
		seats = append(seats, &pb.SeatAvailability{
			SeatId:     s.SeatID,
			SeatNumber: s.SeatNumber,
			SeatClass:  s.SeatClass,
			SeatType:   s.SeatType,
			Status:     stringToProtoSeatStatus(s.Status),
		})
	}

	return &pb.CheckAvailabilityResponse{
		IsAvailable:    result.IsAvailable,
		AvailableSeats: int32(result.AvailableCount),
		Seats:          seats,
		PricePaisa:     result.TotalPricePaisa,
		CheckedAt:      result.CheckedAt.Unix(),
	}, nil
}

func (h *GrpcHandler) HoldSeats(ctx context.Context, req *pb.HoldSeatsRequest) (*pb.HoldSeatsResponse, error) {
	holdDuration := time.Duration(req.HoldDurationSeconds) * time.Second
	if holdDuration == 0 {
		holdDuration = 10 * time.Minute
	}

	result, err := h.inventoryService.HoldSeats(ctx, &service.HoldRequest{
		OrganizationID: req.OrganizationId,
		TripID:         req.TripId,
		FromStation:    req.FromStationId,
		ToStation:      req.ToStationId,
		SeatIDs:        req.SeatIds,
		UserID:         req.UserId,
		SessionID:      req.SessionId,
		HoldDuration:   holdDuration,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "hold failed")
	}

	return &pb.HoldSeatsResponse{
		HoldId:        result.HoldID,
		Success:       result.Success,
		HeldSeatIds:   result.HeldSeatIDs,
		FailedSeatIds: result.FailedSeatIDs,
		ExpiresAt:     result.ExpiresAt.Unix(),
		FailureReason: result.FailureReason,
	}, nil
}

func (h *GrpcHandler) ReleaseSeats(ctx context.Context, req *pb.ReleaseSeatsRequest) (*pb.ReleaseSeatsResponse, error) {
	err := h.inventoryService.ReleaseSeats(ctx, req.OrganizationId, req.HoldId, req.UserId)
	if err != nil {
		return &pb.ReleaseSeatsResponse{Success: false}, nil
	}
	return &pb.ReleaseSeatsResponse{Success: true}, nil
}

func (h *GrpcHandler) ConfirmBooking(ctx context.Context, req *pb.ConfirmBookingRequest) (*pb.ConfirmBookingResponse, error) {
	var passengers []service.PassengerInfo
	for _, p := range req.Passengers {
		passengers = append(passengers, service.PassengerInfo{
			NID:  p.PassengerNid,
			Name: p.PassengerName,
		})
	}

	result, err := h.inventoryService.ConfirmBooking(ctx, req.OrganizationId, req.HoldId, req.OrderId, req.UserId, passengers)
	if err != nil {
		if err == domain.ErrHoldExpired || err == domain.ErrHoldNotFound {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, "booking confirmation failed")
	}

	var confirmed []*pb.ConfirmedSeat
	for _, s := range result.ConfirmedSeats {
		confirmed = append(confirmed, &pb.ConfirmedSeat{
			SeatId:     s.SeatID,
			SeatNumber: s.SeatNumber,
			TicketId:   s.TicketID,
		})
	}

	return &pb.ConfirmBookingResponse{
		Success:        result.Success,
		BookingId:      result.BookingID,
		ConfirmedSeats: confirmed,
		FailureReason:  result.FailureReason,
	}, nil
}

func (h *GrpcHandler) GetSeatMap(ctx context.Context, req *pb.GetSeatMapRequest) (*pb.GetSeatMapResponse, error) {
	result, err := h.inventoryService.GetSeatMap(ctx, req.OrganizationId, req.TripId, req.FromStationId, req.ToStationId)
	if err != nil {
		return nil, status.Error(codes.Internal, "seat map retrieval failed")
	}

	var rows []*pb.SeatRow
	for _, r := range result.Rows {
		var seats []*pb.SeatCell
		for _, s := range r.Seats {
			seats = append(seats, &pb.SeatCell{
				SeatId:     s.SeatID,
				SeatNumber: s.SeatNumber,
				Column:     int32(s.Column),
				SeatType:   s.SeatType,
				SeatClass:  s.SeatClass,
				Status:     stringToProtoSeatStatus(s.Status),
				PricePaisa: s.PricePaisa,
			})
		}
		rows = append(rows, &pb.SeatRow{
			RowNumber: int32(r.RowNumber),
			Seats:     seats,
		})
	}

	return &pb.GetSeatMapResponse{
		Rows:   rows,
		Legend: &pb.SeatMapLegend{StatusColors: result.Legend},
	}, nil
}

func (h *GrpcHandler) InitializeTripInventory(ctx context.Context, req *pb.InitializeTripInventoryRequest) (*pb.InitializeTripInventoryResponse, error) {
	// Map Proto to Service Request
	var segments []service.SegmentDef
	for _, s := range req.Segments {
		segments = append(segments, service.SegmentDef{
			SegmentIndex:  int(s.SegmentIndex),
			FromStationID: s.FromStationId,
			ToStationID:   s.ToStationId,
			DepartureTime: s.DepartureTime,
			ArrivalTime:   s.ArrivalTime,
		})
	}

	var seats []service.SeatDef
	if req.SeatConfig != nil {
		for _, s := range req.SeatConfig.Seats {
			seats = append(seats, service.SeatDef{
				SeatID:     s.SeatId,
				SeatNumber: s.SeatNumber,
				Row:        int(s.Row),
				Column:     int(s.Column),
				SeatType:   s.SeatType,
				SeatClass:  s.SeatClass,
				PricePaisa: s.PricePaisa,
			})
		}
	}

	res, err := h.inventoryService.InitializeTripInventory(ctx, &service.InitializeTripRequest{
		TripID:         req.TripId,
		OrganizationID: req.OrganizationId,
		VehicleID:      req.VehicleId,
		Segments:       segments,
		SeatConfig: service.SeatConfig{
			TotalSeats: int(req.SeatConfig.TotalSeats),
			Seats:      seats,
		},
	})

	if err != nil {
		logger.Error("Failed to initialize trip inventory", "error", err, "trip_id", req.TripId)
		return nil, status.Errorf(codes.Internal, "failed to initialize trip inventory: %v", err)
	}

	return &pb.InitializeTripInventoryResponse{
		Success:         res.Success,
		SegmentsCreated: int32(res.SegmentsCreated),
		SeatsCreated:    int32(res.SeatsCreated),
	}, nil
}

func stringToProtoSeatStatus(s string) pb.SeatStatus {
	switch s {
	case domain.SeatStatusAvailable:
		return pb.SeatStatus_SEAT_STATUS_AVAILABLE
	case domain.SeatStatusHeld:
		return pb.SeatStatus_SEAT_STATUS_HELD
	case domain.SeatStatusBooked:
		return pb.SeatStatus_SEAT_STATUS_BOOKED
	case domain.SeatStatusBlocked:
		return pb.SeatStatus_SEAT_STATUS_BLOCKED
	default:
		return pb.SeatStatus_SEAT_STATUS_UNSPECIFIED
	}
}

func (h *GrpcHandler) JoinWaitlist(ctx context.Context, req *pb.JoinWaitlistRequest) (*pb.JoinWaitlistResponse, error) {
	result, err := h.inventoryService.JoinWaitlist(ctx, &service.WaitlistRequest{
		OrganizationID: req.OrganizationId,
		TripID:         req.TripId,
		UserID:         req.UserId,
		SeatClass:      req.SeatClass,
		RequestedSeats: int(req.RequestedSeats),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to join waitlist")
	}

	return &pb.JoinWaitlistResponse{
		Success:  result.Success,
		Message:  result.Message,
		Position: int32(result.Position),
	}, nil
}

func (h *GrpcHandler) GetUserWaitlist(ctx context.Context, req *pb.GetUserWaitlistRequest) (*pb.GetUserWaitlistResponse, error) {
	entries, err := h.inventoryService.GetUserWaitlist(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user waitlist")
	}

	var pbEntries []*pb.WaitlistEntry
	for _, e := range entries {
		pbEntries = append(pbEntries, &pb.WaitlistEntry{
			TripId:         e.TripID,
			OrganizationId: e.OrganizationID,
			SeatClass:      e.SeatClass,
			RequestedSeats: int32(e.RequestedSeats),
			Status:         e.Status,
			CreatedAt:      e.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetUserWaitlistResponse{
		Entries: pbEntries,
	}, nil
}

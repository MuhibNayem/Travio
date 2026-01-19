package handler

import (
	"context"

	pb "github.com/MuhibNayem/Travio/server/api/proto/search/v1"
	"github.com/MuhibNayem/Travio/server/services/search/internal/searcher"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedSearchServiceServer
	searcher *searcher.Searcher
}

func NewGrpcHandler(searcher *searcher.Searcher) *GrpcHandler {
	return &GrpcHandler{searcher: searcher}
}

func (h *GrpcHandler) SearchTrips(ctx context.Context, req *pb.SearchTripsRequest) (*pb.SearchTripsResponse, error) {
	trips, total, err := h.searcher.SearchTrips(
		ctx, req.Query, req.FromStationId, req.ToStationId, req.Date,
		int(req.Limit), int(req.Offset),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var results []*pb.TripResult
	for _, t := range trips {
		results = append(results, &pb.TripResult{
			TripId:          t.TripID,
			VehicleType:     t.VehicleType,
			VehicleClass:    t.VehicleClass,
			DepartureTime:   t.DepartureTime,
			ArrivalTime:     t.ArrivalTime,
			PricePaisa:      t.PricePaisa,
			TotalSeats:      int32(t.TotalSeats),
			AvailableSeats:  int32(t.AvailableSeats),
			FromStationId:   t.FromStationID,
			FromStationName: t.FromStationName,
			FromCity:        t.FromCity,
			ToStationId:     t.ToStationID,
			ToStationName:   t.ToStationName,
			ToCity:          t.ToCity,
			Date:            t.Date,
			Status:          t.Status,
			RouteId:         t.RouteID,
		})
	}

	return &pb.SearchTripsResponse{
		Results: results,
		Total:   int32(total),
	}, nil
}

func (h *GrpcHandler) SearchStations(ctx context.Context, req *pb.SearchStationsRequest) (*pb.SearchStationsResponse, error) {
	stations, err := h.searcher.SearchStations(ctx, req.Query, int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var results []*pb.StationResult
	for _, s := range stations {
		results = append(results, &pb.StationResult{
			StationId: s.StationID,
			Name:      s.Name,
			Location:  s.Location,
			Division:  s.Division,
		})
	}

	return &pb.SearchStationsResponse{
		Results: results,
	}, nil
}

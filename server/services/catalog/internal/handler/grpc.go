package handler

import (
	"context"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedCatalogServiceServer
	catalogService *service.CatalogService
}

func NewGrpcHandler(catalogService *service.CatalogService) *GrpcHandler {
	return &GrpcHandler{catalogService: catalogService}
}

// --- Station Handlers ---

func (h *GrpcHandler) CreateStation(ctx context.Context, req *pb.CreateStationRequest) (*pb.Station, error) {
	station := &domain.Station{
		OrganizationID: req.OrganizationId,
		Code:           req.Code,
		Name:           req.Name,
		City:           req.City,
		State:          req.State,
		Country:        req.Country,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Timezone:       req.Timezone,
		Address:        req.Address,
		Amenities:      req.Amenities,
	}

	created, err := h.catalogService.CreateStation(ctx, station)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create station")
	}

	return stationToProto(created), nil
}

func (h *GrpcHandler) GetStation(ctx context.Context, req *pb.GetStationRequest) (*pb.Station, error) {
	station, err := h.catalogService.GetStation(ctx, req.Id, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "station not found")
	}
	return stationToProto(station), nil
}

func (h *GrpcHandler) ListStations(ctx context.Context, req *pb.ListStationsRequest) (*pb.ListStationsResponse, error) {
	stations, total, nextToken, err := h.catalogService.ListStations(ctx, req.OrganizationId, req.City, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list stations")
	}

	var protoStations []*pb.Station
	for _, s := range stations {
		protoStations = append(protoStations, stationToProto(s))
	}

	return &pb.ListStationsResponse{
		Stations:      protoStations,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
}

func (h *GrpcHandler) UpdateStation(ctx context.Context, req *pb.UpdateStationRequest) (*pb.Station, error) {
	station := &domain.Station{
		ID:             req.Id,
		OrganizationID: req.OrganizationId,
		Name:           req.Name,
		Address:        req.Address,
		Amenities:      req.Amenities,
		Status:         protoStationStatusToString(req.Status),
	}

	updated, err := h.catalogService.UpdateStation(ctx, station)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update station")
	}
	return stationToProto(updated), nil
}

func (h *GrpcHandler) DeleteStation(ctx context.Context, req *pb.DeleteStationRequest) (*pb.DeleteStationResponse, error) {
	err := h.catalogService.DeleteStation(ctx, req.Id, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete station")
	}
	return &pb.DeleteStationResponse{Success: true}, nil
}

// --- Route Handlers ---

func (h *GrpcHandler) CreateRoute(ctx context.Context, req *pb.CreateRouteRequest) (*pb.Route, error) {
	var stops []domain.RouteStop
	for _, s := range req.IntermediateStops {
		stops = append(stops, domain.RouteStop{
			StationID:              s.StationId,
			Sequence:               int(s.Sequence),
			ArrivalOffsetMinutes:   int(s.ArrivalOffsetMinutes),
			DepartureOffsetMinutes: int(s.DepartureOffsetMinutes),
			DistanceFromOriginKm:   int(s.DistanceFromOriginKm),
		})
	}

	route := &domain.Route{
		OrganizationID:       req.OrganizationId,
		Code:                 req.Code,
		Name:                 req.Name,
		OriginStationID:      req.OriginStationId,
		DestinationStationID: req.DestinationStationId,
		IntermediateStops:    stops,
		DistanceKm:           int(req.DistanceKm),
		EstimatedDurationMin: int(req.EstimatedDurationMinutes),
	}

	created, err := h.catalogService.CreateRoute(ctx, route)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create route")
	}
	return routeToProto(created), nil
}

func (h *GrpcHandler) GetRoute(ctx context.Context, req *pb.GetRouteRequest) (*pb.Route, error) {
	route, err := h.catalogService.GetRoute(ctx, req.Id, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "route not found")
	}
	return routeToProto(route), nil
}

func (h *GrpcHandler) ListRoutes(ctx context.Context, req *pb.ListRoutesRequest) (*pb.ListRoutesResponse, error) {
	routes, total, nextToken, err := h.catalogService.ListRoutes(ctx, req.OrganizationId, req.OriginStationId, req.DestinationStationId, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list routes")
	}

	var protoRoutes []*pb.Route
	for _, r := range routes {
		protoRoutes = append(protoRoutes, routeToProto(r))
	}

	return &pb.ListRoutesResponse{
		Routes:        protoRoutes,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
}

// --- Trip Handlers ---

func (h *GrpcHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.Trip, error) {
	trip := &domain.Trip{
		OrganizationID: req.OrganizationId,
		RouteID:        req.RouteId,
		VehicleID:      req.VehicleId,
		VehicleType:    req.VehicleType,
		VehicleClass:   req.VehicleClass,
		DepartureTime:  timestampToTime(req.DepartureTime),
		TotalSeats:     int(req.TotalSeats),
		Pricing:        pricingFromProto(req.Pricing),
	}

	created, err := h.catalogService.CreateTrip(ctx, trip, "plan_starter")
	if err != nil {
		if _, ok := err.(*service.PlanLimitError); ok {
			return nil, status.Error(codes.ResourceExhausted, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create trip")
	}
	return tripToProto(created), nil
}

func (h *GrpcHandler) GetTrip(ctx context.Context, req *pb.GetTripRequest) (*pb.Trip, error) {
	trip, err := h.catalogService.GetTrip(ctx, req.Id, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "trip not found")
	}
	return tripToProto(trip), nil
}

func (h *GrpcHandler) SearchTrips(ctx context.Context, req *pb.SearchTripsRequest) (*pb.SearchTripsResponse, error) {
	results, total, nextToken, err := h.catalogService.SearchTrips(ctx, req.OriginCity, req.DestinationCity, timestampToTime(req.TravelDate), int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "search failed")
	}

	var protoResults []*pb.TripSearchResult
	for _, r := range results {
		protoResults = append(protoResults, &pb.TripSearchResult{
			Trip:               tripToProto(r.Trip),
			Route:              routeToProto(r.Route),
			OriginStation:      stationToProto(r.OriginStation),
			DestinationStation: stationToProto(r.DestinationStation),
			OperatorName:       r.OperatorName,
		})
	}

	return &pb.SearchTripsResponse{
		Results:       protoResults,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
}

func (h *GrpcHandler) CancelTrip(ctx context.Context, req *pb.CancelTripRequest) (*pb.Trip, error) {
	trip, err := h.catalogService.CancelTrip(ctx, req.Id, req.OrganizationId, req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to cancel trip")
	}
	return tripToProto(trip), nil
}

// --- Converters ---

func stationToProto(s *domain.Station) *pb.Station {
	if s == nil {
		return nil
	}
	return &pb.Station{
		Id:             s.ID,
		OrganizationId: s.OrganizationID,
		Code:           s.Code,
		Name:           s.Name,
		City:           s.City,
		State:          s.State,
		Country:        s.Country,
		Latitude:       s.Latitude,
		Longitude:      s.Longitude,
		Timezone:       s.Timezone,
		Address:        s.Address,
		Amenities:      s.Amenities,
		Status:         stringToProtoStationStatus(s.Status),
		CreatedAt:      s.CreatedAt.Unix(),
		UpdatedAt:      s.UpdatedAt.Unix(),
	}
}

func routeToProto(r *domain.Route) *pb.Route {
	if r == nil {
		return nil
	}
	var stops []*pb.RouteStop
	for _, s := range r.IntermediateStops {
		stops = append(stops, &pb.RouteStop{
			StationId:              s.StationID,
			Sequence:               int32(s.Sequence),
			ArrivalOffsetMinutes:   int32(s.ArrivalOffsetMinutes),
			DepartureOffsetMinutes: int32(s.DepartureOffsetMinutes),
			DistanceFromOriginKm:   int32(s.DistanceFromOriginKm),
		})
	}
	return &pb.Route{
		Id:                       r.ID,
		OrganizationId:           r.OrganizationID,
		Code:                     r.Code,
		Name:                     r.Name,
		OriginStationId:          r.OriginStationID,
		DestinationStationId:     r.DestinationStationID,
		IntermediateStops:        stops,
		DistanceKm:               int32(r.DistanceKm),
		EstimatedDurationMinutes: int32(r.EstimatedDurationMin),
		Status:                   stringToProtoRouteStatus(r.Status),
		CreatedAt:                r.CreatedAt.Unix(),
		UpdatedAt:                r.UpdatedAt.Unix(),
	}
}

func tripToProto(t *domain.Trip) *pb.Trip {
	if t == nil {
		return nil
	}
	return &pb.Trip{
		Id:             t.ID,
		OrganizationId: t.OrganizationID,
		RouteId:        t.RouteID,
		VehicleId:      t.VehicleID,
		VehicleType:    t.VehicleType,
		VehicleClass:   t.VehicleClass,
		DepartureTime:  t.DepartureTime.Unix(),
		ArrivalTime:    t.ArrivalTime.Unix(),
		TotalSeats:     int32(t.TotalSeats),
		AvailableSeats: int32(t.AvailableSeats),
		Pricing:        pricingToProto(t.Pricing),
		Status:         stringToProtoTripStatus(t.Status),
		CreatedAt:      t.CreatedAt.Unix(),
		UpdatedAt:      t.UpdatedAt.Unix(),
	}
}

func pricingToProto(p domain.TripPricing) *pb.TripPricing {
	return &pb.TripPricing{
		BasePricePaisa:  p.BasePricePaisa,
		TaxPaisa:        p.TaxPaisa,
		BookingFeePaisa: p.BookingFeePaisa,
		Currency:        p.Currency,
		ClassPrices:     p.ClassPrices,
	}
}

func pricingFromProto(p *pb.TripPricing) domain.TripPricing {
	if p == nil {
		return domain.TripPricing{}
	}
	return domain.TripPricing{
		BasePricePaisa:  p.BasePricePaisa,
		TaxPaisa:        p.TaxPaisa,
		BookingFeePaisa: p.BookingFeePaisa,
		Currency:        p.Currency,
		ClassPrices:     p.ClassPrices,
	}
}

func timestampToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func stringToProtoStationStatus(s string) pb.StationStatus {
	switch s {
	case domain.StationStatusActive:
		return pb.StationStatus_STATION_STATUS_ACTIVE
	case domain.StationStatusInactive:
		return pb.StationStatus_STATION_STATUS_INACTIVE
	default:
		return pb.StationStatus_STATION_STATUS_UNSPECIFIED
	}
}

func protoStationStatusToString(s pb.StationStatus) string {
	switch s {
	case pb.StationStatus_STATION_STATUS_ACTIVE:
		return domain.StationStatusActive
	case pb.StationStatus_STATION_STATUS_INACTIVE:
		return domain.StationStatusInactive
	default:
		return domain.StationStatusActive
	}
}

func stringToProtoRouteStatus(s string) pb.RouteStatus {
	switch s {
	case domain.RouteStatusActive:
		return pb.RouteStatus_ROUTE_STATUS_ACTIVE
	case domain.RouteStatusInactive:
		return pb.RouteStatus_ROUTE_STATUS_INACTIVE
	case domain.RouteStatusSuspended:
		return pb.RouteStatus_ROUTE_STATUS_SUSPENDED
	default:
		return pb.RouteStatus_ROUTE_STATUS_UNSPECIFIED
	}
}

func stringToProtoTripStatus(s string) pb.TripStatus {
	switch s {
	case domain.TripStatusScheduled:
		return pb.TripStatus_TRIP_STATUS_SCHEDULED
	case domain.TripStatusBoarding:
		return pb.TripStatus_TRIP_STATUS_BOARDING
	case domain.TripStatusDeparted:
		return pb.TripStatus_TRIP_STATUS_DEPARTED
	case domain.TripStatusCancelled:
		return pb.TripStatus_TRIP_STATUS_CANCELLED
	default:
		return pb.TripStatus_TRIP_STATUS_UNSPECIFIED
	}
}

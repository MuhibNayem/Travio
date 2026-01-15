package handler

import (
	"context"
	"fmt"
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
	stations, total, nextToken, err := h.catalogService.ListStations(ctx, req.OrganizationId, req.City, req.SearchQuery, int(req.PageSize), req.PageToken)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("ListStations error: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to list stations: %v", err)
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
		fmt.Printf("CreateRoute error: %v\n", err)
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

func (h *GrpcHandler) ListTrips(ctx context.Context, req *pb.ListTripsRequest) (*pb.ListTripsResponse, error) {
	trips, total, nextToken, err := h.catalogService.ListTrips(ctx, req.OrganizationId, req.RouteId, req.ScheduleId, req.ServiceDate, int(req.PageSize), req.PageToken)
	if err != nil {
		fmt.Printf("ListTrips error: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to list trips: %v", err)
	}

	var protoTrips []*pb.Trip
	for _, t := range trips {
		protoTrips = append(protoTrips, tripToProto(t))
	}

	return &pb.ListTripsResponse{
		Trips:         protoTrips,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
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

// --- Schedule Handlers ---

func (h *GrpcHandler) CreateSchedule(ctx context.Context, req *pb.CreateScheduleRequest) (*pb.Schedule, error) {
	fmt.Printf("CreateSchedule gRPC: orgID=%s, routeID=%s, vehicleID=%s\n", req.OrganizationId, req.RouteId, req.VehicleId)

	schedule := &domain.ScheduleTemplate{
		OrganizationID:       req.OrganizationId,
		RouteID:              req.RouteId,
		VehicleID:            req.VehicleId,
		VehicleType:          req.VehicleType,
		VehicleClass:         req.VehicleClass,
		TotalSeats:           int(req.TotalSeats),
		Pricing:              pricingFromProto(req.Pricing),
		DepartureMinutes:     int(req.DepartureMinutes),
		ArrivalOffsetMinutes: int(req.ArrivalOffsetMinutes),
		Timezone:             req.Timezone,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		DaysOfWeek:           int(req.DaysOfWeek),
	}

	created, err := h.catalogService.CreateSchedule(ctx, schedule)
	if err != nil {
		fmt.Printf("CreateSchedule ERROR: %v\n", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create schedule: %v", err))
	}
	return scheduleToProto(created), nil
}

func (h *GrpcHandler) CreateSchedules(ctx context.Context, req *pb.BulkCreateSchedulesRequest) (*pb.BulkCreateSchedulesResponse, error) {
	var schedules []domain.ScheduleTemplate
	for _, s := range req.Schedules {
		schedules = append(schedules, domain.ScheduleTemplate{
			RouteID:              s.RouteId,
			VehicleID:            s.VehicleId,
			VehicleType:          s.VehicleType,
			VehicleClass:         s.VehicleClass,
			TotalSeats:           int(s.TotalSeats),
			Pricing:              pricingFromProto(s.Pricing),
			DepartureMinutes:     int(s.DepartureMinutes),
			ArrivalOffsetMinutes: int(s.ArrivalOffsetMinutes),
			Timezone:             s.Timezone,
			StartDate:            s.StartDate,
			EndDate:              s.EndDate,
			DaysOfWeek:           int(s.DaysOfWeek),
		})
	}

	created, count, err := h.catalogService.CreateSchedules(ctx, req.OrganizationId, schedules)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create schedules")
	}

	var protoSchedules []*pb.Schedule
	for _, s := range created {
		protoSchedules = append(protoSchedules, scheduleToProto(s))
	}

	return &pb.BulkCreateSchedulesResponse{
		Schedules:    protoSchedules,
		CreatedCount: int32(count),
	}, nil
}

func (h *GrpcHandler) GetSchedule(ctx context.Context, req *pb.GetScheduleRequest) (*pb.Schedule, error) {
	schedule, err := h.catalogService.GetSchedule(ctx, req.Id, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "schedule not found")
	}
	return scheduleToProto(schedule), nil
}

func (h *GrpcHandler) ListSchedules(ctx context.Context, req *pb.ListSchedulesRequest) (*pb.ListSchedulesResponse, error) {
	fmt.Printf("ListSchedules gRPC: orgID=%s, routeID=%s, status=%v\n", req.OrganizationId, req.RouteId, req.Status)

	schedules, total, nextToken, err := h.catalogService.ListSchedules(ctx, req.OrganizationId, req.RouteId, protoScheduleStatusToString(req.Status), int(req.PageSize), req.PageToken)
	if err != nil {
		fmt.Printf("ListSchedules ERROR: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to list schedules: %v", err)
	}

	var protoSchedules []*pb.Schedule
	for _, s := range schedules {
		protoSchedules = append(protoSchedules, scheduleToProto(s))
	}

	return &pb.ListSchedulesResponse{
		Schedules:     protoSchedules,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
}

func (h *GrpcHandler) UpdateSchedule(ctx context.Context, req *pb.UpdateScheduleRequest) (*pb.Schedule, error) {
	schedule := &domain.ScheduleTemplate{
		ID:                   req.Id,
		OrganizationID:       req.OrganizationId,
		TotalSeats:           int(req.TotalSeats),
		Pricing:              pricingFromProto(req.Pricing),
		DepartureMinutes:     int(req.DepartureMinutes),
		ArrivalOffsetMinutes: int(req.ArrivalOffsetMinutes),
		Timezone:             req.Timezone,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		DaysOfWeek:           int(req.DaysOfWeek),
		Status:               protoScheduleStatusToString(req.Status),
	}

	updated, err := h.catalogService.UpdateSchedule(ctx, schedule)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update schedule")
	}
	return scheduleToProto(updated), nil
}

func (h *GrpcHandler) DeleteSchedule(ctx context.Context, req *pb.DeleteScheduleRequest) (*pb.DeleteScheduleResponse, error) {
	if err := h.catalogService.DeleteSchedule(ctx, req.Id, req.OrganizationId); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete schedule")
	}
	return &pb.DeleteScheduleResponse{Success: true}, nil
}

func (h *GrpcHandler) AddScheduleException(ctx context.Context, req *pb.AddScheduleExceptionRequest) (*pb.ScheduleException, error) {
	exception := &domain.ScheduleException{
		ScheduleID:  req.ScheduleId,
		ServiceDate: req.ServiceDate,
		IsAdded:     req.IsAdded,
		Reason:      req.Reason,
	}
	created, err := h.catalogService.AddScheduleException(ctx, exception, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add schedule exception")
	}
	return scheduleExceptionToProto(created), nil
}

func (h *GrpcHandler) ListScheduleExceptions(ctx context.Context, req *pb.ListScheduleExceptionsRequest) (*pb.ListScheduleExceptionsResponse, error) {
	exceptions, err := h.catalogService.ListScheduleExceptions(ctx, req.ScheduleId, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list schedule exceptions")
	}

	var protoExceptions []*pb.ScheduleException
	for _, ex := range exceptions {
		protoExceptions = append(protoExceptions, scheduleExceptionToProto(ex))
	}

	return &pb.ListScheduleExceptionsResponse{Exceptions: protoExceptions}, nil
}

func (h *GrpcHandler) GenerateTripInstances(ctx context.Context, req *pb.GenerateTripInstancesRequest) (*pb.GenerateTripInstancesResponse, error) {
	fmt.Printf("GenerateTripInstances gRPC: scheduleID=%s, orgID=%s, startDate=%s, endDate=%s\n",
		req.ScheduleId, req.OrganizationId, req.StartDate, req.EndDate)

	trips, count, err := h.catalogService.GenerateTripInstances(ctx, req.ScheduleId, req.OrganizationId, req.StartDate, req.EndDate)
	if err != nil {
		fmt.Printf("GenerateTripInstances ERROR: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to generate trips: %v", err)
	}

	var protoTrips []*pb.Trip
	for _, t := range trips {
		protoTrips = append(protoTrips, tripToProto(t))
	}

	return &pb.GenerateTripInstancesResponse{
		Trips:        protoTrips,
		CreatedCount: int32(count),
	}, nil
}

func (h *GrpcHandler) GetScheduleHistory(ctx context.Context, req *pb.GetScheduleHistoryRequest) (*pb.GetScheduleHistoryResponse, error) {
	start := time.Now()
	versions, err := h.catalogService.GetScheduleHistory(ctx, req.ScheduleId, req.OrganizationId)
	if err != nil {
		fmt.Printf("GetScheduleHistory error: %v\n", err)
		return nil, status.Error(codes.Internal, "failed to get schedule history")
	}
	fmt.Printf("GetScheduleHistory took %v\n", time.Since(start))

	var protoVersions []*pb.ScheduleVersion
	for _, v := range versions {
		protoSnapshot := scheduleToProto(&v.Snapshot)
		protoVersions = append(protoVersions, &pb.ScheduleVersion{
			Id:         v.ID,
			ScheduleId: v.ScheduleID,
			Version:    int32(v.Version),
			Snapshot:   protoSnapshot,
			CreatedAt:  v.CreatedAt.Unix(),
		})
	}

	return &pb.GetScheduleHistoryResponse{Versions: protoVersions}, nil
}

func (h *GrpcHandler) ListTripInstances(ctx context.Context, req *pb.ListTripInstancesRequest) (*pb.ListTripInstancesResponse, error) {
	results, total, nextToken, err := h.catalogService.ListTripInstances(
		ctx,
		req.OrganizationId,
		req.ScheduleId,
		req.RouteId,
		req.StartDate,
		req.EndDate,
		protoTripStatusToString(req.Status),
		int(req.PageSize),
		req.PageToken,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list trip instances")
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

	return &pb.ListTripInstancesResponse{
		Results:       protoResults,
		NextPageToken: nextToken,
		TotalCount:    int32(total),
	}, nil
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

func scheduleToProto(s *domain.ScheduleTemplate) *pb.Schedule {
	if s == nil {
		return nil
	}
	return &pb.Schedule{
		Id:                   s.ID,
		OrganizationId:       s.OrganizationID,
		RouteId:              s.RouteID,
		VehicleId:            s.VehicleID,
		VehicleType:          s.VehicleType,
		VehicleClass:         s.VehicleClass,
		TotalSeats:           int32(s.TotalSeats),
		Pricing:              pricingToProto(s.Pricing),
		DepartureMinutes:     int32(s.DepartureMinutes),
		ArrivalOffsetMinutes: int32(s.ArrivalOffsetMinutes),
		Timezone:             s.Timezone,
		StartDate:            s.StartDate,
		EndDate:              s.EndDate,
		DaysOfWeek:           int32(s.DaysOfWeek),
		Status:               stringToProtoScheduleStatus(s.Status),
		CreatedAt:            s.CreatedAt.Unix(),
		UpdatedAt:            s.UpdatedAt.Unix(),
		Version:              int32(s.Version),
	}
}

func scheduleExceptionToProto(e *domain.ScheduleException) *pb.ScheduleException {
	if e == nil {
		return nil
	}
	return &pb.ScheduleException{
		Id:          e.ID,
		ScheduleId:  e.ScheduleID,
		ServiceDate: e.ServiceDate,
		IsAdded:     e.IsAdded,
		Reason:      e.Reason,
		CreatedAt:   e.CreatedAt.Unix(),
	}
}

func tripToProto(t *domain.Trip) *pb.Trip {
	if t == nil {
		return nil
	}
	var segments []*pb.TripSegment
	for _, seg := range t.Segments {
		segments = append(segments, &pb.TripSegment{
			SegmentIndex:   int32(seg.SegmentIndex),
			FromStationId:  seg.FromStationID,
			ToStationId:    seg.ToStationID,
			DepartureTime:  seg.DepartureTime.Unix(),
			ArrivalTime:    seg.ArrivalTime.Unix(),
			AvailableSeats: int32(seg.AvailableSeats),
		})
	}
	return &pb.Trip{
		Id:             t.ID,
		OrganizationId: t.OrganizationID,
		ScheduleId:     t.ScheduleID,
		ServiceDate:    t.ServiceDate,
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
		Segments:       segments,
		CreatedAt:      t.CreatedAt.Unix(),
		UpdatedAt:      t.UpdatedAt.Unix(),
	}
}

func pricingToProto(p domain.TripPricing) *pb.TripPricing {
	var segmentPrices []*pb.SegmentPricing
	for _, seg := range p.SegmentPrices {
		segmentPrices = append(segmentPrices, &pb.SegmentPricing{
			FromStationId:      seg.FromStationID,
			ToStationId:        seg.ToStationID,
			BasePricePaisa:     seg.BasePricePaisa,
			ClassPrices:        seg.ClassPrices,
			SeatCategoryPrices: seg.SeatCategoryPrices,
		})
	}
	return &pb.TripPricing{
		BasePricePaisa:     p.BasePricePaisa,
		TaxPaisa:           p.TaxPaisa,
		BookingFeePaisa:    p.BookingFeePaisa,
		Currency:           p.Currency,
		ClassPrices:        p.ClassPrices,
		SeatCategoryPrices: p.SeatCategoryPrices,
		SegmentPrices:      segmentPrices,
	}
}

func pricingFromProto(p *pb.TripPricing) domain.TripPricing {
	if p == nil {
		return domain.TripPricing{}
	}
	segmentPrices := make([]domain.SegmentPricing, 0, len(p.SegmentPrices))
	for _, seg := range p.SegmentPrices {
		if seg == nil {
			continue
		}
		segmentPrices = append(segmentPrices, domain.SegmentPricing{
			FromStationID:      seg.FromStationId,
			ToStationID:        seg.ToStationId,
			BasePricePaisa:     seg.BasePricePaisa,
			ClassPrices:        seg.ClassPrices,
			SeatCategoryPrices: seg.SeatCategoryPrices,
		})
	}
	return domain.TripPricing{
		BasePricePaisa:     p.BasePricePaisa,
		TaxPaisa:           p.TaxPaisa,
		BookingFeePaisa:    p.BookingFeePaisa,
		Currency:           p.Currency,
		ClassPrices:        p.ClassPrices,
		SeatCategoryPrices: p.SeatCategoryPrices,
		SegmentPrices:      segmentPrices,
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
	case domain.TripStatusInTransit:
		return pb.TripStatus_TRIP_STATUS_IN_TRANSIT
	case domain.TripStatusArrived:
		return pb.TripStatus_TRIP_STATUS_ARRIVED
	case domain.TripStatusDelayed:
		return pb.TripStatus_TRIP_STATUS_DELAYED
	case domain.TripStatusCancelled:
		return pb.TripStatus_TRIP_STATUS_CANCELLED
	default:
		return pb.TripStatus_TRIP_STATUS_UNSPECIFIED
	}
}

func protoTripStatusToString(s pb.TripStatus) string {
	switch s {
	case pb.TripStatus_TRIP_STATUS_SCHEDULED:
		return domain.TripStatusScheduled
	case pb.TripStatus_TRIP_STATUS_BOARDING:
		return domain.TripStatusBoarding
	case pb.TripStatus_TRIP_STATUS_DEPARTED:
		return domain.TripStatusDeparted
	case pb.TripStatus_TRIP_STATUS_CANCELLED:
		return domain.TripStatusCancelled
	case pb.TripStatus_TRIP_STATUS_DELAYED:
		return domain.TripStatusDelayed
	case pb.TripStatus_TRIP_STATUS_ARRIVED:
		return domain.TripStatusArrived
	case pb.TripStatus_TRIP_STATUS_IN_TRANSIT:
		return domain.TripStatusInTransit
	default:
		return ""
	}
}

func stringToProtoScheduleStatus(s string) pb.ScheduleStatus {
	switch s {
	case domain.ScheduleStatusActive:
		return pb.ScheduleStatus_SCHEDULE_STATUS_ACTIVE
	case domain.ScheduleStatusInactive:
		return pb.ScheduleStatus_SCHEDULE_STATUS_INACTIVE
	default:
		return pb.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED
	}
}

func protoScheduleStatusToString(s pb.ScheduleStatus) string {
	switch s {
	case pb.ScheduleStatus_SCHEDULE_STATUS_ACTIVE:
		return domain.ScheduleStatusActive
	case pb.ScheduleStatus_SCHEDULE_STATUS_INACTIVE:
		return domain.ScheduleStatusInactive
	default:
		return ""
	}
}

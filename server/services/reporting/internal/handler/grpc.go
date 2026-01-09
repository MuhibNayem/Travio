package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/reporting/v1"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/query"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GrpcHandler implements the ReportingService gRPC server.
type GrpcHandler struct {
	pb.UnimplementedReportingServiceServer
	engine *query.Engine
}

// NewGrpcHandler creates a new gRPC handler.
func NewGrpcHandler(engine *query.Engine) *GrpcHandler {
	return &GrpcHandler{engine: engine}
}

// Health returns the health status.
func (h *GrpcHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

// GetRevenueReport returns revenue data.
func (h *GrpcHandler) GetRevenueReport(ctx context.Context, req *pb.RevenueReportRequest) (*pb.RevenueReportResponse, error) {
	q := domain.ReportQuery{
		OrganizationID: req.OrganizationId,
		StartDate:      req.StartDate.AsTime(),
		EndDate:        req.EndDate.AsTime(),
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
		SortOrder:      req.SortOrder,
	}

	if q.Limit <= 0 {
		q.Limit = 30
	}

	data, err := h.engine.GetRevenueReport(ctx, q)
	if err != nil {
		return nil, err
	}

	resp := &pb.RevenueReportResponse{
		TotalCount: int64(len(data)),
	}

	for _, d := range data {
		resp.Data = append(resp.Data, &pb.RevenueData{
			OrganizationId:    d.OrganizationID,
			Date:              timestamppb.New(d.Date),
			OrderCount:        d.OrderCount,
			TotalRevenuePaisa: d.TotalRevenuePaisa,
			AvgOrderValue:     d.AvgOrderValue,
			Currency:          d.Currency,
		})
	}

	return resp, nil
}

// GetBookingTrends returns booking trends.
func (h *GrpcHandler) GetBookingTrends(ctx context.Context, req *pb.BookingTrendsRequest) (*pb.BookingTrendsResponse, error) {
	q := domain.ReportQuery{
		OrganizationID: req.OrganizationId,
		StartDate:      req.StartDate.AsTime(),
		EndDate:        req.EndDate.AsTime(),
		Granularity:    req.Granularity,
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
	}

	if q.Limit <= 0 {
		q.Limit = 30
	}
	if q.Granularity == "" {
		q.Granularity = "day"
	}

	data, err := h.engine.GetBookingTrends(ctx, q)
	if err != nil {
		return nil, err
	}

	resp := &pb.BookingTrendsResponse{
		TotalCount: int64(len(data)),
	}

	for _, d := range data {
		resp.Data = append(resp.Data, &pb.BookingTrendData{
			OrganizationId: d.OrganizationID,
			Period:         timestamppb.New(d.Period),
			BookingCount:   d.BookingCount,
			CompletedCount: d.CompletedCount,
			CancelledCount: d.CancelledCount,
			ConversionRate: d.ConversionRate,
		})
	}

	return resp, nil
}

// GetTopRoutes returns top routes.
func (h *GrpcHandler) GetTopRoutes(ctx context.Context, req *pb.TopRoutesRequest) (*pb.TopRoutesResponse, error) {
	q := domain.ReportQuery{
		OrganizationID: req.OrganizationId,
		SortBy:         req.SortBy,
		Limit:          int(req.Limit),
	}

	if q.Limit <= 0 {
		q.Limit = 10
	}

	data, err := h.engine.GetTopRoutes(ctx, q)
	if err != nil {
		return nil, err
	}

	resp := &pb.TopRoutesResponse{}
	for _, d := range data {
		resp.Data = append(resp.Data, &pb.TopRouteData{
			OrganizationId: d.OrganizationID,
			TripId:         d.TripID,
			RouteName:      d.RouteName,
			BookingCount:   d.BookingCount,
			Revenue:        d.Revenue,
			AvgOccupancy:   d.AvgOccupancy,
		})
	}

	return resp, nil
}

// GetOrganizationMetrics returns organization metrics.
func (h *GrpcHandler) GetOrganizationMetrics(ctx context.Context, req *pb.OrganizationMetricsRequest) (*pb.OrganizationMetricsResponse, error) {
	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()

	// Default to last 30 days if not specified
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	m, err := h.engine.GetOrganizationMetrics(ctx, req.OrganizationId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &pb.OrganizationMetricsResponse{
		OrganizationId:     m.OrganizationID,
		TotalOrders:        m.TotalOrders,
		TotalRevenue:       m.TotalRevenue,
		AvgOrderValue:      m.AvgOrderValue,
		TotalCustomers:     m.TotalCustomers,
		RepeatCustomerRate: m.RepeatCustomerRate,
		AvgBookingsPerDay:  m.AvgBookingsPerDay,
		CancellationRate:   m.CancellationRate,
		RefundRate:         m.RefundRate,
	}, nil
}

// ExportReport exports report data.
func (h *GrpcHandler) ExportReport(ctx context.Context, req *pb.ExportReportRequest) (*pb.ExportReportResponse, error) {
	q := domain.ReportQuery{
		OrganizationID: req.OrganizationId,
		StartDate:      req.StartDate.AsTime(),
		EndDate:        req.EndDate.AsTime(),
		Limit:          int(req.MaxRows),
	}

	if q.Limit <= 0 || q.Limit > 100000 {
		q.Limit = 10000
	}

	var data interface{}
	var err error
	var filename string

	switch req.ReportType {
	case "revenue":
		data, err = h.engine.GetRevenueReport(ctx, q)
		filename = "revenue_report"
	case "bookings":
		q.Granularity = "day"
		data, err = h.engine.GetBookingTrends(ctx, q)
		filename = "booking_trends"
	case "routes":
		data, err = h.engine.GetTopRoutes(ctx, q)
		filename = "top_routes"
	default:
		return nil, fmt.Errorf("unknown report type: %s", req.ReportType)
	}

	if err != nil {
		return nil, err
	}

	var exportData []byte
	var contentType string

	switch req.Format {
	case "csv":
		exportData, err = exportToCSV(data)
		contentType = "text/csv"
		filename += ".csv"
	case "json":
		exportData, err = json.Marshal(data)
		contentType = "application/json"
		filename += ".json"
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	if err != nil {
		return nil, err
	}

	return &pb.ExportReportResponse{
		Data:        exportData,
		Filename:    filename,
		ContentType: contentType,
		RowCount:    int64(getRowCount(data)),
	}, nil
}

func exportToCSV(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	switch v := data.(type) {
	case []domain.RevenueReport:
		writer.Write([]string{"organization_id", "date", "order_count", "total_revenue_paisa", "avg_order_value", "currency"})
		for _, r := range v {
			writer.Write([]string{
				r.OrganizationID,
				r.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", r.OrderCount),
				fmt.Sprintf("%d", r.TotalRevenuePaisa),
				fmt.Sprintf("%.2f", r.AvgOrderValue),
				r.Currency,
			})
		}
	case []domain.BookingTrend:
		writer.Write([]string{"organization_id", "period", "booking_count", "completed_count", "cancelled_count", "conversion_rate"})
		for _, r := range v {
			writer.Write([]string{
				r.OrganizationID,
				r.Period.Format(time.RFC3339),
				fmt.Sprintf("%d", r.BookingCount),
				fmt.Sprintf("%d", r.CompletedCount),
				fmt.Sprintf("%d", r.CancelledCount),
				fmt.Sprintf("%.2f", r.ConversionRate),
			})
		}
	case []domain.TopRoute:
		writer.Write([]string{"organization_id", "trip_id", "route_name", "booking_count", "revenue"})
		for _, r := range v {
			writer.Write([]string{
				r.OrganizationID,
				r.TripID,
				r.RouteName,
				fmt.Sprintf("%d", r.BookingCount),
				fmt.Sprintf("%d", r.Revenue),
			})
		}
	default:
		return nil, fmt.Errorf("unsupported data type for CSV export")
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func getRowCount(data interface{}) int {
	switch v := data.(type) {
	case []domain.RevenueReport:
		return len(v)
	case []domain.BookingTrend:
		return len(v)
	case []domain.TopRoute:
		return len(v)
	default:
		return 0
	}
}

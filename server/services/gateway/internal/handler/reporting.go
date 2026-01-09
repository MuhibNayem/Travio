package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	reportingv1 "github.com/MuhibNayem/Travio/server/api/proto/reporting/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ReportingHandler handles reporting-related HTTP requests.
type ReportingHandler struct {
	client *client.ReportingClient
}

// NewReportingHandler creates a new reporting handler.
func NewReportingHandler(c *client.ReportingClient) *ReportingHandler {
	return &ReportingHandler{client: c}
}

// RegisterRoutes registers reporting routes.
func (h *ReportingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/reports", func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/revenue", h.GetRevenueReport)
		r.Get("/bookings", h.GetBookingTrends)
		r.Get("/routes", h.GetTopRoutes)
		r.Get("/metrics", h.GetOrganizationMetrics)
		r.Get("/export", h.ExportReport)
		r.Get("/health", h.Health)
	})
}

// GetRevenueReport returns revenue data.
func (h *ReportingHandler) GetRevenueReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	startDate, endDate := parseDateRange(r)
	limit, offset := parsePagination(r)

	resp, err := h.client.GetRevenueReport(r.Context(), &reportingv1.RevenueReportRequest{
		OrganizationId: orgID,
		StartDate:      timestamppb.New(startDate),
		EndDate:        timestamppb.New(endDate),
		Limit:          int32(limit),
		Offset:         int32(offset),
		SortOrder:      r.URL.Query().Get("sort_order"),
	})
	if err != nil {
		logger.Error("Failed to get revenue report", "error", err)
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBookingTrends returns booking trends.
func (h *ReportingHandler) GetBookingTrends(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	startDate, endDate := parseDateRange(r)
	limit, offset := parsePagination(r)
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "day"
	}

	resp, err := h.client.GetBookingTrends(r.Context(), &reportingv1.BookingTrendsRequest{
		OrganizationId: orgID,
		StartDate:      timestamppb.New(startDate),
		EndDate:        timestamppb.New(endDate),
		Granularity:    granularity,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		logger.Error("Failed to get booking trends", "error", err)
		http.Error(w, "Failed to get trends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetTopRoutes returns top routes.
func (h *ReportingHandler) GetTopRoutes(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	resp, err := h.client.GetTopRoutes(r.Context(), &reportingv1.TopRoutesRequest{
		OrganizationId: orgID,
		SortBy:         r.URL.Query().Get("sort_by"),
		Limit:          int32(limit),
	})
	if err != nil {
		logger.Error("Failed to get top routes", "error", err)
		http.Error(w, "Failed to get routes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetOrganizationMetrics returns organization metrics.
func (h *ReportingHandler) GetOrganizationMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	startDate, endDate := parseDateRange(r)

	resp, err := h.client.GetOrganizationMetrics(r.Context(), &reportingv1.OrganizationMetricsRequest{
		OrganizationId: orgID,
		StartDate:      timestamppb.New(startDate),
		EndDate:        timestamppb.New(endDate),
	})
	if err != nil {
		logger.Error("Failed to get organization metrics", "error", err)
		http.Error(w, "Failed to get metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ExportReport exports report data.
func (h *ReportingHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	startDate, endDate := parseDateRange(r)
	reportType := r.URL.Query().Get("type")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	maxRows, _ := strconv.Atoi(r.URL.Query().Get("max_rows"))
	if maxRows <= 0 {
		maxRows = 10000
	}

	resp, err := h.client.ExportReport(r.Context(), &reportingv1.ExportReportRequest{
		OrganizationId: orgID,
		ReportType:     reportType,
		StartDate:      timestamppb.New(startDate),
		EndDate:        timestamppb.New(endDate),
		Format:         format,
		MaxRows:        int32(maxRows),
	})
	if err != nil {
		logger.Error("Failed to export report", "error", err)
		http.Error(w, "Failed to export", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+resp.Filename)
	w.Write(resp.Data)
}

// Health checks reporting service health.
func (h *ReportingHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.Health(r.Context())
	if err != nil {
		http.Error(w, "Reporting service unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": resp.Status})
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default: last 30 days

	if s := r.URL.Query().Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = t
		}
	}
	if e := r.URL.Query().Get("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endDate = t
		}
	}

	return startDate, endDate
}

func parsePagination(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 30
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return limit, offset
}

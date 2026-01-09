// Package query provides report query engine.
package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/clickhouse"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/domain"
)

// Engine handles report queries.
type Engine struct {
	ch *clickhouse.Client
}

// NewEngine creates a new query engine.
func NewEngine(ch *clickhouse.Client) *Engine {
	return &Engine{ch: ch}
}

// GetRevenueReport retrieves revenue data for an organization.
func (e *Engine) GetRevenueReport(ctx context.Context, q domain.ReportQuery) ([]domain.RevenueReport, error) {
	query := `
		SELECT 
			organization_id,
			date,
			sum(order_count) AS order_count,
			sum(total_revenue_paisa) AS total_revenue_paisa,
			avg(avg_order_value) AS avg_order_value
		FROM daily_revenue_mv
		WHERE organization_id = ?
		  AND date >= ?
		  AND date <= ?
		GROUP BY organization_id, date
		ORDER BY date %s
		LIMIT ? OFFSET ?
	`

	sortOrder := "DESC"
	if strings.ToLower(q.SortOrder) == "asc" {
		sortOrder = "ASC"
	}
	query = fmt.Sprintf(query, sortOrder)

	rows, err := e.ch.Query(ctx, query,
		q.OrganizationID,
		q.StartDate,
		q.EndDate,
		q.Limit,
		q.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []domain.RevenueReport
	for rows.Next() {
		var r domain.RevenueReport
		if err := rows.Scan(
			&r.OrganizationID,
			&r.Date,
			&r.OrderCount,
			&r.TotalRevenuePaisa,
			&r.AvgOrderValue,
		); err != nil {
			logger.Warn("Failed to scan revenue row", "error", err)
			continue
		}
		r.Currency = "BDT"
		results = append(results, r)
	}

	return results, nil
}

// GetBookingTrends retrieves booking trends for an organization.
func (e *Engine) GetBookingTrends(ctx context.Context, q domain.ReportQuery) ([]domain.BookingTrend, error) {
	var groupBy, timeTrunc string

	switch q.Granularity {
	case "hour":
		groupBy = "toStartOfHour(hour)"
		timeTrunc = "toStartOfHour(hour)"
	case "day":
		groupBy = "toDate(hour)"
		timeTrunc = "toDate(hour)"
	case "week":
		groupBy = "toStartOfWeek(hour)"
		timeTrunc = "toStartOfWeek(hour)"
	case "month":
		groupBy = "toStartOfMonth(hour)"
		timeTrunc = "toStartOfMonth(hour)"
	default:
		groupBy = "toDate(hour)"
		timeTrunc = "toDate(hour)"
	}

	query := fmt.Sprintf(`
		SELECT 
			organization_id,
			%s AS period,
			sum(booking_count) AS booking_count,
			sum(completed_count) AS completed_count,
			sum(cancelled_count) AS cancelled_count
		FROM hourly_bookings_mv
		WHERE organization_id = ?
		  AND hour >= ?
		  AND hour <= ?
		GROUP BY organization_id, %s
		ORDER BY period DESC
		LIMIT ? OFFSET ?
	`, timeTrunc, groupBy)

	rows, err := e.ch.Query(ctx, query,
		q.OrganizationID,
		q.StartDate,
		q.EndDate,
		q.Limit,
		q.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []domain.BookingTrend
	for rows.Next() {
		var r domain.BookingTrend
		if err := rows.Scan(
			&r.OrganizationID,
			&r.Period,
			&r.BookingCount,
			&r.CompletedCount,
			&r.CancelledCount,
		); err != nil {
			logger.Warn("Failed to scan booking trend row", "error", err)
			continue
		}
		if r.BookingCount > 0 {
			r.ConversionRate = float64(r.CompletedCount) / float64(r.BookingCount) * 100
		}
		results = append(results, r)
	}

	return results, nil
}

// GetTopRoutes retrieves top routes by bookings or revenue.
func (e *Engine) GetTopRoutes(ctx context.Context, q domain.ReportQuery) ([]domain.TopRoute, error) {
	orderBy := "booking_count DESC"
	if q.SortBy == "revenue" {
		orderBy = "revenue DESC"
	}

	query := fmt.Sprintf(`
		SELECT 
			organization_id,
			trip_id,
			sum(booking_count) AS booking_count,
			sum(revenue) AS revenue
		FROM top_routes_mv
		WHERE organization_id = ?
		GROUP BY organization_id, trip_id
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	rows, err := e.ch.Query(ctx, query, q.OrganizationID, q.Limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []domain.TopRoute
	for rows.Next() {
		var r domain.TopRoute
		if err := rows.Scan(
			&r.OrganizationID,
			&r.TripID,
			&r.BookingCount,
			&r.Revenue,
		); err != nil {
			logger.Warn("Failed to scan top route row", "error", err)
			continue
		}
		results = append(results, r)
	}

	return results, nil
}

// GetOrganizationMetrics retrieves overall metrics for an organization.
func (e *Engine) GetOrganizationMetrics(ctx context.Context, orgID string, startDate, endDate time.Time) (*domain.OrganizationMetrics, error) {
	query := `
		SELECT
			organization_id,
			sum(order_count) AS total_orders,
			sum(total_revenue_paisa) AS total_revenue,
			avg(avg_order_value) AS avg_order_value
		FROM daily_revenue_mv
		WHERE organization_id = ?
		  AND date >= ?
		  AND date <= ?
		GROUP BY organization_id
	`

	rows, err := e.ch.Query(ctx, query, orgID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var m domain.OrganizationMetrics
	if rows.Next() {
		if err := rows.Scan(
			&m.OrganizationID,
			&m.TotalOrders,
			&m.TotalRevenue,
			&m.AvgOrderValue,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
	}

	// Calculate additional metrics
	days := endDate.Sub(startDate).Hours() / 24
	if days > 0 {
		m.AvgBookingsPerDay = float64(m.TotalOrders) / days
	}

	// Get cancellation stats
	cancelQuery := `
		SELECT
			sum(booking_count) AS total,
			sum(cancelled_count) AS cancelled
		FROM hourly_bookings_mv
		WHERE organization_id = ?
		  AND hour >= ?
		  AND hour <= ?
	`

	cancelRows, err := e.ch.Query(ctx, cancelQuery, orgID, startDate, endDate)
	if err == nil {
		defer cancelRows.Close()
		if cancelRows.Next() {
			var total, cancelled int64
			if cancelRows.Scan(&total, &cancelled) == nil && total > 0 {
				m.CancellationRate = float64(cancelled) / float64(total) * 100
			}
		}
	}

	return &m, nil
}

// GetCustomReport executes a custom parameterized query.
func (e *Engine) GetCustomReport(ctx context.Context, queryTemplate string, params map[string]interface{}) ([]map[string]interface{}, error) {
	// Build query with named parameters
	query := queryTemplate
	args := make([]interface{}, 0)

	for key, val := range params {
		placeholder := "{" + key + "}"
		if strings.Contains(query, placeholder) {
			query = strings.Replace(query, placeholder, "?", 1)
			args = append(args, val)
		}
	}

	rows, err := e.ch.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("custom query failed: %w", err)
	}
	defer rows.Close()

	columns := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

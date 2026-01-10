# Reporting Service gRPC Documentation

**Package:** `reporting.v1`  
**Internal DNS:** `reporting:9089`  
**Proto File:** `server/api/proto/reporting/v1/reporting.proto`

## Overview
The Reporting Service provides real-time analytics and business intelligence. It sits on top of a **ClickHouse** OLAP warehouse, leveraging Materialized Views for sub-second query performance over millions of records.

## Key Behaviors

### Aggregation Pipeline
- **Source:** Consumes events (`OrderCreated`, `BookingCancelled`, `PaymentCaptured`) via Kafka.
- **Materialized Views:** Data is pre-aggregated into views like `daily_revenue_mv`, `top_routes_mv`, and `hourly_bookings_mv`.
- **Query Engine:** The service queries these views rather than raw tables, ensuring constant-time performance regardless of historical data volume.

### Time Granularity
- **Dynamic Bucketing:** Supports `hour`, `day`, `week`, and `month` granularities.
- **Timezone:** All aggregations are aligned to **UTC** by default (configurable per tenant).

### Derived Metrics
The service automatically computes sophisticated KPIs on the fly:
- **Conversion Rate:** `(Completed Bookings / Total Initiated) * 100`
- **Cancellation Rate:** `(Cancelled Bookings / Total Bookings) * 100`
- **AOV (Average Order Value):** `Total Revenue / Total Orders`

---

## RPC Methods

### `GetRevenueReport`
Returns financial metrics over a time range.

- **Request:** `GetRevenueReportRequest` (StartDate, EndDate).
- **Response:** `RevenueReport` (Daily breakdown of revenue and order volume).

### `GetBookingTrends`
Returns operational metrics (traffic volume, conversion) bucketed by time.

- **Request:** `GetBookingTrendsRequest`.
- **Granularity:** `HOUR`, `DAY`, `WEEK`, `MONTH`.

### `GetTopRoutes`
Identifies high-performing routes.

- **Request:** `GetTopRoutesRequest`.
- **SortBy:** `REVENUE`, `BOOKINGS`.
- **Usage:** Used for dashboard "Leaderboards".

### `GetCustomReport`
Executes pre-defined SQL templates with safe parameter injection.

- **Use Case:** Specialized reports for Enterprise clients.

---

## Message Definitions

### RevenueReport
| Field | Type | Description |
|-------|------|-------------|
| `date` | `string` | YYYY-MM-DD |
| `total_revenue` | `int64` | In Paisa |
| `order_count` | `int32` | Volume |

### BookingMetrics
| Field | Type | Description |
|-------|------|-------------|
| `period` | `string` | Bucket timestamp |
| `conversion_rate` | `double` | Percentage (0-100) |

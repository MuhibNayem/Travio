# 🚀 Travio Project Completion Plan

> **Goal**: Transform Travio from a 75% complete architecture into a 100% production-ready, FAANG-grade multi-modal travel SaaS platform.

**Current State**: Strong microservices foundation with 19 services, excellent distributed systems patterns (Saga, Outbox, Optimistic Locking), modern Svelte 5 frontend with beautiful UI.

**Gaps to Close**: Incomplete booking flow, missing frontend pages, stub services (Audit, Notification providers), no tests, no CI/CD, no observability, mobile UX incomplete.

---

## 📋 TABLE OF CONTENTS

1. [Phase 1: Critical Backend Gaps](#phase-1-critical-backend-gaps)
2. [Phase 2: Complete Frontend & User Flows](#phase-2-complete-frontend--user-flows)
3. [Phase 3: Testing & Quality Assurance](#phase-3-testing--quality-assurance)
4. [Phase 4: Observability & Monitoring](#phase-4-observability--monitoring)
5. [Phase 5: DevOps & CI/CD](#phase-5-devops--cicd)
6. [Phase 6: Security Hardening](#phase-6-security-hardening)
7. [Phase 7: Performance & Scalability](#phase-7-performance--scalability)
8. [Phase 8: Documentation & Runbooks](#phase-8-documentation--runbooks)
9. [Phase 9: Production Readiness](#phase-9-production-readiness)
10. [Delivery Checklist](#delivery-checklist)

---

## PHASE 1: Critical Backend Gaps

### 1.1 Complete the End-to-End Booking Flow

#### 1.1.1 Order Service: Wire the Saga Orchestrator
**Problem**: `NewOrchestrator(nil, nil)` creates saga without persistence or DLQ.

**Tasks**:
- [ ] Inject actual `*gorm.DB` into `NewOrchestrator` for saga state persistence
- [ ] Inject actual `messaging.DLQProducer` for failed compensation handling
- [ ] Implement `DLQProducer` interface with Kafka topic publishing
- [ ] Add saga state recovery on service restart (load incomplete sagas from DB and resume)
- [ ] Add saga status query endpoint: `GET /v1/sagas/{sagaId}`
- [ ] Add saga retry endpoint: `POST /v1/sagas/{sagaId}/retry`

**Files to modify**:
- `server/services/order/cmd/main.go` — wire dependencies
- `server/services/order/internal/messaging/dlq.go` — implement Kafka DLQ producer
- `server/services/order/internal/saga/orchestrator.go` — add recovery logic
- `server/services/order/internal/handler/grpc.go` — add saga query endpoints

#### 1.1.2 Fulfillment Service: Replace Placeholder Passenger Data
**Problem**: Kafka consumer uses hardcoded passenger data (`"Passenger"`, `"seat-1"`).

**Tasks**:
- [ ] Extend `OrderConfirmedPayload` to include full passenger list
- [ ] Modify Order Service to emit passenger details in Kafka event
- [ ] Update fulfillment consumer to map passenger data to ticket generation
- [ ] Add validation: reject events with missing passenger data (log to DLQ)

**Files to modify**:
- `server/services/fulfillment/internal/consumer/order_events.go` — parse full payload
- `server/services/order/internal/events/publisher.go` — emit complete order data

#### 1.1.3 Implement Checkout Page (Backend Endpoint)
**Problem**: No backend endpoint for checkout session management.

**Tasks**:
- [ ] Create `Checkout` aggregate in Order Service
- [ ] Add `POST /v1/checkout` endpoint (creates checkout session from hold)
- [ ] Add `GET /v1/checkout/{checkoutId}` endpoint (returns session + pricing)
- [ ] Add `POST /v1/checkout/{checkoutId}/confirm` endpoint (triggers booking saga)
- [ ] Integrate Pricing Service for dynamic price calculation at checkout
- [ ] Integrate CRM Service for coupon validation at checkout

**New files**:
- `server/services/order/internal/domain/checkout.go`
- `server/services/order/internal/handler/http_checkout.go`
- `server/services/order/internal/service/checkout.go`

---

### 1.2 Complete Payment Service

#### 1.2.1 Implement Payment Webhook/IPN Handling
**Problem**: Gateway interface defines `ValidateIPN` but no HTTP endpoint processes callbacks.

**Tasks**:
- [ ] Add `POST /v1/payments/ipn` endpoint (handles Instant Payment Notifications)
- [ ] Implement IPN signature validation for each provider (SSLCommerz, bKash, Nagad)
- [ ] Update transaction status based on IPN payload
- [ ] Publish `PaymentCompleted` / `PaymentFailed` events to Kafka
- [ ] Add idempotency to IPN handler (duplicate callbacks from gateways)

**New files**:
- `server/services/payment/internal/handler/http_webhook.go`
- `server/services/payment/internal/service/ipn.go`

#### 1.2.2 Add Refund Status Endpoint
**Problem**: Refunds are processed but no way to query their status.

**Tasks**:
- [ ] Add `GET /v1/payments/{orderId}/refund` endpoint
- [ ] Store refund records in `refunds` table
- [ ] Add refund status tracking (`PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`)

**New files**:
- `server/services/payment/internal/domain/refund.go`
- `server/services/payment/internal/repository/refund.go`

---

### 1.3 Implement Audit Service

**Problem**: Service is literally `http.ListenAndServe(":8092", nil)` — empty.

**Tasks**:
- [ ] Set up Postgres database connection
- [ ] Create `audit_logs` table (write-only, append-only)
- [ ] Implement audit event ingestion endpoint: `POST /v1/audit/log`
- [ ] Add Kafka consumer for audit events from other services
- [ ] Implement compliance query endpoints:
  - `GET /v1/audit/user/{userId}` — all actions by user
  - `GET /v1/audit/resource/{resourceId}` — all actions on resource
  - `GET /v1/audit/timerange?from=&to=` — time-bounded query
- [ ] Implement data retention policy (auto-archive after 90 days)
- [ ] Add gRPC service for other services to emit audit events

**New files**:
- `server/services/audit/cmd/main.go` — rewrite
- `server/services/audit/config/config.go`
- `server/services/audit/internal/domain/audit_log.go`
- `server/services/audit/internal/repository/postgres.go`
- `server/services/audit/internal/handler/grpc.go`
- `server/services/audit/internal/handler/http.go`
- `server/services/audit/internal/service/audit.go`

---

### 1.4 Complete Notification Service

#### 1.4.1 Implement Real Email Provider (SMTP)
**Problem**: Only `ConsoleProvider` exists (logs to stdout).

**Tasks**:
- [ ] Implement `SMTPProvider` using `net/smtp` or `gopkg.in/gomail.v2`
- [ ] Support TLS connections
- [ ] Support HTML email bodies (render from templates)
- [ ] Add attachment support (for PDF tickets)
- [ ] Implement retry logic with exponential backoff
- [ ] Add rate limiting (prevent SMTP throttling)

**New files**:
- `server/services/notification/internal/provider/smtp.go`

#### 1.4.2 Implement Real SMS Provider (Twilio)
**Problem**: Only `ConsoleProvider` exists.

**Tasks**:
- [ ] Implement `TwilioProvider` using Twilio REST API
- [ ] Support message templating
- [ ] Add delivery status webhooks
- [ ] Implement rate limiting per phone number

**New files**:
- `server/services/notification/internal/provider/twilio.go`

#### 1.4.3 Wire Providers into Main
**Tasks**:
- [ ] Update `notification/cmd/main.go` to instantiate SMTP/SMS providers from env vars
- [ ] Add env vars: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`
- [ ] Add env vars: `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM_NUMBER`

---

### 1.5 Complete CRM Service

#### 1.5.1 Integrate Coupons into Checkout Flow
**Problem**: Coupon validation exists but is never called during booking.

**Tasks**:
- [ ] Add `coupon_code` field to `CreateOrderRequest`
- [ ] Order Service calls CRM `ValidateCoupon` before creating order
- [ ] Apply discount to order total
- [ ] Increment coupon `UsageCount` on successful order

**Files to modify**:
- `server/services/order/internal/domain/order.go` — add coupon field
- `server/services/order/internal/service/order.go` — add coupon validation step

#### 1.5.2 Implement Full Support Ticket Lifecycle
**Problem**: Support tickets are basic, no status updates, no assignment.

**Tasks**:
- [ ] Add ticket status transitions: `OPEN → IN_PROGRESS → RESOLVED → CLOSED`
- [ ] Add agent assignment (`assigned_to` field)
- [ ] Add ticket priority escalation
- [ ] Add SLA deadline tracking
- [ ] Add email notification on ticket updates (via Kafka → Notification Service)

**Files to modify**:
- `server/services/crm/internal/domain/models.go` — extend SupportTicket
- `server/services/crm/internal/service/service.go` — add lifecycle methods
- `server/services/crm/internal/repository/postgres.go` — add queries

---

### 1.6 Complete Subscription Service

#### 1.6.1 Implement Billing Automation
**Problem**: No auto-renewal, no payment integration.

**Tasks**:
- [ ] Create billing scheduler (cron job)
- [ ] On billing cycle end: create invoice → trigger payment → update subscription
- [ ] Integrate with Payment Service for automatic subscription charge
- [ ] Handle failed payments (retry logic → grace period → suspend)
- [ ] Implement proration on plan upgrades/downgrades

**New files**:
- `server/services/subscription/internal/worker/billing.go`
- `server/services/subscription/internal/service/billing.go`

---

### 1.7 Complete Reporting Service

#### 1.7.1 Expose Report Endpoints
**Problem**: ClickHouse + Kafka infrastructure exists but no gRPC methods implemented.

**Tasks**:
- [ ] Implement `GetRevenueReport` (daily/monthly revenue breakdown)
- [ ] Implement `GetOccupancyReport` (seat occupancy rates by route/operator)
- [ ] Implement `GetUserActivityReport` (bookings per user, retention)
- [ ] Implement `GetRoutePerformanceReport` (top routes, revenue per km)
- [ ] Add CSV/PDF export for each report type

**Files to modify**:
- `server/services/reporting/internal/handler/grpc.go` — implement all RPC methods
- `server/services/reporting/internal/query/engine.go` — add report queries

---

## PHASE 2: Complete Frontend & User Flows

### 2.1 Critical Missing Pages

#### 2.1.1 Checkout Page
**Priority**: 🔴 CRITICAL — Blocks entire booking flow

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/checkout/[holdId]/+page.svelte`
- [ ] Display held seats with expiry countdown timer
- [ ] Show passenger details form (NID, name, DOB, gender for each seat)
- [ ] Add coupon code input with validation
- [ ] Display price breakdown (base + dynamic pricing + tax - coupon = total)
- [ ] Add payment method selector (SSLCommerz, bKash, Nagad)
- [ ] Add "Proceed to Payment" button → triggers Order Service saga
- [ ] Add terms & conditions checkbox
- [ ] Handle hold expiry (redirect back to search if hold expired)

**New files**:
- `client/src/routes/(app)/checkout/[holdId]/+page.svelte`
- `client/src/lib/api/checkout.ts`

#### 2.1.2 Payment Page
**Priority**: 🔴 CRITICAL

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/payment/[orderId]/+page.svelte`
- [ ] Display order summary (seats, route, passenger names, total)
- [ ] Redirect to payment gateway (SSLCommerz/bKash/Nagad hosted page)
- [ ] Handle payment gateway return (success/cancel URLs)
- [ ] Show payment processing state (spinner while IPN confirms)
- [ ] Show payment success → redirect to confirmation
- [ ] Show payment failure → retry option

**New files**:
- `client/src/routes/(app)/payment/[orderId]/+page.svelte`
- `client/src/lib/api/payment.ts`

#### 2.1.3 Booking Confirmation Page
**Priority**: 🔴 CRITICAL

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/confirmation/[orderId]/+page.svelte`
- [ ] Display booking success message with confetti animation
- [ ] Show ticket summary (QR code, seats, route, passengers)
- [ ] Add "Download PDF" button (triggers Fulfillment Service)
- [ ] Add "Add to Calendar" button
- [ ] Add "Share" button (WhatsApp, Facebook, email)
- [ ] Add "View All Tickets" link

**New files**:
- `client/src/routes/(app)/confirmation/[orderId]/+page.svelte`

#### 2.1.4 Order History Page
**Priority**: 🟡 HIGH

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/orders/+page.svelte`
- [ ] Display paginated list of user's orders
- [ ] Show order status (pending, confirmed, cancelled, completed)
- [ ] Add filters: status, date range, route
- [ ] Add "View Details" → links to confirmation page
- [ ] Add "Cancel Order" button (with cancellation modal)

**New files**:
- `client/src/routes/(app)/orders/+page.svelte`

#### 2.1.5 Ticket View Page
**Priority**: 🟡 HIGH

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/tickets/[ticketId]/+page.svelte`
- [ ] Display full ticket with QR code
- [ ] Show passenger details, seat, route, timing
- [ ] Add "Download PDF" button
- [ ] Add "Add to Wallet" button (Apple Wallet / Google Pay)
- [ ] Show ticket status (active, used, cancelled, expired)

**New files**:
- `client/src/routes/(app)/tickets/[ticketId]/+page.svelte`

---

### 2.2 Complete Existing Placeholder Pages

#### 2.2.1 Dashboard (Organization)
**Problem**: Hardcoded stats, placeholder chart.

**Tasks**:
- [ ] Connect to Reporting Service for real revenue data
- [ ] Connect to Order Service for real booking counts
- [ ] Connect to Catalog Service for active trip counts
- [ ] Implement revenue chart using a charting library (Chart.js or ApexCharts)
- [ ] Replace hardcoded activity feed with real-time WebSocket/Kafka stream
- [ ] Add date range picker (today, 7d, 30d, custom)

**Files to modify**:
- `client/src/routes/(app)/dashboard/+page.svelte`

**New dependencies**:
- `chart.js` or `apexcharts`

#### 2.2.2 Sales Module (Counter Booking)
**Problem**: Empty page.

**Tasks**:
- [ ] Build counter booking interface (for walk-in customers at physical counters)
- [ ] Agent selects route, date, seats (same seat map as customer flow)
- [ ] Agent enters passenger details
- [ ] Cash payment recording (no gateway redirect)
- [ ] Print ticket immediately (using `TicketPrint.svelte` component — already exists)

**Files to modify**:
- `client/src/routes/(app)/organization/sales/+page.svelte`
- `client/src/routes/(app)/organization/sales/counter/+page.svelte`

#### 2.2.3 Finance Page
**Problem**: Not found.

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/organization/finance/+page.svelte`
- [ ] Display revenue breakdown (by route, vehicle, date)
- [ ] Show payment method distribution (SSLCommerz vs bKash vs Nagad vs Cash)
- [ ] Show refund history
- [ ] Display payout schedule (for operator settlements)

**New files**:
- `client/src/routes/(app)/organization/finance/+page.svelte`

#### 2.2.4 Settings Page
**Problem**: Not found.

**Tasks**:
- [ ] Create route: `client/src/routes/(app)/organization/settings/+page.svelte`
- [ ] Organization profile editing (name, address, phone, email, website)
- [ ] Payment configuration (enable/disable gateways, enter credentials)
- [ ] Notification preferences (email/SMS templates customization)
- [ ] Team member management (already exists in `/organization/members`)
- [ ] Danger zone (delete organization)

**New files**:
- `client/src/routes/(app)/organization/settings/+page.svelte`

#### 2.2.5 Operations: Routes & Fleet Pages
**Tasks**:
- [ ] Complete `/organization/operations/routes/` — route creation/editing UI
- [ ] Complete `/organization/operations/fleet/` — fleet asset management UI
- [ ] Wire existing `RouteModal.svelte` and `AssetModal.svelte` to APIs

**Files to modify**:
- `client/src/routes/(app)/organization/operations/routes/+page.svelte`
- `client/src/routes/(app)/organization/operations/fleet/+page.svelte`

---

### 2.3 Mobile Responsiveness

#### 2.3.1 Mobile Menu
**Tasks**:
- [ ] Implement mobile navigation drawer in `Navbar.svelte`
- [ ] Use `Sheet` component from `ui/sheet/` for slide-out menu
- [ ] Include all nav links + auth buttons
- [ ] Add backdrop overlay when menu is open
- [ ] Animate open/close with spring physics

**Files to modify**:
- `client/src/lib/components/layouts/Navbar.svelte`

#### 2.3.2 Seat Map on Mobile
**Tasks**:
- [ ] Make seat buttons responsive (smaller on mobile)
- [ ] Add horizontal scroll container for wide seat maps
- [ ] Pinch-to-zoom support (touch gestures)
- [ ] Show selected seat count in sticky bottom bar

**Files to modify**:
- `client/src/lib/components/blocks/SeatMap.svelte`

#### 2.3.3 Responsive Booking Page
**Tasks**:
- [ ] Stack seat map + summary on mobile (single column)
- [ ] Make summary a bottom sheet instead of sidebar on mobile
- [ ] Reduce padding/margins for mobile

**Files to modify**:
- `client/src/routes/(app)/booking/[tripId]/+page.svelte`

---

### 2.4 Search Improvements

#### 2.4.1 Connect Search to Real API
**Problem**: Search results page exists but uses server-side load function that may return mock data.

**Tasks**:
- [ ] Ensure `/search` route calls `searchApi.searchTrips()` with real parameters
- [ ] Add loading skeleton screens (shimmer effect) instead of spinner
- [ ] Add error state with retry button

**Files to modify**:
- `client/src/routes/(main)/search/+page.svelte`
- `client/src/routes/(main)/search/+page.ts` (server load function)

#### 2.4.2 Implement Search Filters
**Tasks**:
- [ ] Price range slider (min-max)
- [ ] Operator filter (checkboxes)
- [ ] Vehicle type filter (AC, Non-AC, Sleeper, etc.)
- [ ] Departure time filter (morning, afternoon, evening, night)
- [ ] Sort options (cheapest, fastest, earliest, latest)
- [ ] Apply filters without page reload (client-side filtering initially)

**New files**:
- `client/src/lib/components/search/SearchFilters.svelte`

---

### 2.5 Event Ticketing UI

#### 2.5.1 Event Seat Picker
**Tasks**:
- [ ] Create venue seat picker using `VenueLayout.svelte` (already exists)
- [ ] Map venue sections to interactive seat map
- [ ] Support general admission (no seat selection)
- [ ] Show ticket type tiers (VIP, Regular, Early Bird)
- [ ] Add cart for multiple ticket types

**Files to modify**:
- `client/src/lib/components/sales/seatmap/VenueLayout.svelte`

#### 2.5.2 Event Detail Page
**Tasks**:
- [ ] Create route: `client/src/routes/(main)/events/[eventId]/+page.svelte`
- [ ] Display event banner, description, venue, date/time
- [ ] Show available ticket types with prices
- [ ] "Buy Tickets" button → checkout flow

**New files**:
- `client/src/routes/(main)/events/[eventId]/+page.svelte`

---

### 2.6 User Profile & Settings

#### 2.6.1 Profile Page
**Tasks**:
- [ ] Create route: `client/src/routes/(app)/profile/+page.svelte`
- [ ] Edit name, email, phone
- [ ] Change password (current + new + confirm)
- [ ] View active sessions with "Revoke" buttons
- [ ] Delete account option

**New files**:
- `client/src/routes/(app)/profile/+page.svelte`
- `client/src/lib/api/profile.ts`

---

## PHASE 3: Testing & Quality Assurance

### 3.1 Unit Tests (Backend)

**Priority**: 🔴 CRITICAL — Currently 0% test coverage

**Tasks**:
- [ ] Identity Service: Test auth flows (register, login, refresh, logout)
- [ ] Inventory Service: Test seat hold/release/confirm logic
- [ ] Order Service: Test saga execution and compensation
- [ ] Payment Service: Test gateway routing and reconciliation
- [ ] Pricing Service: Test rules engine evaluation
- [ ] Queue Service: Test queue join/admit/token validation
- [ ] Coverage target: 80% minimum

**New files** (examples):
- `server/services/identity/internal/service/auth_test.go`
- `server/services/inventory/internal/service/inventory_test.go`
- `server/services/order/internal/saga/orchestrator_test.go`
- `server/services/pricing/internal/engine/engine_test.go`
- `server/services/queue/internal/service/queue_test.go`

### 3.2 Integration Tests (Backend)

**Tasks**:
- [ ] Test service-to-service gRPC calls (use Docker Compose test environment)
- [ ] Test Kafka event publishing/consuming
- [ ] Test Redis operations (locks, queues, caching)
- [ ] Test ScyllaDB queries (seat inventory)
- [ ] Test end-to-end booking flow (register → search → hold → checkout → pay → confirm)

**New files**:
- `server/tests/integration/booking_flow_test.go`
- `server/tests/integration/event_publishing_test.go`
- `server/tests/integration/cache_invalidation_test.go`

### 3.3 Frontend Tests

**Tasks**:
- [ ] Set up Vitest + Svelte Testing Library
- [ ] Test auth flow (login form, registration form, validation)
- [ ] Test seat map component (selection, max limit, disabled states)
- [ ] Test search form (input validation, station selection)
- [ ] Test checkout page (form submission, coupon application)
- [ ] Test navbar (mobile menu, auth state changes)

**New files**:
- `client/src/lib/components/blocks/SeatMap.test.ts`
- `client/src/routes/(auth)/login/+page.test.ts`
- `client/src/lib/components/layouts/Navbar.test.ts`

### 3.4 E2E Tests (Cypress)

**Tasks**:
- [ ] Install Cypress
- [ ] Write E2E test: User registration → login → search → book → checkout → payment mock → confirmation
- [ ] Write E2E test: Operator flow → create org → add vehicle → schedule trip
- [ ] Write E2E test: Booking cancellation → refund → seat release
- [ ] Write E2E test: Queue flow (simulate high load → wait → admission)

**New files**:
- `client/cypress/e2e/booking.cy.ts`
- `client/cypress/e2e/operator_flow.cy.ts`
- `client/cypress/e2e/cancellation.cy.ts`

---

## PHASE 4: Observability & Monitoring

### 4.1 Distributed Tracing

**Tasks**:
- [ ] Add OpenTelemetry SDK to all Go services
- [ ] Add trace propagation headers in gRPC/HTTP calls
- [ ] Deploy Jaeger or Grafana Tempo for trace collection
- [ ] Create trace ID injection into logs (correlate traces + logs)

**New files**:
- `server/pkg/otel/tracer.go`
- `server/pkg/otel/middleware.go`

### 4.2 Metrics & Dashboards

**Tasks**:
- [ ] Add Prometheus metrics to all services:
  - `http_requests_total` (by status code, method, path)
  - `http_request_duration_seconds` (p50, p95, p99)
  - `grpc_requests_total` (by service, method, status)
  - `db_query_duration_seconds`
  - `cache_hit_ratio`
  - `saga_execution_duration_seconds`
  - `queue_size` (current waiting count)
- [ ] Deploy Prometheus + Grafana
- [ ] Create Grafana dashboards:
  - Service Health Overview
  - Booking Flow Performance
  - Database Performance
  - Queue Metrics
  - Error Rate Alerts

**New files**:
- `server/pkg/metrics/metrics.go`
- `infra/grafana/dashboards/booking_flow.json`
- `infra/grafana/dashboards/service_health.json`
- `infra/grafana/dashboards/database.json`

### 4.3 Alerting

**Tasks**:
- [ ] Configure Alertmanager with notification rules:
  - High error rate (>5% of requests)
  - High latency (p95 > 500ms)
  - Service down (health check failure)
  - Queue overflow (>10,000 waiting)
  - Payment failure rate (>10%)
  - Saga failure rate (>2%)
- [ ] Integrate with email/Slack/PagerDuty

**New files**:
- `infra/alertmanager/alertmanager.yml`
- `infra/prometheus/alerts.yml`

### 4.4 Log Aggregation

**Tasks**:
- [ ] Deploy Loki for log collection
- [ ] Configure structured logging (JSON format) across all services
- [ ] Create Grafana Log Explorer views
- [ ] Add log-based alerts (error patterns, panic detection)

---

## PHASE 5: DevOps & CI/CD

### 5.1 CI/CD Pipeline (GitHub Actions)

**Tasks**:
- [ ] Create `.github/workflows/ci.yml`:
  - Trigger: PR, push to main
  - Steps: lint → build → unit test → integration test → security scan
  - Block merge if any step fails
- [ ] Create `.github/workflows/cd.yml`:
  - Trigger: merge to main
  - Steps: build Docker images → push to registry → deploy to staging
  - Manual approval gate for production deployment
- [ ] Create `.github/workflows/pr_preview.yml`:
  - Deploy preview environment for each PR
  - Comment PR with preview URL

**New files**:
- `.github/workflows/ci.yml`
- `.github/workflows/cd.yml`
- `.github/workflows/pr_preview.yml`

### 5.2 Kubernetes Deployment

**Tasks**:
- [ ] Create Helm chart for entire stack:
  - 19 backend services
  - Infrastructure (Postgres, Redis, Kafka, ScyllaDB, OpenSearch, MinIO, ClickHouse)
  - Gateway (Ingress)
  - Frontend (static site serving)
- [ ] Create HPA (Horizontal Pod Autoscaler) configs:
  - Gateway: 3-10 pods
  - Identity: 2-5 pods
  - Inventory: 3-15 pods (peak scaling)
  - Order: 2-8 pods
  - Others: 1-3 pods
- [ ] Create PDB (Pod Disruption Budget) for high availability
- [ ] Create Network Policies (service-to-service access control)

**New files**:
- `infra/helm/travio/Chart.yaml`
- `infra/helm/travio/values.yaml`
- `infra/helm/travio/templates/*` (all service templates)
- `infra/k8s/hpa.yml`
- `infra/k8s/network-policy.yml`
- `infra/k8s/pdb.yml`

### 5.3 Database Migrations

**Tasks**:
- [ ] Set up `golang-migrate` or `goose` for all Postgres databases
- [ ] Create migration files for each service:
  - Identity: users, organizations, refresh_tokens, invites
  - Catalog: stations, routes, trips
  - Inventory: (ScyllaDB migrations already exist)
  - Order: orders, saga_instances
  - Payment: transactions, refunds, configs
  - Fulfillment: tickets
  - Events: venues, events, ticket_types
  - Fleet: assets, asset_locations
  - CRM: coupons, support_tickets, ticket_messages
  - Subscription: plans, subscriptions, usage_events, invoices
  - Fraud: user_profiles, fraud_cases
- [ ] Add migration execution to CI/CD pipeline (run on deploy)

**New files**:
- `server/migrations/identity/001_initial.sql`
- `server/migrations/catalog/001_initial.sql`
- (one per service)

### 5.4 Environment Management

**Tasks**:
- [ ] Create `.env.dev` (development)
- [ ] Create `.env.staging` (staging)
- [ ] Create `.env.prod` (production, gitignored)
- [ ] Set up HashiCorp Vault or AWS Secrets Manager for secret management
- [ ] Remove all hardcoded secrets from code (e.g., Queue Service token secret)

---

## PHASE 6: Security Hardening

### 6.1 API Security

**Tasks**:
- [ ] Add CSRF protection for cookie-based auth
- [ ] Add request body size limits (prevent DoS)
- [ ] Add input sanitization (prevent XSS)
- [ ] Add SQL injection prevention review (all queries use parameterized statements)
- [ ] Add rate limiting per user (not just per IP) for authenticated endpoints
- [ ] Add API versioning strategy (URL-based: `/v1/`, `/v2/`)

### 6.2 Data Security

**Tasks**:
- [ ] Encrypt PII fields at rest (NID, phone numbers, emails)
- [ ] Implement field-level encryption using AES-256
- [ ] Add data masking in logs (never log passwords, tokens, NIDs)
- [ ] Implement GDPR data deletion endpoint
- [ ] Add data retention policies (auto-delete old bookings after X years)

### 6.3 Infrastructure Security

**Tasks**:
- [ ] Enable mTLS for all service-to-service communication (uncomment TLS config)
- [ ] Set up certificate rotation (cert-manager with Let's Encrypt)
- [ ] Implement Pod Security Standards (restricted)
- [ ] Add image signing (cosign)
- [ ] Enable Kubernetes audit logging
- [ ] Set up WAF (Web Application Firewall) for gateway
- [ ] Add DDoS protection (Cloudflare or AWS Shield)

### 6.4 Payment Security

**Tasks**:
- [ ] PCI-DSS compliance review (ensure card data never touches servers)
- [ ] Add payment amount validation (prevent negative/overflow amounts)
- [ ] Implement payment reconciliation dashboard
- [ ] Add fraud rule: flag bookings with mismatched NID + name
- [ ] Add velocity checks: max bookings per user per hour

---

## PHASE 7: Performance & Scalability

### 7.1 Database Optimization

**Tasks**:
- [ ] Add database connection pooling tuning (max connections, idle timeout)
- [ ] Add read replicas for read-heavy services (Catalog, Search)
- [ ] Implement query optimization:
  - Add missing indexes (review slow query logs)
  - Add covering indexes for frequent queries
  - Review N+1 query patterns
- [ ] Implement database partitioning (orders table by date)

### 7.2 Caching Strategy

**Tasks**:
- [ ] Implement Redis Cluster (instead of single-node)
- [ ] Add multi-level caching:
  - L1: In-memory (sync.Map or singleflight) for hot reads
  - L2: Redis for distributed caching
- [ ] Implement cache warming on deployment (pre-populate station, route caches)
- [ ] Add cache invalidation events (on trip update, station update, etc.)

### 7.3 API Performance

**Tasks**:
- [ ] Add response compression (gzip/brotli)
- [ ] Implement request coalescing (singleflight) for repeated identical requests
- [ ] Add pagination to all list endpoints (some already have it, review gaps)
- [ ] Implement GraphQL for complex queries (optional, for dashboard)

### 7.4 Frontend Performance

**Tasks**:
- [ ] Implement code splitting (SvelteKit automatic route-based splitting)
- [ ] Add image optimization (use `@sveltejs/enhanced-img`)
- [ ] Implement service worker for caching static assets
- [ ] Add lazy loading for heavy components (seat maps, charts)
- [ ] Optimize bundle size (review `pnpm build` output, tree-shake unused code)

---

## PHASE 8: Documentation & Runbooks

### 8.1 API Documentation

**Tasks**:
- [ ] Complete OpenAPI/Swagger spec (`docs/api/rest/openapi.yaml`)
- [ ] Add gRPC API documentation (protobuf comments → generated docs)
- [ ] Set up Swagger UI for REST APIs
- [ ] Set up Buf Registry for protobuf documentation

**Files to create**:
- `docs/api/rest/openapi.yaml` (complete spec)
- `docs/api/grpc.md` (gRPC service documentation)

### 8.2 Operational Runbooks

**Tasks**:
- [ ] Create runbook: "Booking Flow Failure Investigation"
- [ ] Create runbook: "Queue Overflow Response"
- [ ] Create runbook: "Payment Gateway Outage"
- [ ] Create runbook: "Database Failover"
- [ ] Create runbook: "Service Deployment Rollback"
- [ ] Create runbook: "Data Breach Response"
- [ ] Create runbook: "Flash Sale Preparation"

**Files to create**:
- `docs/runbooks/booking-failure.md`
- `docs/runbooks/queue-overflow.md`
- `docs/runbooks/payment-outage.md`
- `docs/runbooks/db-failover.md`
- `docs/runbooks/deployment-rollback.md`
- `docs/runbooks/data-breach.md`
- `docs/runbooks/flash-sale.md`

### 8.3 Developer Documentation

**Tasks**:
- [ ] Create `CONTRIBUTING.md` (development setup, coding standards, PR process)
- [ ] Create `ARCHITECTURE.md` (detailed architecture diagrams)
- [ ] Create `LOCAL_DEVELOPMENT.md` (step-by-step local setup guide)
- [ ] Add inline code comments for complex distributed systems logic
- [ ] Create onboarding guide for new team members

**Files to create**:
- `CONTRIBUTING.md`
- `docs/ARCHITECTURE.md`
- `docs/LOCAL_DEVELOPMENT.md`
- `docs/ONBOARDING.md`

---

## PHASE 9: Production Readiness

### 9.1 Load Testing

**Tasks**:
- [ ] Execute load tests for critical services:
  - Identity: 10K login requests/second
  - Inventory: 50K seat availability checks/second
  - Queue: 100K concurrent users joining queue
  - Gateway: 20K requests/second
- [ ] Document baseline metrics (p50, p95, p99 latency, error rate)
- [ ] Identify and fix bottlenecks
- [ ] Set up automated load testing in CI (weekly)

**Files** (already exist, need execution):
- `server/services/inventory/load_test/load.go`
- `server/services/order/load_test/load.go`
- `server/services/gateway/load_test/load.go`
- `server/services/payment/load_test/load.go`

### 9.2 Chaos Engineering

**Tasks**:
- [ ] Test service resilience:
  - Kill random service → verify circuit breaker opens
  - Kill Redis → verify graceful degradation (which services fail, which degrade)
  - Kill Kafka → verify outbox pattern recovery
  - Kill Postgres → verify health checks and auto-reconnect
  - Network partition → verify mTLS and retry logic
- [ ] Document failure mode behavior

### 9.3 Pre-Launch Checklist

**Tasks**:
- [ ] All critical paths have E2E tests passing
- [ ] Load tests meet performance targets (see Phase 7)
- [ ] Security audit completed (no critical/high vulnerabilities)
- [ ] Penetration testing completed
- [ ] Monitoring dashboards operational
- [ ] Alerting configured and tested
- [ ] Runbooks written and reviewed
- [ ] Database backups configured (automated, tested restore)
- [ ] Disaster recovery plan documented
- [ ] SSL certificates provisioned
- [ ] Domain DNS configured
- [ ] CDN configured for static assets
- [ ] Error tracking set up (Sentry)
- [ ] Analytics set up (privacy-compliant, e.g., Plausible)

---

## 📊 DELIVERY CHECKLIST

### Phase Completion Criteria

| Phase | Estimated Effort | Success Criteria |
|-------|-----------------|------------------|
| **Phase 1: Critical Backend Gaps** | 3 weeks | All services functional, booking flow end-to-end |
| **Phase 2: Frontend & User Flows** | 4 weeks | All pages implemented, responsive, accessible |
| **Phase 3: Testing & QA** | 3 weeks | 80%+ backend coverage, E2E tests passing |
| **Phase 4: Observability** | 2 weeks | Traces, metrics, logs, alerts operational |
| **Phase 5: DevOps & CI/CD** | 2 weeks | Automated build/test/deploy pipeline |
| **Phase 6: Security Hardening** | 2 weeks | Zero critical vulnerabilities, mTLS enabled |
| **Phase 7: Performance** | 2 weeks | Load targets met, caching optimized |
| **Phase 8: Documentation** | 1 week | API docs, runbooks, contributor guides |
| **Phase 9: Production Readiness** | 2 weeks | Load tested, chaos tested, launch-ready |

**Total Estimated Effort**: 21 weeks (~5 months) with 2-3 developers working in parallel

---

## 🎯 PRIORITY ORDER (What to Do First)

If resources are limited, tackle in this order:

### Week 1-4: **Booking Flow Completion** (Revenue Blocker)
1. Checkout page (frontend + backend)
2. Payment page + webhook handling
3. Confirmation page
4. Order history page
5. Ticket view page

### Week 5-8: **Testing Foundation** (Quality Blocker)
1. Unit tests for core services (identity, inventory, order, payment)
2. Integration tests for booking flow
3. E2E test for happy path

### Week 9-12: **Frontend Completion** (UX Blocker)
1. Dashboard with real data
2. Mobile menu + responsive fixes
3. Sales module (counter booking)
4. Settings page
5. Search filters + real API connection

### Week 13-16: **Infrastructure** (Scale Blocker)
1. CI/CD pipeline
2. Observability (metrics, traces, logs)
3. Kubernetes deployment manifests
4. Database migrations

### Week 17-21: **Production Hardening** (Reliability Blocker)
1. Security audit + fixes
2. Load testing + optimization
3. Documentation + runbooks
4. Chaos testing
5. Pre-launch checklist

---

## 📝 APPENDIX

### A. Files Currently Missing (Not Found in Repository)

#### Backend
- `server/services/audit/internal/**/*` (entire service empty)
- `server/services/notification/internal/provider/smtp.go`
- `server/services/notification/internal/provider/twilio.go`
- `server/services/subscription/internal/worker/billing.go`
- Any `*_test.go` files (zero test files exist)

#### Frontend
- `client/src/routes/(app)/checkout/[holdId]/+page.svelte`
- `client/src/routes/(app)/payment/[orderId]/+page.svelte`
- `client/src/routes/(app)/confirmation/[orderId]/+page.svelte`
- `client/src/routes/(app)/orders/+page.svelte`
- `client/src/routes/(app)/tickets/[ticketId]/+page.svelte`
- `client/src/routes/(app)/profile/+page.svelte`
- `client/src/routes/(main)/events/[eventId]/+page.svelte`

#### Infrastructure
- `.github/workflows/*` (no CI/CD)
- `infra/helm/*` (no Kubernetes manifests)
- `server/migrations/*` (no migration files)
- `infra/grafana/*` (no dashboards)
- `infra/prometheus/*` (no alerting rules)

### B. Environment Variables to Add

```bash
# Notification Service
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@travio.com
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+1234567890

# Fraud Service
GOOGLE_API_KEY=your_google_api_key  # For embeddings

# Queue Service
QUEUE_TOKEN_SECRET=CHANGE_THIS_TO_A_STRONG_SECRET_IN_PRODUCTION

# Subscription Service
STRIPE_SECRET_KEY=sk_...  # If using Stripe for billing
```

### C. Dependencies to Add

#### Backend
```bash
go get go.opentelemetry.io/otel          # OpenTelemetry
go get go.opentelemetry.io/otel/trace    # Tracing
go get github.com/prometheus/client_golang  # Metrics
go get gopkg.in/gomail.v2                # Email
go get github.com/twilio/twilio-go       # SMS
go get github.com/golang-migrate/migrate/v4  # DB migrations
```

#### Frontend
```bash
pnpm add chart.js                       # Dashboard charts
pnpm add @sveltejs/enhanced-img         # Image optimization
pnpm add -D vitest @testing-library/svelte  # Unit tests
pnpm add -D cypress                     # E2E tests
```

---

**Document Version**: 1.0
**Last Updated**: April 7, 2026
**Author**: Travio Engineering Team
**Status**: Awaiting Review & Approval

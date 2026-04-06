# 🎯 Travio Completion Tasks

> **Instructions**: Each task is independently trackable. Mark `[x]` when complete. Tasks are ordered by priority and grouped by phase for parallel execution.

---

## 🔴 PHASE 1: Critical Booking Flow (Weeks 1-4)

### 1.1 Backend: Checkout Service
- [ ] `TASK-001` Create `Checkout` domain model in Order Service
- [ ] `TASK-002` Implement `POST /v1/checkout` endpoint (create checkout session from hold)
- [ ] `TASK-003` Implement `GET /v1/checkout/{checkoutId}` endpoint (return session + pricing)
- [ ] `TASK-004` Implement `POST /v1/checkout/{checkoutId}/confirm` endpoint (trigger booking saga)
- [ ] `TASK-005` Integrate Pricing Service for dynamic price calculation at checkout
- [ ] `TASK-006` Integrate CRM Service for coupon validation at checkout
- [ ] `TASK-007` Add unit tests for checkout service

**Files to create**:
- `server/services/order/internal/domain/checkout.go`
- `server/services/order/internal/handler/http_checkout.go`
- `server/services/order/internal/service/checkout.go`
- `server/services/order/internal/handler/http_checkout_test.go`

---

### 1.2 Backend: Saga Orchestrator Wiring
- [ ] `TASK-008` Inject `*gorm.DB` into `NewOrchestrator` in Order Service main.go
- [ ] `TASK-009` Implement `DLQProducer` interface with Kafka topic publishing
- [ ] `TASK-010` Add saga state recovery on service restart (load incomplete sagas from DB)
- [ ] `TASK-011` Add `GET /v1/sagas/{sagaId}` endpoint for status query
- [ ] `TASK-012` Add `POST /v1/sagas/{sagaId}/retry` endpoint
- [ ] `TASK-013` Write integration tests for saga execution + compensation

**Files to modify**:
- `server/services/order/cmd/main.go`
- `server/services/order/internal/messaging/dlq.go`
- `server/services/order/internal/saga/orchestrator.go`
- `server/services/order/internal/handler/grpc.go`

---

### 1.3 Backend: Payment Webhooks
- [ ] `TASK-014` Create `POST /v1/payments/ipn` endpoint (handle IPN from gateways)
- [ ] `TASK-015` Implement SSLCommerz IPN signature validation
- [ ] `TASK-016` Implement bKash IPN signature validation
- [ ] `TASK-017` Implement Nagad IPN signature validation
- [ ] `TASK-018` Update transaction status based on IPN payload
- [ ] `TASK-019` Publish `PaymentCompleted` / `PaymentFailed` events to Kafka
- [ ] `TASK-020` Add idempotency to IPN handler (prevent duplicate processing)
- [ ] `TASK-021` Implement `GET /v1/payments/{orderId}/refund` endpoint
- [ ] `TASK-022` Create refund domain model + repository
- [ ] `TASK-023` Write unit tests for IPN handler

**Files to create**:
- `server/services/payment/internal/handler/http_webhook.go`
- `server/services/payment/internal/service/ipn.go`
- `server/services/payment/internal/domain/refund.go`
- `server/services/payment/internal/repository/refund.go`

---

### 1.4 Backend: Fulfillment Passenger Data
- [ ] `TASK-024` Extend `OrderConfirmedPayload` to include full passenger list
- [ ] `TASK-025` Modify Order Service to emit passenger details in Kafka event
- [ ] `TASK-026` Update fulfillment consumer to map passenger data to ticket generation
- [ ] `TASK-027` Add validation: reject events with missing passenger data (log to DLQ)

**Files to modify**:
- `server/services/fulfillment/internal/consumer/order_events.go`
- `server/services/order/internal/events/publisher.go`

---

### 1.5 Backend: Audit Service (From Scratch)
- [ ] `TASK-028` Set up Postgres database connection in Audit Service
- [ ] `TASK-029` Create `audit_logs` table migration (write-only, append-only)
- [ ] `TASK-030` Implement `POST /v1/audit/log` endpoint
- [ ] `TASK-031` Add Kafka consumer for audit events from other services
- [ ] `TASK-032` Implement `GET /v1/audit/user/{userId}` endpoint
- [ ] `TASK-033` Implement `GET /v1/audit/resource/{resourceId}` endpoint
- [ ] `TASK-034` Implement `GET /v1/audit/timerange?from=&to=` endpoint
- [ ] `TASK-035` Implement data retention policy (auto-archive after 90 days)
- [ ] `TASK-036` Add gRPC service for other services to emit audit events

**Files to create**:
- `server/services/audit/cmd/main.go` (rewrite)
- `server/services/audit/config/config.go`
- `server/services/audit/internal/domain/audit_log.go`
- `server/services/audit/internal/repository/postgres.go`
- `server/services/audit/internal/handler/grpc.go`
- `server/services/audit/internal/handler/http.go`
- `server/services/audit/internal/service/audit.go`
- `server/migrations/audit/001_create_audit_logs.sql`

---

### 1.6 Backend: Notification Providers
- [ ] `TASK-037` Implement `SMTPProvider` for email sending (using gomail)
- [ ] `TASK-038` Add TLS connection support in SMTP provider
- [ ] `TASK-039` Add attachment support for PDF tickets in emails
- [ ] `TASK-040` Implement retry logic with exponential backoff for SMTP
- [ ] `TASK-041` Implement `TwilioProvider` for SMS sending
- [ ] `TASK-042` Add rate limiting for SMS per phone number
- [ ] `TASK-043` Wire SMTP + Twilio providers into notification main.go
- [ ] `TASK-044` Add env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, TWILIO_*
- [ ] `TASK-045` Write unit tests for SMTP and Twilio providers

**Files to create**:
- `server/services/notification/internal/provider/smtp.go`
- `server/services/notification/internal/provider/twilio.go`

**Files to modify**:
- `server/services/notification/cmd/main.go`

---

### 1.7 Backend: CRM Coupon Integration
- [ ] `TASK-046` Add `coupon_code` field to `CreateOrderRequest`
- [ ] `TASK-047` Order Service calls CRM `ValidateCoupon` before creating order
- [ ] `TASK-048` Apply discount to order total
- [ ] `TASK-049` Increment coupon `UsageCount` on successful order
- [ ] `TASK-050` Implement full support ticket lifecycle (status transitions, agent assignment)
- [ ] `TASK-051` Add SLA deadline tracking for support tickets
- [ ] `TASK-052` Add email notification on ticket updates via Kafka

**Files to modify**:
- `server/services/order/internal/domain/order.go`
- `server/services/order/internal/service/order.go`
- `server/services/crm/internal/domain/models.go`
- `server/services/crm/internal/service/service.go`

---

### 1.8 Backend: Subscription Billing
- [ ] `TASK-053` Create billing scheduler (cron job) in Subscription Service
- [ ] `TASK-054` Implement invoice generation on billing cycle end
- [ ] `TASK-055` Integrate Payment Service for automatic subscription charge
- [ ] `TASK-056` Handle failed payments (retry → grace period → suspend)
- [ ] `TASK-057` Implement proration on plan upgrades/downgrades
- [ ] `TASK-058` Implement auto-renewal logic

**Files to create**:
- `server/services/subscription/internal/worker/billing.go`
- `server/services/subscription/internal/service/billing.go`

---

### 1.9 Backend: Reporting Endpoints
- [ ] `TASK-059` Implement `GetRevenueReport` gRPC method (daily/monthly breakdown)
- [ ] `TASK-060` Implement `GetOccupancyReport` gRPC method (seat occupancy rates)
- [ ] `TASK-061` Implement `GetUserActivityReport` gRPC method (bookings per user)
- [ ] `TASK-062` Implement `GetRoutePerformanceReport` gRPC method (top routes)
- [ ] `TASK-063` Add CSV export for each report type

**Files to modify**:
- `server/services/reporting/internal/handler/grpc.go`
- `server/services/reporting/internal/query/engine.go`

---

## 🟠 PHASE 2: Frontend Completion (Weeks 5-8)

### 2.1 Critical Missing Pages
- [ ] `TASK-064` Create checkout page route: `/checkout/[holdId]`
- [ ] `TASK-065` Build held seats display with expiry countdown timer
- [ ] `TASK-066` Build passenger details form (NID, name, DOB, gender per seat)
- [ ] `TASK-067` Add coupon code input with live validation
- [ ] `TASK-068` Build price breakdown display (base + dynamic pricing + tax - coupon)
- [ ] `TASK-069` Build payment method selector (SSLCommerz, bKash, Nagad)
- [ ] `TASK-070` Add terms & conditions checkbox
- [ ] `TASK-071` Implement "Proceed to Payment" button (triggers Order Service saga)
- [ ] `TASK-072` Handle hold expiry (redirect to search if expired)

**Files to create**:
- `client/src/routes/(app)/checkout/[holdId]/+page.svelte`
- `client/src/lib/api/checkout.ts`

---

- [ ] `TASK-073` Create payment page route: `/payment/[orderId]`
- [ ] `TASK-074` Build order summary display (seats, route, passengers, total)
- [ ] `TASK-075` Implement payment gateway redirect (SSLCommerz/bKash/Nagad)
- [ ] `TASK-076` Handle payment gateway return (success/cancel URLs)
- [ ] `TASK-077` Show payment processing state (spinner while IPN confirms)
- [ ] `TASK-078` Show payment success → redirect to confirmation
- [ ] `TASK-079` Show payment failure → retry option

**Files to create**:
- `client/src/routes/(app)/payment/[orderId]/+page.svelte`
- `client/src/lib/api/payment.ts`

---

- [ ] `TASK-080` Create confirmation page route: `/confirmation/[orderId]`
- [ ] `TASK-081` Build booking success message with confetti animation
- [ ] `TASK-082` Display ticket summary (QR code, seats, route, passengers)
- [ ] `TASK-083` Add "Download PDF" button
- [ ] `TASK-084` Add "Add to Calendar" button
- [ ] `TASK-085` Add "Share" button (WhatsApp, Facebook, email)

**Files to create**:
- `client/src/routes/(app)/confirmation/[orderId]/+page.svelte`

---

- [ ] `TASK-086` Create order history page route: `/orders`
- [ ] `TASK-087` Build paginated order list with status badges
- [ ] `TASK-088` Add filters (status, date range, route)
- [ ] `TASK-089` Add "View Details" link to confirmation page
- [ ] `TASK-090` Add "Cancel Order" button with confirmation modal

**Files to create**:
- `client/src/routes/(app)/orders/+page.svelte`

---

- [ ] `TASK-091` Create ticket view page route: `/tickets/[ticketId]`
- [ ] `TASK-092` Display full ticket with QR code
- [ ] `TASK-093` Add "Download PDF" button
- [ ] `TASK-094` Show ticket status (active, used, cancelled, expired)

**Files to create**:
- `client/src/routes/(app)/tickets/[ticketId]/+page.svelte`

---

### 2.2 Complete Placeholder Pages
- [ ] `TASK-095` Connect dashboard to Reporting Service for real revenue data
- [ ] `TASK-096` Connect dashboard to Order Service for real booking counts
- [ ] `TASK-097` Connect dashboard to Catalog Service for active trip counts
- [ ] `TASK-098` Implement revenue chart (Chart.js or ApexCharts)
- [ ] `TASK-099` Replace hardcoded activity feed with real-time data
- [ ] `TASK-100` Add date range picker (today, 7d, 30d, custom)

**Files to modify**:
- `client/src/routes/(app)/dashboard/+page.svelte`

**Dependencies to add**:
- `pnpm add chart.js` or `pnpm add apexcharts`

---

- [ ] `TASK-101` Build counter booking interface (walk-in customer flow)
- [ ] `TASK-102` Add agent seat selection (reuse seat map component)
- [ ] `TASK-103` Build passenger details entry form for counter
- [ ] `TASK-104` Implement cash payment recording (no gateway)
- [ ] `TASK-105` Add immediate ticket printing (use TicketPrint.svelte)

**Files to modify**:
- `client/src/routes/(app)/organization/sales/+page.svelte`
- `client/src/routes/(app)/organization/sales/counter/+page.svelte`

---

- [ ] `TASK-106` Create finance page route: `/organization/finance`
- [ ] `TASK-107` Build revenue breakdown by route/vehicle/date
- [ ] `TASK-108` Build payment method distribution chart
- [ ] `TASK-109` Display refund history table
- [ ] `TASK-110` Display payout schedule (operator settlements)

**Files to create**:
- `client/src/routes/(app)/organization/finance/+page.svelte`

---

- [ ] `TASK-111` Create settings page route: `/organization/settings`
- [ ] `TASK-112` Build organization profile editing form
- [ ] `TASK-113` Build payment configuration UI (enable/disable gateways)
- [ ] `TASK-114` Build notification preferences UI

**Files to create**:
- `client/src/routes/(app)/organization/settings/+page.svelte`

---

### 2.3 Mobile Responsiveness
- [ ] `TASK-115` Implement mobile navigation drawer in Navbar (use Sheet component)
- [ ] `TASK-116` Add backdrop overlay when mobile menu is open
- [ ] `TASK-117` Animate menu open/close with spring physics
- [ ] `TASK-118` Make seat map buttons responsive (smaller on mobile)
- [ ] `TASK-119` Add horizontal scroll container for wide seat maps
- [ ] `TASK-120` Stack seat map + summary on mobile (single column)
- [ ] `TASK-121` Make booking summary a bottom sheet on mobile

**Files to modify**:
- `client/src/lib/components/layouts/Navbar.svelte`
- `client/src/lib/components/blocks/SeatMap.svelte`
- `client/src/routes/(app)/booking/[tripId]/+page.svelte`

---

### 2.4 Search Improvements
- [ ] `TASK-122` Connect search results to real API call (remove mock data)
- [ ] `TASK-123` Add loading skeleton screens (shimmer effect)
- [ ] `TASK-124` Add error state with retry button
- [ ] `TASK-125` Build price range slider filter
- [ ] `TASK-126` Build operator filter (checkboxes)
- [ ] `TASK-127` Build vehicle type filter (AC, Non-AC, Sleeper)
- [ ] `TASK-128` Build departure time filter (morning, afternoon, evening, night)
- [ ] `TASK-129` Build sort options (cheapest, fastest, earliest, latest)
- [ ] `TASK-130` Apply filters without page reload

**Files to modify**:
- `client/src/routes/(main)/search/+page.svelte`
- `client/src/routes/(main)/search/+page.ts`

**Files to create**:
- `client/src/lib/components/search/SearchFilters.svelte`

---

### 2.5 Event Ticketing UI
- [ ] `TASK-131` Complete venue seat picker using VenueLayout.svelte
- [ ] `TASK-132` Map venue sections to interactive seat map
- [ ] `TASK-133` Support general admission (no seat selection)
- [ ] `TASK-134` Show ticket type tiers (VIP, Regular, Early Bird)
- [ ] `TASK-135` Create event detail page: `/events/[eventId]`
- [ ] `TASK-136` Display event banner, description, venue, date/time
- [ ] `TASK-137` Show available ticket types with prices
- [ ] `TASK-138` Add "Buy Tickets" button → checkout flow

**Files to modify**:
- `client/src/lib/components/sales/seatmap/VenueLayout.svelte`

**Files to create**:
- `client/src/routes/(main)/events/[eventId]/+page.svelte`

---

### 2.6 User Profile
- [ ] `TASK-139` Create profile page route: `/profile`
- [ ] `TASK-140` Build edit name, email, phone form
- [ ] `TASK-141` Build change password form (current + new + confirm)
- [ ] `TASK-142` Display active sessions with "Revoke" buttons
- [ ] `TASK-143` Add delete account option with confirmation

**Files to create**:
- `client/src/routes/(app)/profile/+page.svelte`
- `client/src/lib/api/profile.ts`

---

## 🟡 PHASE 3: Testing (Weeks 9-11)

### 3.1 Backend Unit Tests
- [ ] `TASK-144` Write unit tests for Identity Service auth flows (register, login, refresh, logout)
- [ ] `TASK-145` Write unit tests for Inventory Service (seat hold/release/confirm)
- [ ] `TASK-146` Write unit tests for Order Service saga execution
- [ ] `TASK-147` Write unit tests for Order Service compensation
- [ ] `TASK-148` Write unit tests for Payment Service gateway routing
- [ ] `TASK-149` Write unit tests for Payment Service reconciliation
- [ ] `TASK-150` Write unit tests for Pricing Service rules engine
- [ ] `TASK-151` Write unit tests for Queue Service (join/admit/token)
- [ ] `TASK-152` Write unit tests for Checkout Service
- [ ] `TASK-153` Write unit tests for Audit Service
- [ ] `TASK-154` Write unit tests for Notification providers (SMTP, Twilio)
- [ ] `TASK-155` Write unit tests for Subscription billing worker
- [ ] `TASK-156` Achieve 80%+ code coverage across all services

**Files to create**:
- `server/services/identity/internal/service/auth_test.go`
- `server/services/inventory/internal/service/inventory_test.go`
- `server/services/order/internal/saga/orchestrator_test.go`
- `server/services/order/internal/service/checkout_test.go`
- `server/services/pricing/internal/engine/engine_test.go`
- `server/services/queue/internal/service/queue_test.go`
- `server/services/audit/internal/service/audit_test.go`
- (one per service)

---

### 3.2 Backend Integration Tests
- [ ] `TASK-157` Test service-to-service gRPC calls (use Docker Compose test env)
- [ ] `TASK-158` Test Kafka event publishing/consuming
- [ ] `TASK-159` Test Redis operations (locks, queues, caching)
- [ ] `TASK-160` Test ScyllaDB queries (seat inventory)
- [ ] `TASK-161` Test end-to-end booking flow (register → search → hold → checkout → pay → confirm)
- [ ] `TASK-162` Test booking cancellation → refund → seat release

**Files to create**:
- `server/tests/integration/booking_flow_test.go`
- `server/tests/integration/event_publishing_test.go`
- `server/tests/integration/cache_invalidation_test.go`

---

### 3.3 Frontend Tests
- [ ] `TASK-163` Set up Vitest + Svelte Testing Library
- [ ] `TASK-164` Test login form (input validation, error display, success redirect)
- [ ] `TASK-165` Test registration form (account type switching, validation, password strength)
- [ ] `TASK-166` Test seat map component (selection, max limit, disabled states)
- [ ] `TASK-167` Test search form (station selection, date picker, navigation)
- [ ] `TASK-168` Test checkout page (form submission, coupon application)
- [ ] `TASK-169` Test navbar (mobile menu toggle, auth state changes)
- [ ] `TASK-170` Test booking page (seat selection, price calculation, hold request)

**Files to create**:
- `client/vitest.config.ts`
- `client/src/lib/components/blocks/SeatMap.test.ts`
- `client/src/routes/(auth)/login/+page.test.ts`
- `client/src/routes/(auth)/register/+page.test.ts`
- `client/src/lib/components/layouts/Navbar.test.ts`

**Dependencies to add**:
- `pnpm add -D vitest @testing-library/svelte jsdom`

---

### 3.4 E2E Tests (Cypress)
- [ ] `TASK-171` Install and configure Cypress
- [ ] `TASK-172` Write E2E test: Registration → Login → Search → Book → Checkout → Payment mock → Confirmation
- [ ] `TASK-173` Write E2E test: Operator flow → Create org → Add vehicle → Schedule trip
- [ ] `TASK-174` Write E2E test: Booking cancellation → Refund → Seat release
- [ ] `TASK-175` Write E2E test: Queue flow (simulate high load → Wait → Admission)
- [ ] `TASK-176` Write E2E test: Coupon application at checkout

**Files to create**:
- `client/cypress.config.ts`
- `client/cypress/e2e/booking.cy.ts`
- `client/cypress/e2e/operator_flow.cy.ts`
- `client/cypress/e2e/cancellation.cy.ts`
- `client/cypress/e2e/coupon.cy.ts`

**Dependencies to add**:
- `pnpm add -D cypress`

---

## 🟢 PHASE 4: Observability (Weeks 12-13)

### 4.1 Distributed Tracing
- [ ] `TASK-177` Add OpenTelemetry SDK package to Go workspace
- [ ] `TASK-178` Create `server/pkg/otel/tracer.go` (tracer initialization)
- [ ] `TASK-179` Create `server/pkg/otel/middleware.go` (trace injection in HTTP/gRPC)
- [ ] `TASK-180` Add trace propagation headers to all gRPC client calls
- [ ] `TASK-181` Add trace propagation headers to all HTTP client calls
- [ ] `TASK-182` Deploy Jaeger or Grafana Tempo via Docker Compose
- [ ] `TASK-183` Add trace ID injection into structured logs

---

### 4.2 Metrics & Dashboards
- [ ] `TASK-184` Add Prometheus client package to Go workspace
- [ ] `TASK-185` Create `server/pkg/metrics/metrics.go` (common metrics setup)
- [ ] `TASK-186` Add `http_requests_total` metric to all services
- [ ] `TASK-187` Add `http_request_duration_seconds` metric (p50, p95, p99) to all services
- [ ] `TASK-188` Add `grpc_requests_total` metric to all gRPC services
- [ ] `TASK-189` Add `db_query_duration_seconds` metric to all services with DB
- [ ] `TASK-190` Add `cache_hit_ratio` metric to services using Redis
- [ ] `TASK-191` Add `saga_execution_duration_seconds` metric to Order Service
- [ ] `TASK-192` Add `queue_size` metric to Queue Service
- [ ] `TASK-193` Deploy Prometheus via Docker Compose
- [ ] `TASK-194` Deploy Grafana via Docker Compose
- [ ] `TASK-195` Create Grafana dashboard: Service Health Overview
- [ ] `TASK-196` Create Grafana dashboard: Booking Flow Performance
- [ ] `TASK-197` Create Grafana dashboard: Database Performance
- [ ] `TASK-198` Create Grafana dashboard: Queue Metrics
- [ ] `TASK-199` Create Grafana dashboard: Error Rate & Alerts

**Files to create**:
- `server/pkg/metrics/metrics.go`
- `infra/grafana/dashboards/service_health.json`
- `infra/grafana/dashboards/booking_flow.json`
- `infra/grafana/dashboards/database.json`
- `infra/grafana/dashboards/queue_metrics.json`
- `infra/grafana/dashboards/error_rate.json`

---

### 4.3 Alerting
- [ ] `TASK-200` Configure Alertmanager with notification rules
- [ ] `TASK-201` Set up alert: High error rate (>5% of requests)
- [ ] `TASK-202` Set up alert: High latency (p95 > 500ms)
- [ ] `TASK-203` Set up alert: Service down (health check failure)
- [ ] `TASK-204` Set up alert: Queue overflow (>10,000 waiting)
- [ ] `TASK-205` Set up alert: Payment failure rate (>10%)
- [ ] `TASK-206` Set up alert: Saga failure rate (>2%)
- [ ] `TASK-207` Integrate Alertmanager with email/Slack

**Files to create**:
- `infra/alertmanager/alertmanager.yml`
- `infra/prometheus/alerts.yml`

---

## 🔵 PHASE 5: DevOps & CI/CD (Weeks 14-15)

### 5.1 CI/CD Pipeline
- [ ] `TASK-208` Create `.github/workflows/ci.yml` (lint → build → unit test → security scan)
- [ ] `TASK-209` Create `.github/workflows/cd.yml` (build Docker → push → deploy staging)
- [ ] `TASK-210` Create `.github/workflows/pr_preview.yml` (deploy preview env per PR)
- [ ] `TASK-211` Add merge block on CI failure
- [ ] `TASK-212` Add manual approval gate for production deployment

**Files to create**:
- `.github/workflows/ci.yml`
- `.github/workflows/cd.yml`
- `.github/workflows/pr_preview.yml`

---

### 5.2 Kubernetes Deployment
- [ ] `TASK-213` Create Helm chart structure (`Chart.yaml`, `values.yaml`)
- [ ] `TASK-214` Create Kubernetes deployment templates for all 19 services
- [ ] `TASK-215` Create HPA configs (Gateway 3-10, Inventory 3-15, others 1-3)
- [ ] `TASK-216` Create PDB (Pod Disruption Budget) configs
- [ ] `TASK-217` Create Network Policies (service-to-service access control)
- [ ] `TASK-218` Create Ingress config for Gateway
- [ ] `TASK-219` Create ConfigMaps for environment variables
- [ ] `TASK-220` Create Secrets template for sensitive data

**Files to create**:
- `infra/helm/travio/Chart.yaml`
- `infra/helm/travio/values.yaml`
- `infra/helm/travio/templates/gateway-deployment.yaml`
- `infra/helm/travio/templates/identity-deployment.yaml`
- (one per service deployment)
- `infra/k8s/hpa.yml`
- `infra/k8s/network-policy.yml`
- `infra/k8s/pdb.yml`

---

### 5.3 Database Migrations
- [ ] `TASK-221` Set up `golang-migrate` or `goose` in Go workspace
- [ ] `TASK-222` Create migration for Identity Service (users, orgs, refresh_tokens, invites)
- [ ] `TASK-223` Create migration for Catalog Service (stations, routes, trips)
- [ ] `TASK-224` Create migration for Order Service (orders, saga_instances, checkouts)
- [ ] `TASK-225` Create migration for Payment Service (transactions, refunds, configs)
- [ ] `TASK-226` Create migration for Fulfillment Service (tickets)
- [ ] `TASK-227` Create migration for Events Service (venues, events, ticket_types)
- [ ] `TASK-228` Create migration for Fleet Service (assets)
- [ ] `TASK-229` Create migration for CRM Service (coupons, support_tickets)
- [ ] `TASK-230` Create migration for Subscription Service (plans, subscriptions, invoices)
- [ ] `TASK-231` Create migration for Audit Service (audit_logs)
- [ ] `TASK-232` Add migration execution to CI/CD pipeline

**Files to create**:
- `server/migrations/identity/001_initial.sql`
- `server/migrations/catalog/001_initial.sql`
- `server/migrations/order/001_initial.sql`
- `server/migrations/payment/001_initial.sql`
- `server/migrations/fulfillment/001_initial.sql`
- `server/migrations/events/001_initial.sql`
- `server/migrations/fleet/001_initial.sql`
- `server/migrations/crm/001_initial.sql`
- `server/migrations/subscription/001_initial.sql`
- `server/migrations/audit/001_initial.sql`

---

## 🟣 PHASE 6: Security Hardening (Weeks 16-17)

### 6.1 API Security
- [ ] `TASK-233` Add CSRF protection middleware for cookie-based auth
- [ ] `TASK-234` Add request body size limits (prevent DoS)
- [ ] `TASK-235` Add input sanitization middleware (prevent XSS)
- [ ] `TASK-236` Review all SQL queries for parameterized statements (prevent injection)
- [ ] `TASK-237` Add rate limiting per user for authenticated endpoints
- [ ] `TASK-238` Implement API versioning strategy review (`/v1/`, `/v2/`)

---

### 6.2 Data Security
- [ ] `TASK-239` Implement AES-256 encryption for PII fields (NID, phone, email)
- [ ] `TASK-240` Add data masking in logs (never log passwords, tokens, NIDs)
- [ ] `TASK-241` Implement GDPR data deletion endpoint (`DELETE /v1/users/{userId}`)
- [ ] `TASK-242` Add data retention policies (auto-delete old bookings)

---

### 6.3 Infrastructure Security
- [ ] `TASK-243` Enable mTLS for all service-to-service communication
- [ ] `TASK-244` Set up certificate rotation (cert-manager or manual)
- [ ] `TASK-245` Remove hardcoded secrets (Queue Service token secret)
- [ ] `TASK-246` Set up WAF for Gateway (Cloudflare or AWS Shield)
- [ ] `TASK-247` Enable Kubernetes Pod Security Standards (restricted)

---

### 6.4 Payment Security
- [ ] `TASK-248` PCI-DSS compliance review (ensure card data never touches servers)
- [ ] `TASK-249` Add payment amount validation (prevent negative/overflow)
- [ ] `TASK-250` Implement payment reconciliation dashboard
- [ ] `TASK-251` Add fraud rule: flag mismatched NID + name
- [ ] `TASK-252` Add velocity checks: max bookings per user per hour

---

## ⚪ PHASE 7: Performance & Scalability (Weeks 18-19)

### 7.1 Database Optimization
- [ ] `TASK-253` Tune database connection pooling (max connections, idle timeout)
- [ ] `TASK-254` Add read replicas for Catalog and Search services
- [ ] `TASK-255` Add missing indexes (review slow query logs)
- [ ] `TASK-256` Add covering indexes for frequent queries
- [ ] `TASK-257` Review and fix N+1 query patterns
- [ ] `TASK-258` Implement database partitioning (orders table by date)

---

### 7.2 Caching Strategy
- [ ] `TASK-259` Implement Redis Cluster (instead of single-node)
- [ ] `TASK-260` Add L1 in-memory cache (sync.Map) for hot reads
- [ ] `TASK-261` Implement cache warming on deployment
- [ ] `TASK-262` Add cache invalidation events (on trip/station update)

---

### 7.3 API Performance
- [ ] `TASK-263` Add response compression (gzip/brotli)
- [ ] `TASK-264` Implement request coalescing (singleflight) for repeated reads
- [ ] `TASK-265` Add pagination to all list endpoints (review gaps)

---

### 7.4 Frontend Performance
- [ ] `TASK-266` Implement code splitting review (SvelteKit automatic route-based)
- [ ] `TASK-267` Add image optimization (`@sveltejs/enhanced-img`)
- [ ] `TASK-268` Implement service worker for caching static assets
- [ ] `TASK-269` Add lazy loading for heavy components (seat maps, charts)
- [ ] `TASK-270` Optimize bundle size (tree-shake unused code)

---

## 📋 PHASE 8: Documentation & Runbooks (Week 20)

### 8.1 API Documentation
- [ ] `TASK-271` Complete OpenAPI/Swagger spec (`docs/api/rest/openapi.yaml`)
- [ ] `TASK-272` Add gRPC API documentation (`docs/api/grpc.md`)
- [ ] `TASK-273` Set up Swagger UI for REST APIs

---

### 8.2 Operational Runbooks
- [ ] `TASK-274` Write runbook: Booking Flow Failure Investigation
- [ ] `TASK-275` Write runbook: Queue Overflow Response
- [ ] `TASK-276` Write runbook: Payment Gateway Outage
- [ ] `TASK-277` Write runbook: Database Failover
- [ ] `TASK-278` Write runbook: Service Deployment Rollback
- [ ] `TASK-279` Write runbook: Data Breach Response
- [ ] `TASK-280` Write runbook: Flash Sale Preparation

**Files to create**:
- `docs/runbooks/booking-failure.md`
- `docs/runbooks/queue-overflow.md`
- `docs/runbooks/payment-outage.md`
- `docs/runbooks/db-failover.md`
- `docs/runbooks/deployment-rollback.md`
- `docs/runbooks/data-breach.md`
- `docs/runbooks/flash-sale.md`

---

### 8.3 Developer Documentation
- [ ] `TASK-281` Write `CONTRIBUTING.md` (development setup, coding standards, PR process)
- [ ] `TASK-282` Write `docs/ARCHITECTURE.md` (detailed architecture diagrams)
- [ ] `TASK-283` Write `docs/LOCAL_DEVELOPMENT.md` (step-by-step local setup guide)
- [ ] `TASK-284` Add inline code comments for complex distributed systems logic
- [ ] `TASK-285` Write `docs/ONBOARDING.md` (new team member guide)

**Files to create**:
- `CONTRIBUTING.md`
- `docs/ARCHITECTURE.md`
- `docs/LOCAL_DEVELOPMENT.md`
- `docs/ONBOARDING.md`

---

## 🚀 PHASE 9: Production Readiness (Weeks 21)

### 9.1 Load Testing
- [ ] `TASK-286` Execute load test: Identity Service (10K login req/s)
- [ ] `TASK-287` Execute load test: Inventory Service (50K seat checks/s)
- [ ] `TASK-288` Execute load test: Queue Service (100K concurrent users)
- [ ] `TASK-289` Execute load test: Gateway (20K req/s)
- [ ] `TASK-290` Document baseline metrics (p50, p95, p99, error rate)
- [ ] `TASK-291` Fix identified bottlenecks
- [ ] `TASK-292` Set up automated load testing in CI (weekly)

---

### 9.2 Chaos Engineering
- [ ] `TASK-293` Test: Kill random service → verify circuit breaker opens
- [ ] `TASK-294` Test: Kill Redis → verify graceful degradation
- [ ] `TASK-295` Test: Kill Kafka → verify outbox pattern recovery
- [ ] `TASK-296` Test: Kill Postgres → verify health checks and auto-reconnect
- [ ] `TASK-297` Test: Network partition → verify mTLS and retry logic
- [ ] `TASK-298` Document failure mode behavior

---

### 9.3 Pre-Launch Checklist
- [ ] `TASK-299` All critical paths have E2E tests passing
- [ ] `TASK-300` Load tests meet performance targets
- [ ] `TASK-301` Security audit completed (no critical/high vulnerabilities)
- [ ] `TASK-302` Penetration testing completed
- [ ] `TASK-303` Monitoring dashboards operational
- [ ] `TASK-304` Alerting configured and tested
- [ ] `TASK-305` Runbooks written and reviewed
- [ ] `TASK-306` Database backups configured (automated, tested restore)
- [ ] `TASK-307` Disaster recovery plan documented
- [ ] `TASK-308` SSL certificates provisioned
- [ ] `TASK-309` Domain DNS configured
- [ ] `TASK-310` CDN configured for static assets
- [ ] `TASK-311` Error tracking set up (Sentry)
- [ ] `TASK-312` Analytics set up (privacy-compliant)

---

## 📊 PROGRESS TRACKING

### By Phase
| Phase | Total Tasks | Completed | Percentage |
|-------|-------------|-----------|------------|
| Phase 1: Critical Backend | 63 | 63 | 100% ✅ |
| Phase 2: Frontend | 76 | 0 | 0% |
| Phase 3: Testing | 30 | 0 | 0% |
| Phase 4: Observability | 23 | 0 | 0% |
| Phase 5: DevOps & CI/CD | 10 | 0 | 0% |
| Phase 6: Security | 20 | 0 | 0% |
| Phase 7: Performance | 18 | 0 | 0% |
| Phase 8: Documentation | 12 | 0 | 0% |
| Phase 9: Production Readiness | 14 | 0 | 0% |
| **TOTAL** | **312** | **63** | **20.2%** |

### ✅ PHASE 1 COMPLETE! (63 of 63 tasks)

**Phase 1.1: Checkout Service (TASK-001 to TASK-007)** ✅
- Domain model, repository, service, HTTP handler
- Pricing integration, coupon validation
- 6 endpoints: Create, Get, Update, Confirm, List, GetByHoldID

**Phase 1.2: Saga Orchestrator Wiring (TASK-008 to TASK-013)** ✅
- DLQ producer wired to Kafka
- Saga state recovery on startup
- Saga retry endpoint
- Order service updated with checkout repo

**Phase 1.3: Payment Webhooks (TASK-014 to TASK-023)** ✅
- IPN handler for SSLCommerz, bKash, Nagad
- Redis idempotency for duplicate prevention
- Kafka event publishing on payment status change
- Refund domain model + repository
- Refund status endpoint

**Phase 1.4: Fulfillment Passenger Data (TASK-024 to TASK-027)** ✅
- Order event publisher includes full passenger list
- Fulfillment consumer parses passenger details
- Validation: reject events with missing passenger data
- Real passenger data used for ticket generation

**Phase 1.5: Audit Service (TASK-028 to TASK-036)** ✅
- Complete service from scratch (was empty)
- Proto file created + generated
- Write-only audit log repository
- gRPC + HTTP handlers
- Data retention service (90-day auto-archive)
- Kafka consumer for audit events

**Phase 1.6: Notification Providers (TASK-037 to TASK-045)** ✅
- SMTP provider with TLS, retries, attachments
- Twilio provider for SMS
- Environment variable configuration
- Health checks for both providers

**Phase 1.7: CRM Coupon Integration (TASK-046 to TASK-052)** ✅
- Coupon code field added to CreateOrderRequest
- Order Service calls CRM ValidateCoupon before creating order
- Discount applied to order total
- Coupon UsageCount incremented on successful order
- Support ticket lifecycle with status transitions
- Agent assignment and SLA tracking

**Phase 1.8: Subscription Billing (TASK-053 to TASK-058)** ✅
- Billing scheduler (cron job) implemented
- Invoice generation on billing cycle end
- Payment Service integration for automatic charges
- Failed payment handling (retry → grace period → suspend)
- Proration on plan upgrades/downgrades
- Auto-renewal logic

**Phase 1.9: Reporting Endpoints (TASK-059 to TASK-063)** ✅
- GetRevenueReport (daily/monthly breakdown)
- GetOccupancyReport (seat occupancy rates)
- GetUserActivityReport (bookings per user)
- GetRoutePerformanceReport (top routes)
- CSV/JSON export for each report type

**Compilation Status**: ✅ ALL SERVICES COMPILE
- order ✅
- payment ✅
- audit ✅
- notification ✅
- fulfillment ✅
- subscription ✅
- reporting ✅
- identity ✅
- catalog ✅
- inventory ✅
- pricing ✅

### By Priority
| Priority | Tasks | Description |
|----------|-------|-------------|
| 🔴 P0 (Blocker) | TASK-001 to TASK-072 | Booking flow completion |
| 🟠 P1 (Critical) | TASK-073 to TASK-156 | Missing pages + testing |
| 🟡 P2 (Important) | TASK-157 to TASK-220 | Observability + DevOps |
| 🟢 P3 (Should Have) | TASK-221 to TASK-270 | Security + Performance |
| 🔵 P4 (Nice to Have) | TASK-271 to TASK-312 | Docs + Production readiness |

---

**How to use this file**: When working on Travio, find the next unchecked task in priority order. Complete it, then mark `[x]`. This file serves as the single source of truth for project completion progress.

**Last Updated**: April 7, 2026

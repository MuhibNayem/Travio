# Nationwide Ticketing System - APIs and Schemas

This document outlines API surfaces and data schemas. Internal calls are gRPC.
External calls are REST/GraphQL. Kafka events are protobuf-encoded.

## 1) API Conventions

- Idempotency: Use `Idempotency-Key` header for write endpoints.
- Correlation IDs: `X-Request-ID` and `X-Trace-ID`.
- Auth: JWT or opaque tokens with introspection.
- Pagination: `page`, `page_size`, `next_token`.

## 2) Service APIs (Selected Endpoints)

### 2.0 Subscription Service (SaaS Core)

REST:
- POST `/v1/subscriptions/plans` (Admin)
- GET `/v1/subscriptions/plans`
- POST `/v1/subscriptions/checkout`
- POST `/v1/subscriptions/webhook` (Stripe/Payment Provider)

### 2.1 Identity Service

REST:
- POST `/v1/auth/register`
- POST `/v1/auth/login`
- POST `/v1/auth/mfa/verify`
- POST `/v1/auth/refresh`
- GET `/v1/users/{id}`

### 2.2 Catalog Service

REST:
- POST `/v1/events`
- GET `/v1/events/{id}`
- GET `/v1/events?geo=lat,lng&radius=km&date=...`
- POST `/v1/venues`
### 2.2 Catalog Service (Transport Enabled)

REST:
- POST `/v1/events` (Generic)
- POST `/v1/trips` (Transport)
- POST `/v1/routes`
- GET `/v1/trips?from=station_id&to=station_id&date=...`
- POST `/v1/venues` (Stations)


### 2.3 Inventory Service

gRPC:
- ReserveTickets(event_id, seats[], hold_ttl)
- ReleaseHold(hold_id)
- ConfirmAllocation(order_id, hold_id)

REST:
- GET `/v1/inventory/{event_id}`

### 2.4 Pricing and Promotions Service

gRPC:
- QuotePrice(event_id, seat_ids[], promo_code)

REST:
- POST `/v1/promotions`
- GET `/v1/promotions/{code}`

### 2.5 Order Service

REST:
- POST `/v1/orders/draft`
- POST `/v1/orders/confirm`
- POST `/v1/orders/{id}/cancel`
- GET `/v1/orders/{id}`

### 2.6 Payment Service

REST:
- POST `/v1/payments/authorize`
- POST `/v1/payments/capture`
- POST `/v1/payments/refund`

### 2.7 Fulfillment Service

REST:
- POST `/v1/tickets/issue`
- POST `/v1/tickets/reissue`
- POST `/v1/tickets/transfer`
- GET `/v1/tickets/{id}`

### 2.8 Fraud Service

gRPC:
- ScoreTransaction(order_id, user_id, signals)

REST:
- POST `/v1/risk/rules`
- GET `/v1/risk/blocks`

### 2.9 Notification Service

REST:
- POST `/v1/notifications/email`
- POST `/v1/notifications/sms`
- POST `/v1/notifications/webhook`

### 2.10 Search Service

REST:
- GET `/v1/search/events?q=&filters=...`

### 2.11 Reporting Service

REST:
- GET `/v1/reports/sales?event_id=`
- GET `/v1/reports/finance?range=`

## 3) Kafka Topics and Events

Topics:
- `catalog.event.created`
- `inventory.hold.created`
- `inventory.hold.expired`
- `order.created`
- `order.confirmed`
- `payment.authorized`
- `payment.captured`
- `fulfillment.issued`
- `fraud.score.updated`

Event rules:
- Events must be immutable and versioned.
- Use schema registry to prevent breaking changes.
- Include `event_id`, `timestamp`, `trace_id`, `source`.

## 4) Data Schemas (Core & Transport)

### 4.0 SaaS Multi-Tenancy (Common PostgreSQL)

Table: `organizations` (Tenants)
- id (UUID, PK)
- name (VARCHAR)
- slug (VARCHAR, UNIQUE)
- status (ENUM: active, suspended, trial)
- created_at (TIMESTAMP)

Table: `subscriptions`
- id (UUID, PK)
- organization_id (UUID, FK)
- plan_id (VARCHAR)
- status (ENUM: active, past_due, canceled)
- current_period_end (TIMESTAMP)
- stripe_subscription_id (VARCHAR)

### 4.1 PostgreSQL (Orders and Payments)

Table: `orders`
- id (UUID, PK)
- user_id (UUID, FK)
- event_id (UUID)
- status (ENUM: draft, confirmed, canceled, refunded)
- total_amount (DECIMAL)
- currency (CHAR(3))
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
- idempotency_key (VARCHAR)

Table: `order_items`
- id (UUID, PK)
- order_id (UUID, FK)
- ticket_id (UUID)
- seat_id (VARCHAR)
- price (DECIMAL)
- fee (DECIMAL)

Table: `payments`
- id (UUID, PK)
- order_id (UUID, FK)
- gateway (VARCHAR)
- status (ENUM: authorized, captured, failed, refunded)
- amount (DECIMAL)
- currency (CHAR(3))
- token_ref (VARCHAR)
- created_at (TIMESTAMP)

### 4.2 ScyllaDB (Inventory and Tickets)

Table: `inventory_by_event`
- partition key: event_id
- clustering: section_id, seat_id
- columns: status (available, held, sold), price_tier, hold_id, hold_expires_at

Table: `inventory_by_trip_segment`
- partition key: trip_id
- clustering: segment_idx, seat_id
- columns: status (available, held, sold), hold_id, hold_expires_at
- note: segment_idx represents the leg (e.g., 0 for A->B, 1 for B->C)


Table: `holds_by_id`
- partition key: hold_id
- columns: event_id, user_id, seat_ids, expires_at (TTL)

Table: `tickets_by_id`
- partition key: ticket_id
- columns: order_id, event_id, seat_id, status, barcode, issued_at

### 4.3 Redis Keys (Examples)

- `rate:ip:{ip}` => counter with TTL
- `session:{session_id}` => session data
- `queue:token:{token}` => admission token

## 5) Example Protobuf (Inventory Hold Event)

```
syntax = "proto3";

message InventoryHoldCreated {
  string event_id = 1;
  string hold_id = 2;
  string user_id = 3;
  repeated string seat_ids = 4;
  int64 expires_at_unix = 5;
  string trace_id = 6;
}
```

## 6) Vendor/Partner Schema (PostgreSQL)

Table: `vendors`
- id (UUID, PK)
- name (VARCHAR)
- status (ENUM: pending, verified, active, suspended)
- contact_email (VARCHAR)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

Table: `vendor_kyc`
- id (UUID, PK)
- vendor_id (UUID, FK)
- status (ENUM: pending, approved, rejected)
- provider_ref (VARCHAR)
- reviewed_at (TIMESTAMP)

Table: `vendor_contracts`
- id (UUID, PK)
- vendor_id (UUID, FK)
- version (VARCHAR)
- accepted_at (TIMESTAMP)
- revenue_share (DECIMAL)

Table: `vendor_payouts`
- id (UUID, PK)
- vendor_id (UUID, FK)
- provider (VARCHAR)
- payout_account_ref (VARCHAR)
- status (ENUM: pending, verified, blocked)

Table: `vendor_audit_events`
- id (UUID, PK)
- vendor_id (UUID, FK)
- event_type (VARCHAR)
- payload (JSONB)
- created_at (TIMESTAMP)

# Nationwide Ticketing System - Delivery Roadmap

Assumes 2-week sprints with parallel squads (Platform, Core, Data, Risk).
Dependencies listed per sprint. Adjust based on team size.

## Sprint 0 (Setup)

Deliverables:
- Repo scaffolding and Go service template.
- IaC base modules for VPC, Kubernetes, Kafka, Postgres.
- CI/CD pipeline and baseline observability.

Dependencies:
- Cloud accounts, IAM, and network approvals.

## Detailed Service-to-Service Implementation Plan (Immediate Execution)

This section details the strict order of operations for the "Strict Isolation" build phase.

### Phase 1: Foundation & Identity (The "Who")
1.  **Identity Service**: Implement `Register`, `Login`.
2.  **API Gateway**: Configure basic routing to Identity.
3.  **Authentication**: Verify JWT issuance and middleware in `pkg/server`.

### Phase 2: Core Data (The "What")
1.  **Catalog Service**: Implement `CreateStation`, `CreateRoute`, `CreateTrip`.
2.  **Data Seeding**: Script to populate initial routes (Dhaka->Chittagong).
3.  **Public API**: `GET /trips` (Search) endpoint.

### Phase 3: Inventory & Availability (The "How Many")
1.  **Inventory Service**: Implement `InitializeInventory` (allocating seats to segments).
2.  **Availability Logic**: `GetAvailability` (checking segments for a trip).
3.  **Holds**: `HoldSeat` with TTL (Redis/ScyllaDB).

### Phase 4: Transactions (The "Deal")
1.  **Order Service**: `CreateOrder` (Pending state).
2.  **Saga Pattern**: Orchestrate `Inventory.Reserve` -> `Payment.Charge` -> `Order.Confirm`.
3.  **Payment Service**: Mock gateway integration.

### Phase 5: Result (The "Ticket")
1.  **Fulfillment Service**: Listen to `OrderConfirmed` events, generate QR.
2.  **Notification Service**: Send email/SMS stub.

---

## Sprint 1 (Identity + Catalog MVP)

Deliverables:
- Identity Service (register/login/MFA).
- Catalog Service CRUD (events, venues).
- API Gateway routing and auth integration.

Dependencies:
- Sprint 0 infrastructure.

## Sprint 2 (Inventory MVP)

Deliverables:
- Inventory Service with holds and TTL in ScyllaDB.
- Basic seat map endpoints and data ingestion.
- Inventory API contracts and load tests.

Dependencies:
- Kafka and ScyllaDB provisioned.

## Sprint 3 (Order and Checkout)

Deliverables:
- Order Service draft/confirm flows.
- Idempotency key support.
- Order events published to Kafka.

Dependencies:
- Inventory holds API stable.

## Sprint 4 (Payments)

Deliverables:
- Payment Service with gateway integration.
- PCI boundary and tokenization.
- Fraud Service basic velocity checks.

Dependencies:
- Order Service confirm flow.

## Sprint 5 (Fulfillment + Notifications)

Deliverables:
- Ticket issuance and QR generation.
- Email/SMS confirmation via Notification Service.
- Audit Service initial logging.

Dependencies:
- Payment capture events.

## Sprint 6 (Search and Discovery)

Deliverables:
- Search Service index pipeline.
- Catalog indexing and faceted search.

Dependencies:
- Catalog service stable schemas.

## Sprint 7 (Queue + Rate Limiting)

Deliverables:
- Queue Service admission tokens.
- Edge rate limiting and load shedding.
- Peak load rehearsal.

Dependencies:
- API Gateway and Redis.

## Sprint 8 (Multi-Region Rollout)

Deliverables:
- Active-active routing.
- ScyllaDB multi-DC replication.
- Kafka cross-region replication.

Dependencies:
- Stable core services, DR strategy.

## Sprint 9 (Promotions + Dynamic Pricing)

Deliverables:
- Pricing/Promotions Service.
- Experimentation and feature flag integration.

Dependencies:
- Catalog and Order services.

## Sprint 10 (Reporting and Analytics)

Deliverables:
- Sales reports and exports.
- Stream processing jobs for aggregates.

Dependencies:
- Kafka event stability.

## Sprint 11 (Hardening and Compliance)

Deliverables:
- PCI and GDPR audits.
- Chaos testing and full runbook.
- Performance tuning and SLO dashboards.

Dependencies:
- All core services functional.


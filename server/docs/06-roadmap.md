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


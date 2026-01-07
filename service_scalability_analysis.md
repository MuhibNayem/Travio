# Service Scalability Analysis (FAANG Benchmark)

## Overview
This document analyzes the current state of Travio's backend services against high-scale ("FAANG") requirements: **100k+ concurrents**, **strict consistency**, **p95 < 200ms**, and **resiliency**.

## 1. Inventory Service
**Criticality**: Highest (Hot Path)
- **Current State**: 
    - **Database**: Uses ScyllaDB with `gocql`.
    - **Consistency**: Implements Lightweight Transactions (LWT) (`IF status = ?`) for atomic seat holds. This is excellent for preventing overselling.
    - **Optimization**: Uses `LoggedBatch` for multi-segment updates.
    - **Saga Participation**: Not explicitly seen yet, but DB primitives exist.
- **FAANG Gaps**:
    - **Hot Row Contention**: LWTs on popular events might contend. Consider **Inventory Partitioning** (sharding inventory by section/row logical groups) if locking becomes a bottleneck.
    - **Caching**: [GetSeatAvailability](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/inventory/internal/repository/scylla.go#47-75) hits Scylla directly. Needs a **Redis Read-Through Cache** or **Client-Side Caching** for seat maps (static data).
    - **Concurrency control**: Current batch size in [InitializeTrip](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/inventory/internal/repository/scylla.go#22-46) might be too large for massive vehicles (trains/stadiums). Needs pagination/streaming.

## 2. Order Service
**Criticality**: High (Transactional Integrity)
- **Current State**:
    - **Architecture**: Implements a [Saga](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/order/internal/saga/orchestrator.go#29-41) pattern (`internal/saga`), which is correct for distributed transactions.
    - **Orchestration**: [orchestrator.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/order/internal/saga/orchestrator.go) likely manages state transitions.
- **FAANG Gaps**:
    - **Idempotency**: Needs strict enforcement key check at the edge (Gateway) and within `CreateOrder`.
    - **Dead Letter Queues (DLQ)**: Failed saga steps must route to a DLQ for manual intervention or automated retry/compensation.
    - **State Machine Persistence**: Saga state must be persisted (e.g., in Postgres or Redis) to survive service restarts *during* a transaction.

## 3. Gateway Service
**Criticality**: High (Entry Point)
- **Current State**: 
    - **Middlewares**: Implements [rate_limit.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/gateway/internal/middleware/rate_limit.go) and [queue.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/gateway/internal/middleware/queue.go).
    - **Rate Limiting**: Checks [rate_limit.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/gateway/internal/middleware/rate_limit.go) implementation (verify if Redis-backed or in-memory).
    - **Queue Integration**: [queue.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/gateway/internal/middleware/queue.go) middleware suggests simple integration.
- **FAANG Gaps**:
    - **Circuit Breaking**: No evidence of `gobreaker` or similar circuit breaking middleware in `internal/middleware`.
    - **Aggregated Request Merging**: Request collapsing (singleflight) for high-traffic read endpoints is missing.
    - **Edge Authentication caching**: If `identity` validation happens here, it needs a local cache for public keys/JWKS.

## 4. Identity Service
**Criticality**: Medium (Cached often)
- **Current State**: 
    - **Storage**: [postgres.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/identity/internal/repository/postgres.go) and [refresh_token.go](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/identity/internal/repository/refresh_token.go) indicate pure PostgreSQL usage.
    - **Token Management**: Refresh tokens stored in DB.
- **FAANG Gaps**:
    - **Token Blacklist/Revocation**: No Redis repository found for fast revocation checks.
    - **High-Read Caching**: User profiles and role data should be cached (Redis) to avoid hitting Postgres on every [GetUser](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/inventory/internal/repository/redis_hold.go#112-131) or [ValidateToken](file:///Users/a.k.mmuhibullahnayem/Developer/Travio/server/services/gateway/internal/middleware/queue.go#13-14) call.

## 5. Missing/New Services
- **Pricing**: Empty. Needs a rules engine (e.g., CEL or expr) to evaluate dynamic pricing without hardcoding logic.
- **Queue**: Empty. Needs a "Virtual Waiting Room" implementation using Redis Sorted Sets (`ZADD` + `ZRANK`) to throttle traffic before it hits the Gateway.

## 6. Catalog Service
**Criticality**: Medium (Read Heavy)
- **Current State**: Standard CRUD (`handler`, `service`, `repository`).
- **FAANG Gaps**:
    - **Caching**: No evidence of multi-level caching (L1 In-Memory, L2 Redis). Catalog data is highly cacheable and should rarely hit the DB.
    - **Search Integration**: No obvious hook to update Search Service indexes upon modification (`CDC` or `Outbox` needed).

## 7. Fulfillment Service
**Criticality**: High (User Delight)
- **Current State**: Robust implementation with `pdf` and `qr` generation packages.
- **FAANG Gaps**:
    - **Async Processing**: Ticket generation can be slow. Should be strictly decoupled via Kafka consumers (which seems to be there: `internal/consumer`).
    - **Storage**: PDF/QR assets should be offloaded to CDNs/S3, not served directly from service instances.

## 9. Notification Service
**Criticality**: Medium (Async)
- **Current State**: Implements `consumer` and `provider` packages. Decent abstraction.
- **FAANG Gaps**:
    - **Templates**: Likely hardcoded. Needs a template engine (e.g., using S3/DB for dynamic templates) to avoid redeploys for copy changes.
    - **Rate Limiting**: Sending emails/SMS needs strict rate limiting per provider (SES/Twilio) to avoid account suspension.

## 10. Payment Service
**Criticality**: Highest (Revenue)
- **Current State**: `gateway` package suggests multi-provider support.
- **FAANG Gaps**:
    - **PCI Isolation**: Card data handling should be strictly minimal.
    - **Idempotency**: Critical here. Ensure `payment_id` generation is deterministic based on `order_id` + `attempt`.
    - **Reconciliation**: Missing a reconciliation job/worker to verify gateway status vs. DB status.

## 11. Queue Service & Reporting Service
**Criticality**: High (Queue) / Low (Reporting)
- **Current State**: 
    - `queue`: Standard internal structure. Lacks advanced logic visible at surface.
    - `reporting`: Shell service (`cmd`). **NOT IMPLEMENTED**.
- **FAANG Gaps**:
    - **Queue Logic**: A true "Virtual Waiting Room" needs optimized Redis Lua scripts for atomic `ZADD`/`ZREM` operations, which aren't immediately obvious.
    - **Reporting**: Needs an ETL pipeline (e.g., Debezium -> Kafka Connect -> OLAP) rather than synchronous API calls which will kill the transactional DBs.

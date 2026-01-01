# Nationwide Ticketing System - Data Migrations and Partitioning

## 1) PostgreSQL Strategy

### 1.1 Partitioning

- Orders partitioned by `created_at` (monthly) for efficient archival.
- Order_items partitioned by `order_id` hash.
- Payments partitioned by `created_at` for reconciliation workloads.

Example (PostgreSQL declarative partitioning):
- `orders_2025_01`, `orders_2025_02`, ...
- `payments_2025_01`, `payments_2025_02`, ...

### 1.2 Indexing

- `orders (user_id, created_at)`
- `orders (event_id, status)`
- `payments (order_id)`
- `order_items (order_id)`

### 1.3 Migrations

- Use a migration tool (golang-migrate, goose).
- Migrations are immutable and versioned.
- Every migration includes up/down scripts.

### 1.4 Archival

- Move partitions older than N months to cold storage.
- Maintain read replicas for reporting.

## 2) ScyllaDB Strategy

### 2.1 Partitioning

- `inventory_by_event` partitioned by `event_id`.
- Clustering by `section_id` then `seat_id`.
- `inventory_by_trip_segment` partitioned by `trip_id`.
- Clustering by `segment_idx` then `seat_id`.
- `holds_by_id` partitioned by `hold_id` with TTL.
- `tickets_by_id` partitioned by `ticket_id`.

### 2.2 Consistency

- Use QUORUM for writes in inventory holds.
- Use LOCAL_QUORUM for read paths within region.
- Use TTL for hold expiration to avoid cleanup jobs.

### 2.3 Schema Evolution

- Add columns only (no breaking changes).
- Decommission columns after two release cycles.

## 3) Redis Strategy

- Keys prefixed by service (e.g. `inventory:`, `queue:`).
- TTLs for rate limiters and sessions.
- Use Redis Cluster for scaling.

## 4) Search Indexing

- Separate index per region.
- Rebuild indices asynchronously on schema changes.
- Use aliases for zero-downtime reindexing.

## 5) Migration Phases

Phase 1:
- Initialize PostgreSQL schemas.
- Create ScyllaDB tables for inventory.
- Seed Redis and OpenSearch.

Phase 2:
- Backfill catalog data into search indexes.
- Introduce new payment fields via additive migration.

Phase 3:
- Partition pruning and data archival pipeline.


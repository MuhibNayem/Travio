# Catalog Service

Manages Stations, Routes, and Trips.

## Scalability Features

-   **Read-Through Caching**:
    -   **Stations**: `GetByID` requests are cached in Redis (24h TTL).
    -   **Routes**: `GetByID` requests are cached in Redis (24h TTL).
    -   **Pattern**: Decorator Pattern (`CachedStationRepository` wraps `PostgresStationRepository`).
    -   Implementation: `internal/repository/redis.go`.

## Setup

```bash
# 1. Start dependencies
docker compose up -d postgres redis

# 2. Run Service
go run cmd/main.go
```

## Configuration
-   `REDIS_ADDR`: Redis address (default: `localhost:6379`)
-   `DB_HOST`: Postgres host (default: `localhost`)

## Verification

### Load Test
Simulates concurrent access to verify caching performance.

```bash
go run load_test/load.go
```

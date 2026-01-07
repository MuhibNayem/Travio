# Search Service

Full-text search for trips and stations using OpenSearch.

## Scalability Features

-   **Query Result Caching**:
    -   `SearchTrips`: Results cached in Redis (5m TTL).
    -   `SearchStations`: Results cached in Redis (1h TTL).
    -   **Pattern**: Cache key derived from query hash.
    -   Implementation: `internal/searcher/searcher.go`.

## Setup

```bash
# 1. Start dependencies
docker compose up -d opensearch kafka redis

# 2. Run Service
go run cmd/main.go
```

## Configuration
-   `REDIS_ADDR`: Redis address (default: `localhost:6379`)
-   `OPENSEARCH_URL`: OpenSearch URL (default: `http://localhost:9200`)

## Verification

### Load Test
Simulates repeated searches with identical queries.

```bash
go run load_test/load.go
```
First request hits OpenSearch; subsequent requests hit Redis cache.

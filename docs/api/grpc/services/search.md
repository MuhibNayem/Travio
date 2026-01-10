# Search Service gRPC Documentation

**Package:** `search.v1`  
**Internal DNS:** `search:9088`  
**Proto File:** `server/api/proto/search/v1/search.proto`

## Overview
The Search Service provides high-performance, fuzzy-search capabilities over the trip catalog. It syncs data from the authoritative `Catalog` service into **OpenSearch** for querying.

## Key Behaviors

### Search Engine (OpenSearch)
- **Indices:** Maintains `trips` and `stations` indices.
- **Query Logic:**
  - **Trips:** Uses strict filtering (`term`) for `from_station`, `to_station`, and `date` to ensure exact matches for itinerary validity. Uses `multi_match` for general query strings (Route Name, Operator Name).
  - **Stations:** Uses `fuzziness: AUTO` to handle typos in station names or locations.

### Caching Strategy
- **Deterministic Keys:** Cache keys are generated using a SHA256 hash of all query parameters.
- **TTL Policies:**
  - `Trips`: **5 minutes** (Balances availability accuracy with cache hit rate).
  - `Stations`: **1 hour** (Station data changes rarely).
- **Format:** Results are cached as compressed JSON blobs in Redis.

### Indexing Pipeline
- **Real-time Updates:** The `Indexer` component listens to Kafka events (`trip.created`, `trip.updated`) to upsert documents immediately (`Refresh=true`).
- **Initialization:** On startup, the service verifies index existence and creates them with default mapppings if missing.

---

## RPC Methods

### `SearchTrips`
Retrieves trips matching the criteria.

- **Request:** `SearchTripsRequest`
- **Response:** `SearchTripsResponse` (Includes `total` count for pagination).
- **Cache:** Yes (5m).

### `SearchStations`
Fuzzy searches for stations by name or location.

- **Request:** `SearchStationsRequest`
- **Response:** `SearchStationsResponse`.
- **Cache:** Yes (1h).

---

## Message Definitions

### SearchTripsRequest
| Field | Type | Description |
|-------|------|-------------|
| `query` | `string` | Optional text search (e.g., "Hanif Enterprise") |
| `from_station_id` | `string` | **Required** Origin UUID |
| `to_station_id` | `string` | **Required** Destination UUID |
| `date` | `string` | **Required** ISO8601 Date (`YYYY-MM-DD`) |
| `limit` | `int32` | Pagination limit |
| `offset` | `int32` | Pagination offset |

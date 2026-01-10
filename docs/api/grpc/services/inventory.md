# Inventory Service gRPC Documentation

**Package:** `inventory.v1`  
**Internal DNS:** `inventory:9083`  
**Proto File:** `server/api/proto/inventory/v1/inventory.proto`

## Overview
The Inventory Service manages the real-time availability of seats across all trips. It handles high-concurrency seat locking ("holds") and serves seat map visualizations.

## Key Behaviors

### Concurrency & Locking
- **Hybrid Locking Strategy:**
  1. **Optimistic Redis Lock (Fast):** Acquires a temporary (10s TTL) lock on all requested seat/segment pairs to prevent race conditions at the API layer.
  2. **ScyllaDB Persistent Hold (Authoritative):** If the Redis lock succeeds, a persistent record is created in ScyllaDB with the actual expiration (default 10 minutes).
- **Anti-Scalping:** Enforces a hard limit of **2 concurrent active holds** per user.

### Caching Strategy
- **Seat Maps:** Uses a **Read-Through** caching pattern.
  - The entire seat inventory for a trip is cached in Redis.
  - Cache TTL is short (**5 seconds**) to ensure near real-time accuracy for high-traffic searches while protecting the DB.
  - Cache invalidation occurs immediately upon successful `HoldSeats` or `ConfirmBooking` operations.
- **Segment Logic:** Availability is calculated dynamically based on the requested `fromStation` and `toStation`. A seat is only "Available" if it is free across **all** segments of the requested journey.

---

## RPC Methods

### `HoldSeats`
Temporarily reserves seats for a user to allow them to proceed to payment.

- **Request:** `HoldRequest`
- **Response:** `HoldResult`
  - `hold_id`: UUID used for booking confirmation.
  - `expires_at`: Timestamp (standard 10m window).
- **Errors:**
  - `MAX_HOLDS_EXCEEDED` (if user > 2 holds).
  - `SEAT_UNAVAILABLE` (if any segment is booked).

### `ConfirmBooking`
Finalizes a hold into a permanent booking.

- **Pre-condition:** Must provide a valid, non-expired `hold_id`.
- **Validation:** Passenger count must match the number of held seats.
- **Response:** `BookingResult` (allocates `ticket_id`s).

### `GetSeatMap`
Returns the visual layout of seats for a specific trip segment.

- **Request:** `GetSeatMapRequest`
- **Response:** `SeatMapResult` (Legend + Rows/Cells).

---

## Message Definitions

### HoldRequest
| Field | Type | Description |
|-------|------|-------------|
| `trip_id` | `string` | Target Trip |
| `seat_ids` | `string` | List of specific Seat UUIDs to lock |
| `hold_duration` | `int64` | Optional. Override default 10m (Admin only) |

### SeatStatus Enum
- `AVAILABLE`
- `HELD` (Locked by a user)
- `BOOKED` (Sold)
- `BLOCKED` (Operational block)

### SeatMapResult
Contains `rows` and `legend`.
- **Legend:**
  - `available`: `#00FF00`
  - `held`: `#FFFF00`
  - `booked`: `#FF0000`

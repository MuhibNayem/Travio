# Catalog Service gRPC Documentation

**Package:** `catalog.v1`  
**Internal DNS:** `catalog:9082`  
**Proto File:** `server/api/proto/catalog/v1/catalog.proto`

## Overview
The Catalog Service is the system of record for the core travel inventory static data: Stations, Routes, and Scheduled Trips. It strictly enforces SaaS plan limits for operators.

## Key Behaviors

### SaaS Limit Enforcement
- **Scheduling Horizon:** When creating trips, the service checks the Operator's subscription plan.
  - Example: A 'Basic' plan may only schedule trips 30 days out. Attempting to schedule for 60 days out returns `PLAN_LIMIT_EXCEEDED`.
- **Route Limits:** The number of active routes is capped by the plan tier.

### Trip Scheduling
- **Auto-Calculation:** `ArrivalTime` is automatically derived from the `DepartureTime` + Route's `EstimatedDuration`.
- **Validation:** Ensures Origin and Destination stations exist and are active.

### Data Model
- **Route:** A logical connection between A and B (e.g., "Dhaka to Chittagong").
- **Trip:** A concrete instance of a Route at a specific time (e.g., "Dhaka-CTG at 10:00 AM on Dec 25").

---

## RPC Methods

### `CreateTrip`
Schedules a new trip instance.

- **Request:** `CreateTripRequest`.
- **Validation:** Checks SaaS limits (`MaxScheduleDays`).
- **Response:** `Trip`.

### `SearchTrips`
Retrieves trips matching criteria (City-to-City).

- **Note:** This is the *Admin/Operator* search. Public user search is handled by the dedicated **Search Service**.

### `Metric Access`
Admin endpoints to manage Stations and Routes.

---

## Message Definitions

### Trip
| Field | Type | Description |
|-------|------|-------------|
| `route_id` | `string` | Link to parent route |
| `departure_time` | `string` | ISO8601 |
| `arrival_time` | `string` | Calculated automatically |
| `vehicle_type` | `string` | AC_BUS, NON_AC_BUS |

### Station
| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Display Name |
| `city` | `string` | Location Filter |
| `geo_lat` | `double` | Optional |
| `geo_long` | `double` | Optional |

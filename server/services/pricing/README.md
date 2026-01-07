# Pricing Service

Dynamic pricing engine with database-driven rules evaluation.

## Features

-   **Dynamic Rules Engine**: Uses `expr-lang/expr` for safe expression evaluation
-   **Database-Driven**: Rules stored in PostgreSQL, hot-reloadable
-   **Default Rules**:
    | Rule | Condition | Multiplier |
    |------|-----------|------------|
    | Weekend Surge | Saturday/Sunday | ×1.20 |
    | Early Bird | 30+ days ahead | ×0.85 |
    | Last Minute | <3 days | ×1.50 |
    | High Demand | >80% occupancy | ×1.25 |
    | Business Class | seat_class="business" | ×1.40 |

## API

### Calculate Price
```bash
POST /api/v1/pricing/calculate
{
  "trip_id": "TRIP-001",
  "seat_class": "economy",
  "date": "2026-01-11",
  "quantity": 2,
  "base_price_paisa": 100000,
  "occupancy_rate": 0.5
}
```

### Get Rules
```bash
GET /api/v1/pricing/rules
```

## Setup

```bash
# Create database
createdb travio_pricing

# Run service
go run cmd/main.go
```

## Configuration
-   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
-   `GRPC_PORT`: HTTP port (default: 50058)

## Verification

```bash
go run load_test/load.go
```

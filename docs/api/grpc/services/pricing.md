# Pricing Service gRPC Documentation

**Package:** `pricing.v1`  
**Internal DNS:** `pricing:50058`  
**Proto File:** `server/api/proto/pricing/v1/pricing.proto`

## Overview
The Pricing Service is a dynamic, rule-based engine that calculates final ticket prices. It allows operators to define granular rules (e.g., "Weekend Surcharge", "Last Minute Discount") without code changes.

## Key Behaviors

### Rules Engine Architecture
- **In-Memory Evaluation:** Rules are loaded from the database into highly optimized in-memory partitions for sub-millisecond calculation.
- **Update Propagation:** `RefreshRules` rebuilds the engine state on demand (zero-downtime updates).

### Hierarchy & Overrides
1. **Global Rules:** Apply to all organizations by default.
2. **Organization Rules:**
   - **Overrides:** If an Org defines a rule with the same name as a Global rule, the Org rule "wins" (O(1) override).
   - **Extensions:** New rules specific to an Org are added to the evaluation chain.

### Evaluation Context
Rules are evaluated against a rich context:
- `occupied_count` / `total_seats` (for dynamic yield management)
- `days_until_departure`
- `seat_class`
- `passenger_count`

---

## RPC Methods

### `CalculatePrice`
Computes the final price for a booking draft.

- **Request:** `CalculatePriceRequest` (includes `base_price`, `occupancy`, `org_id`).
- **Response:** `CalculatePriceResponse`
  - `final_price_paisa`: The computed amount.
  - `applied_rules`: List of rules that triggered, for transparency/receipts.

### `GetRules`, `CreateRule`, `UpdateRule`
CRUD operations for managing the rule definitions.

---

## Message Definitions

### PricingRule
| Field | Type | Description |
|-------|------|-------------|
| `condition` | `string` | CEL-like expression (e.g., `occupancy > 0.8 && request.days_until < 2`) |
| `multiplier` | `double` | Price modifier (e.g., `1.2` for +20%, `0.9` for -10%) |
| `priority` | `int32` | Evaluation order (Higher = Later) |

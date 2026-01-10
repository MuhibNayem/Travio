# Subscription Service gRPC Documentation

**Package:** `subscription.v1`  
**Internal DNS:** `subscription:50060`  
**Proto File:** `server/api/proto/subscription/v1/subscription.proto`

## Overview
The Subscription Service handles SaaS billing, plan management, and usage tracking for Tenant Organizations. It manages the lifecycle of subscriptions and calculates entitlements.

## Key Behaviors

### Entitlement Logic
- **Plan Enforcement:** Capabilities are tied to the Organization's `plan_id`.
- **Usage Tracking:** Features like "max_users" or "api_calls" are metered via `RecordUsage`.
- **Pre-Paid Model:** Subscriptions are treated as active immediately upon creation, generating a "PAID" initial invoice (simplified for MVP).

### Billing Cycles
- **Month:** Adds 1 calendar month.
- **Year:** Adds 1 calendar year.
- **Default:** Adds 30 days if interval is unrecognized.

### Invoicing
- **Auto-Generation:** An initial invoice is automatically created when a subscription is started.
- **Line Items:** Includes base plan fees and distinct usage overages (if applicable).

---

## RPC Methods

### `CreateSubscription`
Subscribes an organization to a plan.

- **Request:** `CreateSubscriptionRequest`
- **Response:** `Subscription`
- **Side Effect:** Generates a `paid` Invoice.

### `RecordUsage`
Meters a usage event for checking quota limits (e.g., specific feature usage).

- **Request:** `RecordUsageRequest` (includes `idempotency_key`).
- **Response:** `RecordUsageResponse`.

### `GetEntitlement`
Retrieves the current standing of an organization (Active/Past Due) and its quotas.

---

## Message Definitions

### Plan
| Field | Type | Description |
|-------|------|-------------|
| `interval` | `string` | `month`, `year` |
| `price_paisa` | `int64` | Base cost |
| `features` | `map<string,string>` | Key-value feature flags |

### Invoice
| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | `paid`, `open`, `void` |
| `line_items` | `[]LineItem` | Breakdown of costs |

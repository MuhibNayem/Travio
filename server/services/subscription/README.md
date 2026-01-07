# Subscription Service

Handles Subscriptions, Plans, Invoices, and Usage-Based Billing for Organizations.

## Features

-   **Plan Management**: Create flexible plans with base fees and features.
-   **Subscription Lifecycle**: Create, Cancel, and track status (Active, Past Due).
-   **Usage-Based Billing**: Automatically tracks usage (e.g., ticket sales) and adds line items to invoices.
    -   **Idempotency**: `RecordUsage` RPC handles duplicate events gracefully.
    -   **Dynamic Commission**: Plans support `usage_price_paisa` for per-unit fees.
-   **Invoicing**: Generates detailed JSON-based invoices with line items.

## API Usage

### Billing & Usage

**Record Usage (Internal RPC)**
Called by `Order Service` when a ticket is sold.
```protobuf
rpc RecordUsage(RecordUsageRequest) returns (RecordUsageResponse);
```

**Get Invoices**
```http
GET /v1/invoices?subscription_id={id}
```

## Architecture

### Usage Tracking Flow
1.  **Order Confirmed**: Order Service calls `SubscriptionService.RecordUsage`.
2.  **Event Logged**: `usage_events` table stores the event with a unique `idempotency_key` (OrderID).
3.  **Invoice Generation**: Background job (or on-demand) sums `usage_events` + Base Fee -> Creates `Invoice`.
4.  **Audit**: Invoices store a snapshot of charges in `line_items` (JSONB).

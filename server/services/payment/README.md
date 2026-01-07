# Payment Service

Reliable payment processing service with Idempotency and Reconciliation.

## Features

-   **Idempotency**: Prevents duplicate charges for the same Order using deterministic hash keys (OrderID + Attempt).
-   **Reconciliation**: Background worker (`reconciler.go`) periodically checks Gateway status for stuck `PENDING` transactions.

-   **Persistence**: PostgreSQL (`transactions` table) stores all attempt states.
-   **Organization-Owned Payments**: Dynamic Gateway Factory resolves credentials per Organization ID (`payment_configs` table).
-   **Admin Control**: Secure APIs (`PUT /payment-config`) for organizations to manage their own gateway credentials (SSLCommerz, bKash, Nagad).

## Architecture

### Dynamic Gateway Resolution
Payments are processed using the Organization's own credentials, not the Platform's.
1.  **Registry**: `Factory` interface loads config from DB.
2.  **Encryption**: Credentials stored securely (JSONB).
3.  **Isolation**: Each payment is sandboxed to the specific Organization context.

## Setup

```bash
# 1. Start dependencies
docker compose up -d postgres redis

# 2. Run Service
go run cmd/main.go
```

## Verification

### Load Test (Idempotency)
Simulates concurrent requests for the same Order. Only **one** should trigger a new Gateway call; others should return the cached pending/success state.

```bash
go run load_test/load.go
```

> **Note**: Requires PostgreSQL and Redis. If running locally without Docker, ensure `travio_payment` DB exists.

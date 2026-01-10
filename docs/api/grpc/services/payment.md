# Payment Service gRPC Documentation

**Package:** `payment.v1`  
**Internal DNS:** `payment:9085`  
**Proto File:** `server/api/proto/payment/v1/payment.proto`

## Overview
The Payment Service acts as an abstraction layer over multiple payment gateways (SSLCommerz, bKash, Nagad), providing a unified API for processing transactions. It handles multi-tenancy, ensuring each organization uses its own merchant credentials.

## Key Behaviors

### Multi-Tenancy
- **Dynamic Configuration:** Payment credentials (merchant IDs, secrets) are loaded dynamically based on the `organization_id` of the request.
- **Context Propagation:** Return and Cancel URLs are automatically enriched with `?org={org_id}` to ensure callbacks are routed to the correct tenant context.

### Idempotency
- **Deterministic Keys:** Transactions are de-duplicated using a UUIDv5 hash generated from the `OrderID` and `Attempt` count.
- **Strict Reuse:** If a `CreatePayment` request is repeated for an Order that is already `PENDING` or `SUCCESS`, the service returns the existing session info instead of creating a new one.

### Gateway Abstraction
- **Factory Pattern:** The service uses a provider registry to instantiate the correct gateway client at runtime.
- **Sandbox Mode:** Supports per-organization sandbox switching via the `is_active` / `is_sandbox` flags in the configuration.

---

## RPC Methods

### `CreatePayment`
Initiates a payment session with a third-party gateway.

- **Request:** `CreatePaymentRequest`
- **Behavior:** Persists a `PENDING` transaction and returns the gateway's redirect URL.
- **Response:** `PaymentResult` (contains `redirect_url`, `session_id`).

### `UpdatePaymentConfig`
(Admin) Updates the merchant credentials for a specific organization and gateway.

- **Request:** `UpdatePaymentConfigRequest`
- **Note:** Credentials are validated against the provider factory before saving.

### `GetPaymentStatus`
Retrieves the current status of a payment.

- **Request:** `GetPaymentStatusRequest`
- **Response:** `Transaction` details.

---

## Message Definitions

### CreatePaymentRequest
| Field | Type | Description |
|-------|------|-------------|
| `order_id` | `string` | Target Order UUID |
| `organization_id` | `string` | **Required** for loading merchant config |
| `payment_method` | `string` | Gateway Code (e.g., `bkash`, `sslcommerz`) |
| `amount_paisa` | `int64` | Amount in smallest currency unit |
| `return_url` | `string` | Base URL for success callback |

### Encryption & Security
> **Note:** Merchant credentials stored in the DB are currently stored as JSON blobs. Future enhancements will include encryption-at-rest for these sensitive fields.

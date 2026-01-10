# Order Service gRPC Documentation

**Package:** `order.v1`  
**Internal DNS:** `order:9084`  
**Proto File:** `server/api/proto/order/v1/order.proto`

## Overview
The Order Service manages the booking lifecycle using the **Saga Pattern** for distributed transactions. It orchestrates Payment, Inventory, and Fulfillment services to ensure eventual consistency.

## Key Behaviors

### Distributed Transaction (Saga)
- **Creation:** `CreateOrder` immediately persists a `PENDING` order and initiates an asynchronous `BookingSaga`.
- **Steps:**
  1. **Inventory Hold:** Locks seats for 15 minutes.
  2. **Payment Authorization:** Charges the user's payment method.
  3. **Order Confirmation:** Updates status to `CONFIRMED`.
  4. **Fulfillment:** Generates tickets.
- **Failures:** If any step fails, a **Compensating Transaction** rolls back previous steps (e.g., releasing seat holds, refuding payments).

### Business Rules
- **Expiry:** Pending orders expire automatically after **15 minutes** if payment is not completed.
- **Idempotency:** `CreateOrder` supports `IdempotencyKey` to safely retry requests without creating duplicate bookings.
- **Tax & Fees:** Currently fixed at **5% VAT** and **20 BDT Booking Fee** per passenger.
- **Cancellation:** Only `CONFIRMED` orders can be cancelled. Cancellation triggers a refund saga.

> [!WARNING]
> **Production Note:** Base seat pricing is currently using placeholder logic (Fallbacks to 800 BDT). Dynamic pricing integration with `pricing-service` is pending final wiring.

---

## RPC Methods

### `CreateOrder`
Initiates a new booking. Returns `202 Accepted` behavior (order created in `PENDING` state).

- **Request:** `CreateOrderRequest`
- **Response:** `domain.Order` (with `saga_id`)

### `CancelOrder`
Initiates a cancellation and refund process.

- **Pre-condition:** Order must be `CONFIRMED`.
- **Effect:** Triggers `CancellationSaga`, refunds payment to source, releases inventory.
- **Response:** `CancelOrderResponse` (includes `RefundInfo`)

### `GetOrderStatus`
Returns not just the order status, but the granular status of the underlying Saga (e.g., which step is currently executing).

---

## Message Definitions

### Order Status State Machine
| Status | Description | Transition To |
|--------|-------------|---------------|
| `PENDING` | Saga started, awaiting payment | `CONFIRMED`, `FAILED` |
| `CONFIRMED` | Payment captured, tickets ready | `CANCELLED`, `REFUNDED` |
| `FAILED` | Saga failed (payment declined, timeout) | Terminal |
| `CANCELLED` | User initiated cancellation | `REFUNDED` |
| `REFUNDED` | Money returned | Terminal |

### CreateOrderRequest
| Field | Type | Description |
|-------|------|-------------|
| `idempotency_key` | `string` | **Required** for safe retries (UUIDv4 recommended) |
| `hold_id` | `string` | Optional. Pre-reserved seat hold ID from Inventory service |
| `payment_method` | `string` | `card`, `bkash`, `nagad` |


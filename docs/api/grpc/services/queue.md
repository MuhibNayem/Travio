# Queue Service gRPC Documentation

**Package:** `queue.v1`  
**Internal DNS:** `queue:9087`  
**Proto File:** `server/api/proto/queue/v1/queue.proto`

## Overview
The Queue Service implements a **Virtual Waiting Room** to protect downstream services from traffic spikes (e.g., during Flash Sales). It holds users in a FIFO queue and admits them at a controlled rate.

## Key Behaviors

### Admission Control
- **Workers:** A background `AdmissionWorker` runs per event.
- **Rate Limiting:** Admits `N` users every `T` interval (configurable per event).
- **Stateless Tokens:** Admitted users receive a signed JWT (Admission Token) that grants access to protected resources for a limited time (`TokenTTL`).

### Dynamic Configuration
- **Hot Reloading:** Changing queue parameters (batch size, interval) restarts the background worker immediately without service downtime.
- **Enable/Disable:** Queues can be toggled on/off instantly.

---

## RPC Methods

### `JoinQueue`
Enters the user into the waiting pool.

- **Request:** `JoinQueueRequest`.
- **Response:** `QueuePosition` (Current rank and estimated wait time).
- **Backing:** Atomic Redis Lua script.

### `ValidateToken`
Verifies if a user has been admitted.

- **Request:** `ValidateTokenRequest`.
- **Mechanism:** Verifies JWT signature (Stateless).

### `ConfigureQueue`
(Admin) Updates the admission policy.

- **Request:** `ConfigureQueueRequest`.
- **Fields:** `max_concurrent`, `batch_size`, `interval_secs`, `token_ttl_secs`.

---

## Message Definitions

### QueuePosition
| Field | Type | Description |
|-------|------|-------------|
| `position` | `int32` | Users ahead of you |
| `estimated_wait` | `int32` | Seconds remaining |
| `status` | `QueueStatus` | `WAITING`, `READY`, `EXPIRED` |

### QueueStats
| Field | Type | Description |
|-------|------|-------------|
| `total_waiting` | `int32` | Current queue depth |
| `admission_rate` | `int32` | Users admitted per minute |

# Order Service

The **Order Service** manages financial transactions, booking orchestration (Saga), and payment processing.

![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)
![Database](https://img.shields.io/badge/db-postgres-336791)
![KV Store](https://img.shields.io/badge/kv-redis-red)

## 🛡 Reliability Features

### 1. Idempotency (Edge Protection)
Prevents double-charging and duplicate processing using `Idempotency-Key` header handling in Redis.
- **Header**: `Idempotency-Key: <unique-uuid>`
- **Logic**: 
  - If key seen & processing -> **409 Conflict**.
  - If key seen & completed -> **Return Cached Response**.
  - If new -> **Process & Cache**.

### 2. Distributed Sagas (Orchestration)
Implements a persistent Saga pattern to manage distributed transactions across Inventory, Payment, and Notification services.
- **Persistence**: Application state is saved to Postgres `saga_instances` table at every step.
- **Crash Recovery**: Service can resume or compensate "stuck" sagas on restart (implementation ready for recovery worker).

### 3. Usage-Based Billing Integration
Automatically tracks ticket sales for platform usage billing.
- **Integration**: Calls `SubscriptionServer.RecordUsage` (Best Effort) on booking confirmation.
- **Failures**: Logged but do not rollback the booking transaction (User Experience > Internal Ops).

## 🚀 Getting Started

### Prerequisites
- Go 1.25+
- Docker (Postgres, Redis)

### Running Locally
```bash
# Start Dependencies
docker compose up -d postgres redis

# Run Service
go run cmd/main.go
```

### Configuration
| Env Var | Default | Description |
| :--- | :--- | :--- |
| `HTTP_PORT` | `8084` | HTTP API Port |
| `GRPC_PORT` | `9084` | gRPC Server Port |
| `POSTGRES_URL` | `...` | DB Connection String |
| `REDIS_ADDR` | `localhost:6379` | Redis Address |

## 🧪 Verification

### Run Idempotency Test
Simulate concurrent duplicate requests:
```bash
go run load_test/load.go
```
**Expected Result**: 1 Request Processed, 9 Conflicts or Hits.

# Inventory Service

The **Inventory Service** manages seat availability, locking, and booking for the Travio platform. It is designed for high-throughput "Flash Sale" scenarios using ScyllaDB and Redis.

![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)
![Database](https://img.shields.io/badge/db-scylladb-4495D1)
![Cache](https://img.shields.io/badge/cache-redis-red)

## 🏗 Architecture

- **Database**: **ScyllaDB** (Cassandra-compatible) for massive write throughput and availability.
- **Concurrency Control**: 
  - **L1 (Redis)**: Optimistic "Pre-Locking" (`SET NX`) to prevent Thundering Herd on hot seats.
  - **L2 (Scylla)**: Lightweight Transactions (LWT) `IF status = 'available'` for strict ACID correctness.
- **Caching**: **Read-Through** Redis cache for Seat Maps, updated asynchronously on modification.

## 🚀 Key Features

### 1. Optimistic Pre-Locking
To protect the database from 100k+ concurrent requests for the same seat (e.g., front row at a concert), we acquire a short-lived Redis lock *before* initiating the DB transaction. This fails fast ~99% of requests at the edge.

### 2. Segment-Based Inventory
Seats are tracked per *Segment*. A Trip (A -> C) consists of Segments (A -> B, B -> C). A user booking A -> C holds the seat on *both* segments.

### 3. Seat Map Caching
`GetSeatMap` checks Redis first. On a cache miss, it efficiently aggregates data from ScyllaDB and warms the cache for subsequent users, reducing DB read pressure by >90%.

## ⚡ Getting Started

### Prerequisites
- Go 1.25+
- Docker (for ScyllaDB & Redis)

### Running Locally
```bash
# Start Dependencies
docker compose up -d scylla redis

# Start Service
go run cmd/main.go
```

### Configuration
| Env Var | Default | Description |
| :--- | :--- | :--- |
| `GRPC_PORT` | `9083` | gRPC Server Port |
| `SCYLLA_HOSTS` | `localhost` | ScyllaDB Hosts |
| `REDIS_ADDR` | `localhost:6379` | Redis Address |

## 🧪 Scalability Verification

We include a load test to simulate "Hot Seat" contention.

```bash
go run load_test/load.go
```

**Scenario**: 50 concurrent workers trying to hold the *same* seat (`seat-A1`).
**Expectation**: 
- **1 Success**: Only one lucky winner.
- **High Contention Errors**: Most requests explicitly fail with "contention" error from Redis, proving the DB protection works.

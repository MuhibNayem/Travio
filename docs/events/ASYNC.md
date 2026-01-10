# Event-Driven Communication (Kafka)

Travio uses Apache Kafka for asynchronous service communication, ensuring high availability and eventual consistency.

## 📡 Message Bus Overview

| Topic | Producer | Consumers | Payload Type |
|-------|----------|-----------|--------------|
| `order.created` | Order Service | Fraud, Notification, Inventory | Proto: `OrderCreated` |
| `payment.processed`| Payment Service | Order, Subscription | Proto: `PaymentProcessed`|
| `inventory.held` | Inventory Service | Order | Proto: `InventoryHeld` |
| `trip.updated` | Catalog Service | Search | Proto: `TripUpdated` |

## 🛠 Schema Registry
Common event schemas are located in `server/pkg/events/schemas`. All messages are serialized using **Protobuf** for performance.

## 🔄 Consumer Groups
- `search-indexer`: Re-indexes trips in OpenSearch upon catalog updates.
- `notification-worker`: Sends emails/SMS on order confirmation or payment failure.
- `reporting-aggregator`: Syncs transactional data to ClickHouse for OLAP.

## 🛡 Reliability
- **At-least-once delivery:** Handled via Kafka ack configuration.
- **Idempotency:** Consumers implement idempotency checks using `event_id` or `idempotency_key`.
- **Dead Letter Queues (DLQ):** Failed messages are routed to `.failed` topics for manual inspection.

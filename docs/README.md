# Travio Project Documentation 🚀

Welcome to the official documentation for **Travio**, a high-performance, FAANG-grade travel booking microservice ecosystem.

## 📌 Quick Links

- [**System Architecture**](./architecture/OVERVIEW.md) - High-level system design and technology stack.
- [**REST API (Gateway)**](./api/rest/openapi.yaml) - Public API documentation (OpenAPI 3.0).
- [**gRPC Services**](./api/grpc/README.md) - Internal service-to-service communication.
- [**Event-Driven Communication**](./events/ASYNC.md) - Kafka message schemas and producers/consumers.

## 🏗 Technology Stack

- **Backend:** Go (1.25.3), gRPC, Chi (REST Gateway)
- **Databases:** PostgreSQL, ClickHouse, Redis, OpenSearch, ScyllaDB
- **Messaging:** Kafka, Zookeeper
- **Observability:** OpenTelemetry (Tracing, Metrics, Logs)
- **Infrastructure:** Docker Compose, Kubernetes Ready

## 🛠 Documentation Generation
This documentation is maintained via automated tools and manual updates to ensure absolute accuracy with the production codebase.

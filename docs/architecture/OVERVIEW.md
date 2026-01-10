# System Architecture

## Overview
Travio is built using a **Cloud-Native Microservices Architecture**. All services communicate via **gRPC** for internal requests and export a public-facing **REST API** through a centralized Gateway.

## Service Map

```mermaid
graph TD
    User([User]) --> Gateway[API Gateway]
    
    subgraph "Core Services"
        Gateway --> Identity[Identity Service]
        Gateway --> Catalog[Catalog Service]
        Gateway --> Inventory[Inventory Service]
        Gateway --> Order[Order Service]
        Gateway --> Payment[Payment Service]
    end
    
    subgraph "Internal Intelligence & Search"
        Gateway --> Search[Search Service]
        Gateway --> Fraud[Fraud Service]
        Gateway --> Pricing[Pricing Service]
    end
    
    subgraph "Post-Booking & Support"
        Gateway --> Fulfillment[Fulfillment Service]
        Gateway --> Subscription[Subscription Service]
        Gateway --> Reporting[Reporting Service]
        Gateway --> Notification[Notification Service]
    end
    
    subgraph "Infrastructure"
        IdentityData[(PostgreSQL)] --- Identity
        ClickHouse[(ClickHouse)] --- Reporting
        OpenSearch[(OpenSearch)] --- Search
        OpenSearch --- Fraud
        Redis[(Redis)] --- Gateway
        Kafka{Kafka} --- Search
        Kafka --- Reporting
        Kafka --- Notification
    end
```

## Data Consistency
- **Transactional Data:** PostgreSQL (OLTP)
- **Analytical Data:** ClickHouse (OLAP)
- **Real-time Search:** OpenSearch
- **Caching:** Redis 
- **Event Bus:** Kafka (Pub/Sub)

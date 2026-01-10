# gRPC Service Documentation

This directory contains technical references for all internal Travio microservices.

## 📖 Service Documentation

| Service | Package | Internal DNS | Description |
|---------|---------|--------------|-------------|
| [**Identity**](./services/identity.md) | `identity.v1` | `identity:9081` | IAM, Authentication, Referrals |
| [**Order**](./services/order.md) | `order.v1` | `order:9084` | Booking Saga, Lifecycle Management |
| [**Catalog**](./services/catalog.md) | `catalog.v1` | `catalog:9082` | Trip Scheduling, Stations |
| [**Inventory**](./services/inventory.md) | `inventory.v1` | `inventory:9083` | Seat Availability, Holds, Maps |
| [**Payment**](./services/payment.md) | `payment.v1` | `payment:9085` | Gateways, Transactions |
| [**Fulfillment**](./services/fulfillment.md) | `fulfillment.v1` | `fulfillment:9086` | Ticket Generation, Validation |
| [**Search**](./services/search.md) | `search.v1` | `search:9088` | Trip Discovery, OpenSearch |
| [**Pricing**](./services/pricing.md) | `pricing.v1` | `pricing:50058` | Dynamic Pricing Engine |
| [**Fraud**](./services/fraud.md) | `fraud.v1` | `fraud:50090` | Risk Analysis |
| [**Reporting**](./services/reporting.md) | `reporting.v1` | `reporting:50091` | OLAP Analytics |
| [**Subscription**](./services/subscription.md) | `subscription.v1` | `subscription:50060` | SaaS Billing |
| [**Queue**](./services/queue.md) | `queue.v1` | `queue:9087` | Virtual Waiting Room |
| [**Operator**](./services/operator.md) | `operator.v1` | `operator:50059` | Vendor Management |
| [**Notification**](./services/notification.md) | `notification.v1` | `notification:9090` | Email, SMS Dispatcher |

> [!NOTE]
> All services communicate via **gRPC** over mTLS.
> - **Protocol:** HTTP/2 (gRPC)
> - **Serialization:** Protobuf 3
> - **Security:** mTLS required for inter-service communication

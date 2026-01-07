# Notification Service

Handles email and SMS notifications triggered by Kafka events.

## Scalability Features

-   **Provider Rate Limiting**:
    -   Email: 10 req/s (SES safe limit)
    -   SMS: 5 req/s (Twilio safe limit)
    -   Uses token bucket algorithm (`golang.org/x/time/rate`)
    -   Implementation: `internal/provider/rate_limiter.go`

-   **Template Engine**:
    -   Templates stored in `internal/service/templates/`
    -   Uses Go's `text/template` with `embed.FS`
    -   No redeployment needed for template changes (when using external storage)

## Setup

```bash
# 1. Start Kafka
docker compose up -d kafka

# 2. Run Service
KAFKA_BROKERS=localhost:9092 go run cmd/main.go
```

## Configuration
-   `KAFKA_BROKERS`: Kafka broker addresses

## Verification

### Load Test
Tests rate limiting by sending 100 emails rapidly.

```bash
go run load_test/load.go
```

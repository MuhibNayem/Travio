# Fulfillment Service

Generates Tickets (PDF/QR) and manages fulfillment lifecycle.

## Features

-   **Ticket Generation**: Generates PDF tickets with embedded QR codes.
-   **Async Storage**: Offloads PDF storage to **MinIO** (S3 Compatible).
-   **Secure Access**: Returns **Presigned URLs** (15m expiry) instead of public links or raw bytes.
-   **QR Validation**: Validates tickets via cryptographic signature.

## Setup

```bash
# 1. Start dependencies (Postgres, MinIO)
docker compose up -d postgres minio

# 2. Run Service
go run cmd/main.go
```

## MinIO Configuration
Ensure `config/config.go` or ENV vars match your MinIO setup.
-   Default Bucket: `tickets` (Private)
-   Default Endpoint: `localhost:9000`

## Verification

### Load Test
Simulates concurrent ticket generation and PDF storage.

```bash
go run load_test/load.go
```

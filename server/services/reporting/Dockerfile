# Multi-stage Dockerfile for Reporting Service
FROM golang:1.25.3-alpine AS builder

# Install git and ca-certificates for HTTPS
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go workspace files

# Copy module files
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY api/go.mod api/go.sum ./api/
COPY services/reporting/go.mod services/reporting/go.sum ./services/reporting/

# Init new go.work since we only copied specific services
RUN go work init ./pkg ./api ./services/reporting

# Download dependencies
RUN cd services/reporting && go mod download

# Copy source code
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/reporting/ ./services/reporting/

# Build the binary
WORKDIR /app/services/reporting
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /reporting-service ./cmd

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /reporting-service .

EXPOSE 50091

CMD ["./reporting-service"]

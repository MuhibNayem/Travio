# Multi-stage Dockerfile for Fraud Service
FROM golang:1.25.3-alpine AS builder

# Install git and ca-certificates for HTTPS
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go workspace files

# Copy module files
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY api/go.mod api/go.sum ./api/
COPY services/fraud/go.mod services/fraud/go.sum ./services/fraud/

# Init new go.work since we only copied specific services
RUN go work init ./pkg ./api ./services/fraud

# Download dependencies
RUN cd services/fraud && go mod download

# Copy source code
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/fraud/ ./services/fraud/

# Build the binary
WORKDIR /app/services/fraud
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /fraud-service ./cmd

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /fraud-service .

EXPOSE 50090

CMD ["./fraud-service"]

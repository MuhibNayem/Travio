FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy shared modules
COPY api ./api
COPY pkg ./pkg

# Copy service module
COPY services/pricing ./services/pricing

# Build the binary
WORKDIR /app/services/pricing
RUN go mod download
RUN go build -o /app/pricing-service ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/pricing-service .

EXPOSE 8080 50051

CMD ["./pricing-service"]

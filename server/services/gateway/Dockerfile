FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.work and dependencies first for better caching
COPY go.work go.work.sum ./
COPY api ./api
COPY pkg ./pkg
COPY services/gateway ./services/gateway

WORKDIR /app/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add wget ca-certificates
WORKDIR /root/
COPY --from=builder /gateway-service .

EXPOSE 8080
CMD ["./gateway-service"]

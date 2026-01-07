# TRAVIO - Inventory Service Dockerfile
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./services/inventory/cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /app/service /app/service
USER appuser
EXPOSE 8083 9083
ENTRYPOINT ["/app/service"]

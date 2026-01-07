FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.work and api proto first for better caching
COPY go.work go.work.sum ./
COPY api ./api
COPY pkg ./pkg
COPY services/identity ./services/identity

WORKDIR /app/services/identity
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /identity-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add wget
WORKDIR /root/
COPY --from=builder /identity-service .

EXPOSE 8081 9081
CMD ["./identity-service"]

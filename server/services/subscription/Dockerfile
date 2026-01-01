FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY pkg ./pkg
COPY services/subscription ./services/subscription

WORKDIR /app/services/subscription
RUN go mod download
RUN go build -o /subscription-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /subscription-service .

EXPOSE 8093 9093
CMD ["./subscription-service"]

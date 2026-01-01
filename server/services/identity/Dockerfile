FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared pkg (Context needs to be root of monorepo)
COPY pkg ./pkg
COPY services/identity ./services/identity

WORKDIR /app/services/identity
RUN go mod download
RUN go build -o /identity-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /identity-service .

EXPOSE 8081
CMD ["./identity-service"]

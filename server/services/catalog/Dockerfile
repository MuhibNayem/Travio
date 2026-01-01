FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY pkg ./pkg
COPY services/catalog ./services/catalog

WORKDIR /app/services/catalog
RUN go mod download
RUN go build -o /catalog-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /catalog-service .

EXPOSE 8082
CMD ["./catalog-service"]

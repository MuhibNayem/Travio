FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.work and api proto first for better caching
COPY go.work go.work.sum ./
COPY api ./api
COPY pkg ./pkg
COPY services/catalog ./services/catalog

WORKDIR /app/services/catalog
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /catalog-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add wget
WORKDIR /root/
COPY --from=builder /catalog-service .

EXPOSE 8082 9082
CMD ["./catalog-service"]

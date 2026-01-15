FROM golang:1.25.3-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY api ./api
COPY services/operator ./services/operator

WORKDIR /app/services/operator
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go mod download
RUN go build -o operator-service cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/services/operator/operator-service .

CMD ["./operator-service"]

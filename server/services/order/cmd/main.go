package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/order/config"
	"github.com/MuhibNayem/Travio/server/services/order/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/order/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/order/internal/messaging"
	"github.com/MuhibNayem/Travio/server/services/order/internal/middleware"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
	"github.com/MuhibNayem/Travio/server/services/order/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger.Init("order-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/travio_order?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
	}

	// GORM for Sagas
	gormDB, err := gorm.Open(postgres.Open("postgres://postgres:postgres@localhost:5432/travio_order?sslmode=disable"), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to GORM DB", "error", err)
	}

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	// Kafka DLQ
	dlq, err := messaging.NewKafkaDLQProducer([]string{"localhost:9092"}, "order-saga-dlq")
	if err != nil {
		logger.Error("Failed to initialize DLQ producer", "error", err)
		// We proceed without DLQ (nil), but log error
	} else {
		logger.Info("Initialized DLQ producer")
	}

	// Service clients
	inventoryClient, err := clients.NewInventoryClient(cfg.Services.InventoryAddr)
	if err != nil {
		logger.Error("Failed to connect to inventory service", "error", err)
	}

	paymentClient, err := clients.NewPaymentClient(cfg.Services.PaymentAddr)
	if err != nil {
		logger.Error("Failed to connect to payment service", "error", err)
	}

	nidClient, err := clients.NewNIDClient(cfg.Services.NIDAddr)
	if err != nil {
		logger.Error("Failed to connect to NID service", "error", err)
	}

	notificationClient := clients.NewNotificationClient()

	subscriptionClient, err := clients.NewSubscriptionClient(cfg.Services.SubscriptionAddr)
	if err != nil {
		logger.Error("Failed to connect to subscription service", "error", err)
	}

	// Saga dependencies
	sagaDeps := &saga.BookingDependencies{
		NIDService:          nidClient,
		InventoryService:    inventoryClient,
		PaymentService:      paymentClient,
		SubscriptionService: subscriptionClient,
		NotificationSvc:     notificationClient,
	}

	// Repository and service
	orderRepo := repository.NewOrderRepository(db)
	var dlqProducer messaging.DLQProducer // interface
	if dlq != nil {
		dlqProducer = dlq
	}
	orderService := service.NewOrderService(db, gormDB, dlqProducer, orderRepo, sagaDeps)
	grpcHandler := handler.NewGrpcHandler(orderService)

	// Idempotency Middleware
	idempotency := middleware.NewIdempotencyMiddleware(redisClient)

	// HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterOrderServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Order service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)

	// Wrap mux with Idempotency Middleware
	handler := idempotency.Middleware(mux)

	srv.Start(handler)
}

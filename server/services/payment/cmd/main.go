package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/payment/config"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/model"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/service"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger.Init("payment-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to GORM DB", "error", err)
		panic(err)
	} else {
		logger.Info("Connected to DB, running migrations...")
		_ = db.AutoMigrate(&model.Transaction{}, &model.PaymentConfig{})
		
		// Migrate refund table (TASK-022)
		refundRepo := repository.NewRefundRepository(db)
		_ = refundRepo.AutoMigrate()
	}

	repo := repository.NewTransactionRepository(db)
	configRepo := repository.NewPaymentConfigRepository(db)
	refundRepo := repository.NewRefundRepository(db)

	// Initialize Redis client for IPN idempotency (TASK-020)
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis not available, IPN idempotency disabled", "error", err)
		redisClient = nil
	}

	// Initialize Kafka producer for payment events (TASK-019)
	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	var kafkaProducer *kafka.Producer
	kafkaProd, err := kafka.NewProducer(brokers)
	if err != nil {
		logger.Warn("Kafka not available, payment events will not be published", "error", err)
	} else {
		kafkaProducer = kafkaProd
		logger.Info("Kafka producer initialized")
	}

	// Initialize payment gateways registry with Factories
	registry := gateway.NewRegistry()

	// Register Factories
	registry.Register("sslcommerz", &gateway.SSLCommerzFactory{})
	registry.Register("bkash", &gateway.BKashFactory{})
	registry.Register("nagad", &gateway.NagadFactory{})

	// Start Reconciliation Worker
	reconciler := worker.NewReconciler(repo, configRepo, registry, 5*time.Minute)
	go reconciler.Start(context.Background())

	// Service and handler
	paymentService := service.NewPaymentService(registry, repo, configRepo)
	grpcHandler := handler.NewGrpcHandler(paymentService, registry, repo, configRepo)

	// IPN Webhook Handler (TASK-014 to TASK-021)
	ipnHandler := handler.NewIPNHandler(
		paymentService,
		registry,
		repo,
		refundRepo,
		configRepo,
		redisClient,
		kafkaProducer,
	)

	// HTTP mux with webhook routes (TASK-014)
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Register IPN webhook routes
	r.Route("/v1/payments/ipn", func(r chi.Router) {
		ipnHandler.RegisterRoutes(r)
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterPaymentServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Payment service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(r)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

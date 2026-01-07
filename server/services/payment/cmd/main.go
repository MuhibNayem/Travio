package main

import (
	"context"
	"net/http"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/payment/config"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/model"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/service"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/worker"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger.Init("payment-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	// Using default local credentials consistent with other services
	dsn := "postgres://postgres:postgres@localhost:5432/travio_payment?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to GORM DB", "error", err)
	} else {
		logger.Info("Connected to DB, running migrations...")
		_ = db.AutoMigrate(&model.Transaction{})
		// Ensure payment_configs table? Actually we rely on init-db.sql, but GORM AutoMigrate is safer
		// Define PaymentConfig in model? Repository defines it in internal/repository/config.go but better if it was in internal/model.
		// For now, let's skip auto-migrating PaymentConfig unless we move it to model package.
		// It's defined in repository package, so we can't reference it easily here without import cycle or weirdness.
		// Assuming init-db.sql handles it.
	}

	repo := repository.NewTransactionRepository(db)
	configRepo := repository.NewPaymentConfigRepository(db)

	// Initialize payment gateways registry with Factories
	registry := gateway.NewRegistry()

	// Register Factories
	registry.Register("sslcommerz", &gateway.SSLCommerzFactory{})
	registry.Register("bkash", &gateway.BKashFactory{})
	registry.Register("nagad", &gateway.NagadFactory{})

	// Start Reconciliation Worker
	reconciler := worker.NewReconciler(repo, configRepo, registry, 5*time.Minute)
	go reconciler.Start(context.Background())
	// logger.Warn("Reconciliation worker temporarily disabled during dynamic gateway refactor")

	// Service and handler
	// Service and handler
	paymentService := service.NewPaymentService(registry, repo, configRepo)
	grpcHandler := handler.NewGrpcHandler(paymentService, registry, repo, configRepo)

	// HTTP mux for health (IPN webhooks might need updates too, skipping dynamic IPN for now)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterPaymentServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Payment service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

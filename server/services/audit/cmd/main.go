package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/audit/v1"
	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/audit/config"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger.Init("audit-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	dsn := cfg.Database.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("Failed to connect to DB", "error", err)
	}
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping DB", "error", err)
	}

	// Create audit table if not exists
	auditRepo := repository.NewAuditRepository(db)
	if err := auditRepo.CreateTable(context.Background()); err != nil {
		logger.Warn("Audit table creation warning", "error", err)
	}

	// Kafka consumer for audit events
	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	var kafkaConsumer *kafka.Consumer
	if getEnv("KAFKA_ENABLED", "true") == "true" {
		kafkaCons, err := kafka.NewConsumer(brokers, "audit-service", []string{"audit-events"})
		if err != nil {
			logger.Warn("Kafka not available, audit event consumption disabled", "error", err)
		} else {
			kafkaConsumer = kafkaCons
			logger.Info("Kafka consumer initialized")
		}
	}

	// Service and handlers
	auditSvc := service.NewAuditService(auditRepo)
	grpcHandler := handler.NewGrpcHandler(auditSvc)
	httpHandler := handler.NewHTTPHandler(auditSvc)

	// Start Kafka consumer if available
	if kafkaConsumer != nil {
		go func() {
			logger.Info("Starting Kafka consumer for audit events...")
			if err := kafkaConsumer.Start(); err != nil {
				logger.Error("Kafka consumer error", "error", err)
			}
		}()
		defer kafkaConsumer.Stop()
	}

	// Start data retention cleanup worker
	retentionSvc := service.NewRetentionService(auditRepo, 90*24*time.Hour) // 90 days
	go retentionSvc.StartCleanup(context.Background(), 24*time.Hour)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuditServiceServer(grpcServer, grpcHandler)

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("audit.v1.AuditService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", "port", cfg.GRPCPort, "error", err)
	}

	// Start HTTP server for health and API endpoints
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Register HTTP audit routes
		httpHandler.RegisterRoutes(mux)

		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTPPort), mux); err != nil {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("Audit Service listening", "grpc_port", cfg.GRPCPort, "http_port", cfg.HTTPPort)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Shutting down Audit Service")
		grpcServer.GracefulStop()
		retentionSvc.Stop()
	}()

	// Start gRPC server
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", "error", err)
	}

	logger.Info("Audit Service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

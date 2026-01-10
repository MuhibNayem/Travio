package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/events/v1"
	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/events/config"
	"github.com/MuhibNayem/Travio/server/services/events/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/events/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/events/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("events-service")
	cfg := config.Load()

	// Database Setup
	logger.Info("Connecting to Postgres...", "dsn", cfg.Database.DSN())
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
		panic(err)
	}
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping DB", "error", err)
		panic(err)
	}

	// Kafka Setup
	logger.Info("Initializing Kafka Producer...", "brokers", cfg.Kafka.Brokers)
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		logger.Error("Failed to create Kafka producer", "error", err)
		panic(err)
	}

	// Dependency Injection
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo, producer)
	grpchandler := handler.NewGRPCHandler(svc)

	// Setup Server
	srv := server.New(cfg.Server)

	// Register gRPC
	pb.RegisterEventServiceServer(srv.GRPC(), grpchandler)

	// HTTP Mux (Health check)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Info("Starting Events Service...")
	srv.Start(mux)
}

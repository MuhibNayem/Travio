package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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

const (
	maxDBRetries      = 10
	initialRetryDelay = 2 * time.Second
)

func main() {
	logger.Init("events-service")
	cfg := config.Load()

	// Database Setup with retry to handle slow DNS/network on startup
	db, err := connectWithRetry(cfg.Database.DSN(), maxDBRetries, initialRetryDelay)
	if err != nil {
		logger.Error("Failed to establish DB connection after retries", "error", err)
		panic(err)
	}
	logger.Info("Connected to Postgres")

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

func connectWithRetry(dsn string, maxRetries int, delay time.Duration) (*sql.DB, error) {
	var attempt int
	for {
		logger.Info("Connecting to Postgres...", "dsn", dsn, "attempt", attempt+1)
		db, err := sql.Open("pgx", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				return db, nil
			} else {
				err = pingErr
			}
		}

		if db != nil {
			_ = db.Close()
		}

		attempt++
		if maxRetries > 0 && attempt >= maxRetries {
			return nil, fmt.Errorf("postgres connection failed after %d attempts: %w", attempt, err)
		}

		logger.Warn("Postgres unreachable, retrying...", "error", err, "next_delay", delay)
		time.Sleep(delay)
		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}
}

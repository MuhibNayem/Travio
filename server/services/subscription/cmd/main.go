package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger first
	logger.Init("subscription-service")

	// 1. Load Config
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found")
	}

	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "travio_subscription"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50060" // Default port for Subscription
	}

	// 2. Connect to DB
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to connect to DB", "error", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping DB", "error", err)
	}
	logger.Info("Connected to PostgreSQL", "db", dbName)

	// 3. Init Layers
	repo := repository.NewPostgresRepository(db)
	svc := service.NewSubscriptionService(repo)

	// 4. Start HTTP health server for Docker healthchecks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		httpPort := os.Getenv("HTTP_PORT")
		if httpPort == "" {
			httpPort = "8060"
		}
		logger.Info("Starting HTTP health endpoint", "port", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Warn("HTTP health server failed", "error", err)
		}
	}()

	// 5. Start Server
	go func() {
		if err := handler.StartGRPCServer(grpcPort, svc); err != nil {
			logger.Fatal("Failed to start gRPC server", "error", err)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Subscription Service...")
}

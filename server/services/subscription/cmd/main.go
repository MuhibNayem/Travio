package main

import (
	"database/sql"
	"fmt"
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
	// 1. Load Config
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := "travio_subscription"
	grpcPort := os.Getenv("SUBSCRIPTION_SERVICE_PORT")
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

	// 4. Start Server
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

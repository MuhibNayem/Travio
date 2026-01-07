package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/pricing/config"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/pricing/internal/service"
)

func main() {
	logger.Init("pricing-service")
	cfg := config.Load()

	// Connect to PostgreSQL
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPostgresRepository(db)
	if err := repo.InitSchema(context.Background()); err != nil {
		logger.Error("Failed to initialize schema", "error", err)
	}
	if err := repo.SeedDefaultRules(context.Background()); err != nil {
		logger.Error("Failed to seed default rules", "error", err)
	}

	// Initialize service
	svc, err := service.NewPricingService(repo)
	if err != nil {
		logger.Error("Failed to create pricing service", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	httpHandler := handler.NewHTTPHandler(svc)
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)

	httpPort := fmt.Sprintf(":%d", cfg.GRPCPort)
	server := &http.Server{
		Addr:    httpPort,
		Handler: mux,
	}

	go func() {
		logger.Info("Pricing HTTP server starting", "port", cfg.GRPCPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	logger.Info("Pricing service started", "port", cfg.GRPCPort)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down pricing service")
	server.Shutdown(context.Background())
}

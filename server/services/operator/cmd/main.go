package main

import (
	"context"
	"net/http"
	"os"

	pb "github.com/MuhibNayem/Travio/server/api/proto/operator/v1"
	"github.com/MuhibNayem/Travio/server/pkg/database/postgres"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/operator/config"
	"github.com/MuhibNayem/Travio/server/services/operator/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/operator/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/operator/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("operator-service")
	cfg := config.Load()

	logger.Info("Connecting to database...", "dbname", cfg.Database.DBName)
	db, err := postgres.Connect(context.Background(), cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Dependency Injection
	repo := repository.NewPostgresRepository(db)
	svc := service.NewVendorService(repo)
	grpcHandler := handler.NewGrpcHandler(svc)

	// Server Setup
	srv := server.New(cfg.Server)

	// Register gRPC
	pb.RegisterVendorServiceServer(srv.GRPC(), grpcHandler)

	// HTTP Mux (Health check)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Info("Operator Service Starting...", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/crm/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/crm/config"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/crm/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("crm-service")
	cfg := config.Load()

	// Database Setup
	logger.Info("Connecting to database...", "dsn", cfg.Database.DSN())
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
		panic(err)
	}
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping DB", "error", err)
		panic(err)
	}

	// Dependency Injection
	repo := repository.NewCRMRepository(db)
	svc := service.NewCRMService(repo)
	grpchandler := handler.NewGRPCHandler(svc)

	// Setup Server
	srv := server.New(cfg.Server)

	// Register gRPC
	pb.RegisterCRMServiceServer(srv.GRPC(), grpchandler)

	// HTTP Mux (Health check)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Info("Starting CRM Service...")
	srv.Start(mux)
}

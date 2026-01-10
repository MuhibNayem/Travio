package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/fleet/config"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("fleet-service")
	cfg := config.Load()

	// Postgres Setup (Assets)
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

	// Scylla Setup (Locations)
	logger.Info("Connecting to ScyllaDB...", "host", cfg.ScyllaHost)
	scyllaRepo, err := repository.NewLocationRepository(cfg.ScyllaHost)
	if err != nil {
		logger.Error("Failed to connect to Scylla", "error", err)
		// panic(err) // Optional: fail soft if Scylla is down? usage is critical so maybe panic.
		// For now, let's log error but proceed to let assets work even if tracking is down contextually.
		// Actually, user wants production scale. It should probably fail or retry.
		// Let's panic to ensure visibility during startup.
		panic(err)
	}
	defer scyllaRepo.Close()

	// Dependency Injection
	assetRepo := repository.NewAssetRepository(db)
	svc := service.NewFleetService(assetRepo, scyllaRepo)
	grpchandler := handler.NewGRPCHandler(svc)

	// Setup Server
	srv := server.New(cfg.Server)

	// Register gRPC
	pb.RegisterFleetServiceServer(srv.GRPC(), grpchandler)

	// HTTP Mux (Health check)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Info("Starting Fleet Service...")
	srv.Start(mux)
}

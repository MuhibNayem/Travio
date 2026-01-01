package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/catalog/config"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("catalog-service")
	cfg := config.Load()

	// Database Setup
	logger.Info("Connecting to Postgres...")
	db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/travio_catalog?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
	}

	// Dependency Injection
	stationRepo := repository.NewStationRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	tripRepo := repository.NewTripRepository(db)
	catalogService := service.NewCatalogService(stationRepo, routeRepo, tripRepo)
	grpcHandler := handler.NewGrpcHandler(catalogService)

	// HTTP Mux (for health checks and REST fallback)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start gRPC and HTTP servers
	srv := server.New(cfg.Server)
	pb.RegisterCatalogServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Catalog service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

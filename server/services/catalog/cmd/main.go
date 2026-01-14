package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/catalog/config"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/events"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Init("catalog-service")
	cfg := config.Load()

	// Database Setup
	logger.Info("Connecting to Postgres...")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
	}

	// Redis Setup
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Dependency Injection (With Caching Decorators)
	stationRepo := repository.NewStationRepository(db)
	cachedStationRepo := repository.NewCachedStationRepository(stationRepo, rdb)

	routeRepo := repository.NewRouteRepository(db)
	cachedRouteRepo := repository.NewCachedRouteRepository(routeRepo, rdb)

	publisher := events.NewPublisher(db)
	tripRepo := repository.NewTripRepository(db, publisher)
	cachedTripRepo := repository.NewCachedTripRepository(tripRepo, rdb)
	scheduleRepo := repository.NewScheduleRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Entitlement Checker Setup
	subFetcher, err := entitlement.NewSubscriptionFetcher(cfg.SubscriptionURL)
	if err != nil {
		logger.Error("Failed to connect to subscription service", "error", err)
		// We might want to fail hard here in production, but for now we continue (will fail open/closed based on config)
	} else {
		defer subFetcher.Close()
	}

	entConfig := entitlement.DefaultConfig()
	entConfig.Enabled = true
	entConfig.FailOpen = true // Allow traffic if subscription service is down (availability over consistency)
	entConfig.CacheTTL = 5 * time.Minute

	entChecker := entitlement.NewCachedChecker(rdb, subFetcher, entConfig)
	entChecker.StartInvalidationListener(context.Background())

	entChecker.StartInvalidationListener(context.Background())

	// Clients Setup
	fleetClient, err := clients.NewFleetClient(cfg.FleetURL)
	if err != nil {
		logger.Error("Failed to create fleet client", "error", err)
	}

	inventoryClient, err := clients.NewInventoryClient(cfg.InventoryURL)
	if err != nil {
		logger.Error("Failed to create inventory client", "error", err)
	}

	catalogService := service.NewCatalogService(cachedStationRepo, cachedRouteRepo, cachedTripRepo, scheduleRepo, entChecker, fleetClient, inventoryClient, auditRepo)
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

package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	pb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"github.com/MuhibNayem/Travio/server/pkg/database/scylladb"
	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/inventory/config"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/consumer"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/service"
	"google.golang.org/grpc"
)

func main() {
	logger.Init("inventory-service")
	cfg := config.Load()

	// ScyllaDB Connection
	logger.Info("Connecting to ScyllaDB...")
	scyllaCfg := scylladb.Config{
		Hosts:          cfg.ScyllaDB.Hosts,
		Keyspace:       cfg.ScyllaDB.Keyspace,
		Consistency:    cfg.ScyllaDB.Consistency, // Use config value (defaults to ONE for dev)
		Timeout:        cfg.ScyllaDB.Timeout,
		ConnectTimeout: 10 * time.Second,
	}

	scyllaSession, err := scylladb.NewSession(scyllaCfg)
	if err != nil {
		logger.Fatal("Failed to connect to ScyllaDB", "error", err)
	}
	defer func() {
		if scyllaSession != nil {
			scyllaSession.Close()
		}
	}()

	// Redis Connection
	logger.Info("Connecting to Redis...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Entitlement Checker Setup
	var entitlementInterceptor grpc.UnaryServerInterceptor
	subscriptionAddr := os.Getenv("SUBSCRIPTION_URL")
	if subscriptionAddr == "" {
		subscriptionAddr = "localhost:50060"
	}
	fetcher, err := entitlement.NewSubscriptionFetcher(subscriptionAddr)
	if err != nil {
		logger.Warn("Failed to connect to subscription service for entitlement", "error", err)
	} else {
		checker := entitlement.NewCachedChecker(redisClient, fetcher, entitlement.DefaultConfig())
		entitlementInterceptor = entitlement.UnaryServerInterceptor(entitlement.InterceptorConfig{
			Checker:                   checker,
			SkipMethods:               []string{"GetTrip", "ListTrips", "GetSeatMap", "Health"}, // Read operations skip
			RequireActiveSubscription: true,
			QuotaKey:                  entitlement.FeatureMaxTripsPerMonth,
		})
		logger.Info("Entitlement enforcement enabled for Inventory Service")
	}

	// Dependency Injection
	scyllaRepo := repository.NewScyllaRepository(scyllaSession)

	// Initialize Schema (Migrations)
	if err := scyllaRepo.InitSchema(context.Background(), scyllaCfg.Keyspace); err != nil {
		logger.Error("Failed to initialize ScyllaDB schema", "error", err)
		// Depending on policy, we might want to panic here or continue
		// For now, log error but continue (as it might be intermittent connection or existing schema)
	}

	// Clients
	fleetClient, err := clients.NewFleetClient(cfg.FleetURL)
	if err != nil {
		logger.Error("Failed to create fleet client", "error", err)
	}

	redisRepo := repository.NewRedisRepository(redisClient)
	holdRepo := repository.NewHoldRepository(redisClient)
	inventoryService := service.NewInventoryService(scyllaRepo, holdRepo, redisRepo)
	grpcHandler := handler.NewGrpcHandler(inventoryService)

	// Event Consumer
	// Group ID usually "inventory-service"
	consumer, err := consumer.New(cfg.KafkaBrokers, "inventory-service", inventoryService, fleetClient)
	if err != nil {
		logger.Error("Failed to create Kafka consumer", "error", err)
		// We might want to exit or continue. Standard is to fail if we depend on events.
	} else {
		if err := consumer.Start(); err != nil {
			logger.Error("Failed to start Kafka consumer", "error", err)
		} else {
			defer consumer.Stop()
		}
	}

	// HTTP Mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start servers with entitlement interceptor
	var serverOpts []grpc.ServerOption
	if entitlementInterceptor != nil {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(entitlementInterceptor))
	}
	srv := server.NewWithOptions(cfg.Server, serverOpts...)
	pb.RegisterInventoryServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Inventory service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

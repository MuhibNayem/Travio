package main

import (
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"

	pb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/inventory/config"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/service"
)

func main() {
	logger.Init("inventory-service")
	cfg := config.Load()

	// ScyllaDB Connection
	logger.Info("Connecting to ScyllaDB...")
	cluster := gocql.NewCluster(cfg.ScyllaDB.Hosts...)
	cluster.Keyspace = cfg.ScyllaDB.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = cfg.ScyllaDB.Timeout
	cluster.ConnectTimeout = 10 * time.Second

	scyllaSession, err := cluster.CreateSession()
	if err != nil {
		logger.Error("Failed to connect to ScyllaDB", "error", err)
		// Don't crash in scaffold mode - will fail at runtime
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

	// Dependency Injection
	scyllaRepo := repository.NewScyllaRepository(scyllaSession)
	holdRepo := repository.NewHoldRepository(redisClient)
	inventoryService := service.NewInventoryService(scyllaRepo, holdRepo)
	grpcHandler := handler.NewGrpcHandler(inventoryService)

	// HTTP Mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start servers
	srv := server.New(cfg.Server)
	pb.RegisterInventoryServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Inventory service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

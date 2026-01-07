package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/queue/config"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/service"
	"google.golang.org/grpc"

	pb "github.com/MuhibNayem/Travio/server/api/proto/queue/v1"
)

func main() {
	logger.Init("queue-service")
	cfg := config.Load()

	// Initialize repository
	repo, err := repository.NewQueueRepository(cfg.RedisAddr)
	if err != nil {
		logger.Error("Failed to initialize queue repository", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize service
	// In production, TOKEN_SECRET should come from secure config/vault
	tokenSecret := "travio-super-secret-key-change-in-prod"
	queueService := service.NewQueueService(repo, tokenSecret)

	// gRPC server (commented out until proto is generated)
	// grpcHandler := handler.NewGrpcHandler(queueService)

	// HTTP server
	httpHandler := handler.NewHTTPHandler(queueService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", httpHandler.Health)
	mux.HandleFunc("/v1/queue/join", httpHandler.JoinQueue)
	mux.HandleFunc("/v1/queue/position", httpHandler.GetPosition)
	mux.HandleFunc("/v1/queue/validate", httpHandler.ValidateToken)
	mux.HandleFunc("/v1/queue/stats", httpHandler.GetStats)

	// Start servers
	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTPPort)
		logger.Info("Queue HTTP server starting", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Start gRPC server
	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewGrpcHandler(queueService)
	pb.RegisterQueueServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		return
	}

	logger.Info("Queue gRPC server starting", "port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server error", "error", err)
	}
}

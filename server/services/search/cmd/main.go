package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/MuhibNayem/Travio/server/api/proto/search/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/search/config"
	"github.com/MuhibNayem/Travio/server/services/search/internal/consumer"
	"github.com/MuhibNayem/Travio/server/services/search/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/search/internal/indexer"
	"github.com/MuhibNayem/Travio/server/services/search/internal/searcher"
	"github.com/opensearch-project/opensearch-go/v2"
	"google.golang.org/grpc"
)

func main() {
	logger.Init("search-service")
	cfg := config.Load()

	// Initialize OpenSearch client
	osClient, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{cfg.OpenSearchURL},
	})
	if err != nil {
		logger.Error("Failed to create OpenSearch client", "error", err)
		os.Exit(1)
	}

	// Initialize Indexer
	idx := indexer.New(osClient)
	if err := idx.InitIndices(context.Background()); err != nil {
		logger.Error("Failed to initialize indices", "error", err)
	}

	// Initialize Kafka Consumer
	consumer, err := consumer.New(cfg.KafkaBrokers, cfg.GroupID, idx)
	if err != nil {
		logger.Error("Failed to create Kafka consumer", "error", err)
		os.Exit(1)
	}

	if err := consumer.Start(); err != nil {
		logger.Error("Failed to start Kafka consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Stop()

	// Initialize Searcher
	searcher := searcher.New(osClient)

	// Start gRPC server
	grpcHandler := handler.NewGrpcHandler(searcher)
	grpcServer := grpc.NewServer()
	pb.RegisterSearchServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("Search gRPC server starting", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	logger.Info("Search service started", "opensearch", cfg.OpenSearchURL)

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	grpcServer.GracefulStop()
	logger.Info("Search service stopping")
}

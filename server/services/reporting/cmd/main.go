package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/MuhibNayem/Travio/server/api/proto/reporting/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/aggregation"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/clickhouse"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/consumer"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger.Init("reporting-service")
	logger.Info("Starting Reporting Service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load ClickHouse config
	chConfig := clickhouse.LoadConfigFromEnv()
	logger.Info("ClickHouse config loaded",
		"host", chConfig.Host,
		"database", chConfig.Database,
	)

	// Connect to ClickHouse
	chClient, err := clickhouse.NewClient(chConfig)
	if err != nil {
		logger.Fatal("Failed to connect to ClickHouse", "error", err)
	}
	defer chClient.Close()

	// Ping ClickHouse
	if err := chClient.Ping(ctx); err != nil {
		logger.Fatal("ClickHouse ping failed", "error", err)
	}
	logger.Info("ClickHouse connected")

	// Initialize schema
	if err := chClient.InitSchema(ctx); err != nil {
		logger.Warn("Schema initialization warning", "error", err)
	}

	// Initialize query engine
	queryEngine := query.NewEngine(chClient)

	// Initialize gRPC handler
	grpcHandler := handler.NewGrpcHandler(queryEngine)

	// Start Kafka consumer (if enabled)
	kafkaEnabled := os.Getenv("KAFKA_ENABLED") != "false"
	var kafkaConsumer *consumer.Consumer
	if kafkaEnabled {
		kafkaConfig := consumer.LoadConfigFromEnv()
		kafkaConsumer = consumer.NewConsumer(kafkaConfig, chClient)
		if err := kafkaConsumer.Start(ctx); err != nil {
			logger.Warn("Kafka consumer start failed, continuing without", "error", err)
			kafkaConsumer = nil
		} else {
			go kafkaConsumer.CleanupDeduplicationCache()
		}
	} else {
		logger.Info("Kafka consumer disabled")
	}

	// Start aggregation scheduler (if enabled)
	schedulerEnabled := os.Getenv("SCHEDULER_ENABLED") != "false"
	var scheduler *aggregation.Scheduler
	if schedulerEnabled {
		scheduler = aggregation.NewScheduler(chClient)
		if err := scheduler.Start(); err != nil {
			logger.Warn("Scheduler start failed", "error", err)
			scheduler = nil
		}
	} else {
		logger.Info("Aggregation scheduler disabled")
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterReportingServiceServer(grpcServer, grpcHandler)

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("reporting.v1.ReportingService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection
	reflection.Register(grpcServer)

	// Get port
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = os.Getenv("REPORTING_GRPC_PORT")
	}
	if port == "" {
		port = "50091"
	}

	// Start listening
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("Failed to listen", "port", port, "error", err)
	}

	// Start HTTP health server for Docker healthchecks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		logger.Info("Starting HTTP health endpoint", "port", 8080)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Warn("HTTP health server failed", "error", err)
		}
	}()

	logger.Info("Reporting Service listening", "port", port)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Shutting down Reporting Service")

		grpcServer.GracefulStop()

		if kafkaConsumer != nil {
			kafkaConsumer.Stop()
		}

		if scheduler != nil {
			scheduler.Stop()
		}
	}()

	// Start server
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", "error", err)
	}

	fmt.Println("Reporting Service stopped")
}

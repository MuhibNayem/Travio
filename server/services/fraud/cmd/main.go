package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fraud/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/client"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/profile"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/rag"
	"github.com/MuhibNayem/Travio/server/services/fraud/internal/service"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger.Init("fraud-service")
	logger.Info("Starting Fraud Service with User Profiling + RAG")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Z.AI config
	zaiConfig := client.LoadConfigFromEnv()
	if zaiConfig.APIKey == "" {
		logger.Warn("ZAI_API_KEY not set - fraud detection will use fallback (fail-open)")
	} else {
		logger.Info("Z.AI API configured", "base_url", zaiConfig.BaseURL)
	}

	// Connect to Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis connection failed, caching disabled", "error", err)
		redisClient = nil
	} else {
		logger.Info("Redis connected", "addr", redisAddr)
	}

	// Block threshold
	blockThreshold := 70
	if t := os.Getenv("FRAUD_THRESHOLD"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			blockThreshold = v
		}
	}

	// Initialize Z.AI client
	zaiClient := client.NewClient(zaiConfig, redisClient)

	// Initialize fraud service
	fraudSvc := service.NewFraudService(zaiClient, blockThreshold)

	// === User Profiling Setup ===
	if os.Getenv("PROFILE_ENABLED") != "false" {
		dbURL := os.Getenv("FRAUD_DATABASE_URL")
		if dbURL == "" {
			dbURL = os.Getenv("DATABASE_URL")
		}
		if dbURL != "" {
			db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
			if err != nil {
				logger.Warn("PostgreSQL connection failed, profiling disabled", "error", err)
			} else {
				logger.Info("PostgreSQL connected for user profiling")

				profileStore := profile.NewStore(db)
				if err := profileStore.AutoMigrate(); err != nil {
					logger.Warn("Profile table migration failed", "error", err)
				}

				deviationThreshold := 2.0
				if t := os.Getenv("PROFILE_DEVIATION_THRESHOLD"); t != "" {
					if v, err := strconv.ParseFloat(t, 64); err == nil {
						deviationThreshold = v
					}
				}
				profileAnalyzer := profile.NewAnalyzer(deviationThreshold)

				fraudSvc.WithProfiling(profileStore, profileAnalyzer)
				logger.Info("User profiling enabled", "deviation_threshold", deviationThreshold)
			}
		} else {
			logger.Info("FRAUD_DATABASE_URL not set, profiling disabled")
		}
	}

	// === RAG Setup ===
	if os.Getenv("RAG_ENABLED") != "false" {
		googleAPIKey := os.Getenv("GOOGLE_API_KEY")
		if googleAPIKey == "" {
			googleAPIKey = os.Getenv("GEMINI_API_KEY")
		}

		if googleAPIKey != "" {
			embedder := rag.NewEmbedder()

			// OpenSearch for vector similarity
			osURL := os.Getenv("OPENSEARCH_URL")
			if osURL == "" {
				osURL = "http://localhost:9200"
			}

			osClient, err := opensearch.NewClient(opensearch.Config{
				Addresses: []string{osURL},
				Username:  os.Getenv("OPENSEARCH_USERNAME"),
				Password:  os.Getenv("OPENSEARCH_PASSWORD"),
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			})
			if err != nil {
				logger.Warn("OpenSearch connection failed, RAG disabled", "error", err)
			} else {
				// Get DB for RAG case storage
				dbURL := os.Getenv("FRAUD_DATABASE_URL")
				if dbURL == "" {
					dbURL = os.Getenv("DATABASE_URL")
				}
				var ragDB *gorm.DB
				if dbURL != "" {
					ragDB, _ = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
				}

				topK := 5
				if t := os.Getenv("RAG_TOP_K"); t != "" {
					if v, err := strconv.Atoi(t); err == nil {
						topK = v
					}
				}

				minScore := 0.7
				if t := os.Getenv("RAG_SIMILARITY_THRESHOLD"); t != "" {
					if v, err := strconv.ParseFloat(t, 64); err == nil {
						minScore = v
					}
				}

				retriever := rag.NewRetriever(osClient, ragDB, embedder, topK, minScore)

				// Initialize OpenSearch index
				if err := retriever.InitializeIndex(ctx); err != nil {
					logger.Warn("Failed to initialize OpenSearch index", "error", err)
				}

				fraudSvc.WithRAG(embedder, retriever)
				logger.Info("RAG enabled", "top_k", topK, "min_score", minScore)
			}
		} else {
			logger.Info("GOOGLE_API_KEY not set, RAG disabled")
		}
	}

	// Initialize gRPC handler
	grpcHandler := handler.NewGrpcHandler(fraudSvc)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterFraudServiceServer(grpcServer, grpcHandler)

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("fraud.v1.FraudService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Get port
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = os.Getenv("FRAUD_GRPC_PORT")
	}
	if port == "" {
		port = "50090"
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

	logger.Info("Fraud Service listening", "port", port)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Shutting down Fraud Service")
		grpcServer.GracefulStop()
		if redisClient != nil {
			redisClient.Close()
		}
	}()

	// Start server
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", "error", err)
	}

	fmt.Println("Fraud Service stopped")
}

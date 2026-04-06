package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	paymentv1 "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	pricingv1 "github.com/MuhibNayem/Travio/server/api/proto/pricing/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/order/config"
	"github.com/MuhibNayem/Travio/server/services/order/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/order/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/order/internal/messaging"
	"github.com/MuhibNayem/Travio/server/services/order/internal/middleware"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
	"github.com/MuhibNayem/Travio/server/services/order/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.Init("order-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	dsn := cfg.Database.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
		panic(err)
	}
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping DB", "error", err)
		panic(err)
	}

	// GORM for Sagas and Checkout
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to GORM DB", "error", err)
		panic(err)
	}

	// Redis
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6388")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Kafka DLQ (TASK-009: Implemented)
	brokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	dlqTopic := getEnv("DLQ_TOPIC", "order-saga-dlq")
	dlq, err := messaging.NewKafkaDLQProducer(brokers, dlqTopic)
	if err != nil {
		logger.Warn("Failed to initialize DLQ producer, proceeding without", "error", err)
	} else {
		logger.Info("Initialized DLQ producer", "topic", dlqTopic)
	}

	// Service clients
	inventoryClient, err := clients.NewInventoryClient(cfg.Services.InventoryAddr)
	if err != nil {
		logger.Error("Failed to connect to inventory service", "error", err)
	}

	paymentClient, err := clients.NewPaymentClient(cfg.Services.PaymentAddr)
	if err != nil {
		logger.Error("Failed to connect to payment service", "error", err)
	}

	nidClient, err := clients.NewNIDClient(cfg.Services.NIDAddr)
	if err != nil {
		logger.Warn("NID service not available, NID verification disabled", "error", err)
	}

	notificationClient := clients.NewNotificationClient()

	subscriptionClient, err := clients.NewSubscriptionClient(cfg.Services.SubscriptionAddr)
	if err != nil {
		logger.Error("Failed to connect to subscription service", "error", err)
	}

	// Connect to Pricing Service for dynamic pricing
	var pricingClient pricingv1.PricingServiceClient
	pricingAddr := getEnv("PRICING_URL", "localhost:50058")
	if pricingConn, err := grpc.Dial(pricingAddr, grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
		pricingClient = pricingv1.NewPricingServiceClient(pricingConn)
		logger.Info("Connected to pricing service", "addr", pricingAddr)
	} else {
		logger.Warn("Failed to connect to pricing service, using fallback", "error", err)
	}

	// Connect to Payment Service for payment creation
	var _ paymentv1.PaymentServiceClient
	// (paymentClient already created above for saga)

	// Saga dependencies
	sagaDeps := &saga.BookingDependencies{
		NIDService:          nidClient,
		InventoryService:    inventoryClient,
		PaymentService:      paymentClient,
		SubscriptionService: subscriptionClient,
		NotificationSvc:     notificationClient,
	}

	// Repository and service (TASK-008: Checkout repo injected)
	orderRepo := repository.NewOrderRepository(db)
	checkoutRepo := repository.NewCheckoutRepository(gormDB)
	
	var dlqProducer messaging.DLQProducer
	if dlq != nil {
		dlqProducer = dlq
	}
	
	// Connect to Catalog Service for enriching orders with station names
	var catalogClient *clients.CatalogClient
	catalogAddr := getEnv("CATALOG_URL", "localhost:9082")
	if cc, cerr := clients.NewCatalogClient(catalogAddr); cerr == nil {
		catalogClient = cc
		logger.Info("Connected to catalog service for order enrichment", "addr", catalogAddr)
	} else {
		logger.Warn("Catalog service not available, orders will show IDs", "error", cerr)
	}

	// Connect to CRM Service for coupon validation (TASK-046)
	var couponValidator service.CRMClient
	crmAddr := getEnv("CRM_URL", "localhost:9094")
	if crmConn, err := grpc.Dial(crmAddr, grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
		couponValidator = clients.NewCRMClientFromConn(crmConn)
		logger.Info("Connected to CRM service for coupon validation", "addr", crmAddr)
	} else {
		logger.Warn("CRM service not available, coupon validation disabled", "error", err)
	}
	
	orderService := service.NewOrderService(db, gormDB, dlqProducer, orderRepo, checkoutRepo, sagaDeps, couponValidator, catalogClient)
	grpcHandler := handler.NewGrpcHandler(orderService)

	// Checkout Handler (TASK-006)
	checkoutHandler := handler.NewCheckoutHandler(
		service.NewCheckoutService(checkoutRepo, pricingClient, nil, couponValidator),
	)

	// Recover incomplete sagas on startup (TASK-010)
	go func() {
		time.Sleep(5 * time.Second) // Wait for DB to be ready
		if err := orderService.RecoverIncompleteSagas(context.Background()); err != nil {
			logger.Error("Failed to recover incomplete sagas", "error", err)
		}
	}()

	// Idempotency Middleware
	idempotency := middleware.NewIdempotencyMiddleware(redisClient)

	// HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register checkout routes (TASK-004, TASK-005, TASK-006)
	mux.Handle("/v1/checkout", http.StripPrefix("/v1/checkout", checkoutRoutes(checkoutHandler)))

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterOrderServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Order service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)

	// Wrap mux with Idempotency Middleware
	httpHandler := idempotency.Middleware(mux)

	srv.Start(httpHandler)
}

func checkoutRoutes(h *handler.CheckoutHandler) http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateCheckout(w, r)
		case http.MethodGet:
			h.ListCheckouts(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/hold/{holdId}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetCheckoutByHoldID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetCheckout(w, r)
		case http.MethodPatch:
			h.UpdateCheckout(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/{id}/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ConfirmCheckout(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return r
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

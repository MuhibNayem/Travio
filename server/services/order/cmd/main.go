package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/order/config"
	"github.com/MuhibNayem/Travio/server/services/order/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/order/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/order/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/order/internal/saga"
	"github.com/MuhibNayem/Travio/server/services/order/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("order-service")
	cfg := config.Load()

	// Database
	logger.Info("Connecting to PostgreSQL...")
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/travio_order?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
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
		logger.Error("Failed to connect to NID service", "error", err)
	}

	notificationClient := clients.NewNotificationClient()

	// Saga dependencies
	sagaDeps := &saga.BookingDependencies{
		NIDService:       nidClient,
		InventoryService: inventoryClient,
		PaymentService:   paymentClient,
		NotificationSvc:  notificationClient,
	}

	// Repository and service
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(db, orderRepo, sagaDeps)
	grpcHandler := handler.NewGrpcHandler(orderService)

	// HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterOrderServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Order service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

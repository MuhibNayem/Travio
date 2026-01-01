package main

import (
	"database/sql"
	"fmt"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fulfillment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/config"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/pdf"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/qr"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("fulfillment-service")
	cfg := config.Load()

	// Database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
	}

	// Initialize components
	ticketRepo := repository.NewTicketRepository(db)
	qrGenerator := qr.NewGenerator(cfg.QRSecretKey)
	pdfGenerator := pdf.NewGenerator(cfg.CompanyName, "")

	// Service
	fulfillmentService := service.NewFulfillmentService(ticketRepo, qrGenerator, pdfGenerator)
	grpcHandler := handler.NewGrpcHandler(fulfillmentService)

	// HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterFulfillmentServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Fulfillment service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

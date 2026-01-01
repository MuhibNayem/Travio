package main

import (
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/payment/config"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/service"
)

func main() {
	logger.Init("payment-service")
	cfg := config.Load()

	// Initialize payment gateways
	registry := gateway.NewRegistry()

	// SSLCommerz
	sslcommerz := gateway.NewSSLCommerz(gateway.SSLCommerzConfig{
		StoreID:     cfg.SSLCommerz.StoreID,
		StorePasswd: cfg.SSLCommerz.StorePasswd,
		IsSandbox:   cfg.SSLCommerz.IsSandbox,
	})
	registry.Register("sslcommerz", sslcommerz)
	registry.SetFallback(sslcommerz)

	// bKash
	bkash := gateway.NewBKash(gateway.BKashConfig{
		AppKey:    cfg.BKash.AppKey,
		AppSecret: cfg.BKash.AppSecret,
		Username:  cfg.BKash.Username,
		Password:  cfg.BKash.Password,
		IsSandbox: cfg.BKash.IsSandbox,
	})
	registry.Register("bkash", bkash)

	// Nagad
	nagad := gateway.NewNagad(gateway.NagadConfig{
		MerchantID:     cfg.Nagad.MerchantID,
		MerchantNumber: cfg.Nagad.MerchantNumber,
		PublicKey:      cfg.Nagad.PublicKey,
		PrivateKey:     cfg.Nagad.PrivateKey,
		IsSandbox:      cfg.Nagad.IsSandbox,
	})
	registry.Register("nagad", nagad)

	// Service and handler
	paymentService := service.NewPaymentService(registry)
	grpcHandler := handler.NewGrpcHandler(paymentService, registry)

	// HTTP mux for health and IPN webhooks
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	srv := server.New(cfg.Server)
	pb.RegisterPaymentServiceServer(srv.GRPC(), grpcHandler)

	logger.Info("Payment service starting", "grpc_port", cfg.Server.GRPCPort, "http_port", cfg.Server.HTTPPort)
	srv.Start(mux)
}

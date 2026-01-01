package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/identity/config"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/handler"
)

func main() {
	logger.Init("identity-service")
	cfg := config.Load()

	// Initialize handlers
	authHandler := handler.NewAuthHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/v1/orgs", authHandler.CreateOrganization)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start both gRPC and HTTP servers
	srv := server.New(cfg.Server)

	// Register gRPC services here (using generated code)
	// identity.RegisterIdentityServiceServer(srv.GRPC(), myGrpcHandler)

	srv.Start(mux)
}

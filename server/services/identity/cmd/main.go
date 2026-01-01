package main

import (
	"database/sql"
	"net/http"

	pb "github.com/MuhibNayem/Travio/server/api/proto/identity/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/identity/config"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/identity/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger.Init("identity-service")
	cfg := config.Load()

	// Database Setup
	logger.Info("Connecting to Postgres...")
	db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/travio_identity?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
	}

	// Dependency Injection
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrgRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepo, orgRepo, refreshTokenRepo)
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("/v1/auth/logout-all", authHandler.LogoutAll)
	mux.HandleFunc("/v1/auth/sessions", authHandler.GetSessions)
	mux.HandleFunc("/v1/orgs", authHandler.CreateOrganization)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start both gRPC and HTTP servers
	srv := server.New(cfg.Server)

	// Register gRPC services
	grpcHandler := handler.NewGrpcHandler(authService)
	pb.RegisterIdentityServiceServer(srv.GRPC(), grpcHandler)

	srv.Start(mux)
}

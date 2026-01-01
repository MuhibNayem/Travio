package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	GRPCPort        int
	HTTPPort        int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	config     Config
}

func New(cfg Config) *Server {
	return &Server{
		config:     cfg,
		grpcServer: grpc.NewServer(),
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// GRPC returns the internal grpc.Server to register services
func (s *Server) GRPC() *grpc.Server {
	return s.grpcServer
}

func (s *Server) Start(httpHandler http.Handler) {
	s.httpServer.Handler = httpHandler

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
		if err != nil {
			logger.Error("Failed to listen tcp for gRPC", "error", err)
			os.Exit(1)
		}
		logger.Info("Starting gRPC server", "port", s.config.GRPCPort)

		// Enable reflection for tools like grpcurl
		reflection.Register(s.grpcServer)

		if err := s.grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Start HTTP Server (Gateway/REST)
	go func() {
		if s.config.HTTPPort > 0 {
			logger.Info("Starting HTTP server", "port", s.config.HTTPPort)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("HTTP server failed", "error", err)
				os.Exit(1)
			}
		}
	}()

	<-ctx.Done()
	s.Shutdown()
}

func (s *Server) Shutdown() {
	logger.Info("Shutting down servers...")
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Shutdown gRPC
	s.grpcServer.GracefulStop()

	// Shutdown HTTP
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP Server forced to shutdown", "error", err)
	}
	logger.Info("Servers exited")
}

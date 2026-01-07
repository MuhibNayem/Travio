package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/config"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger.Init("api-gateway")
	cfg := config.Load()

	// Initialize Chi router
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Rate limiter (optional - depends on Redis availability)
	rateLimiter := middleware.NewRateLimiter(cfg.RedisURL, 100, 60) // 100 req/min
	defer rateLimiter.Close()
	r.Use(rateLimiter.Middleware)

	// Health endpoints (no auth required)
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)

	// Initialize gRPC handlers
	catalogHandler, err := handler.NewCatalogHandler(cfg.CatalogURL)
	if err != nil {
		logger.Error("Failed to connect to catalog service", "error", err)
	} else {
		defer catalogHandler.Close()
	}

	inventoryHandler, err := handler.NewInventoryHandler(cfg.InventoryURL)
	if err != nil {
		logger.Error("Failed to connect to inventory service", "error", err)
	} else {
		defer inventoryHandler.Close()
	}

	orderHandler, err := handler.NewOrderHandler(cfg.OrderURL)
	if err != nil {
		logger.Error("Failed to connect to order service", "error", err)
	} else {
		defer orderHandler.Close()
	}

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Auth routes - proxy to Identity service
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", proxyTo(cfg.IdentityURL))
			r.Post("/login", proxyTo(cfg.IdentityURL))
			r.Post("/refresh", proxyTo(cfg.IdentityURL))
			r.Post("/logout", proxyTo(cfg.IdentityURL))
			r.Post("/logout-all", proxyTo(cfg.IdentityURL))
			r.Get("/sessions", proxyTo(cfg.IdentityURL))
		})

		// Organization routes - proxy to Identity
		r.Post("/orgs", proxyTo(cfg.IdentityURL))

		// Catalog routes
		if catalogHandler != nil {
			r.Get("/stations", catalogHandler.ListStations)
			r.Get("/trips/search", catalogHandler.SearchTrips)
		}

		// Inventory routes
		if inventoryHandler != nil {
			r.Get("/trips/{tripId}/availability", inventoryHandler.CheckAvailability)
			r.Get("/trips/{tripId}/seatmap", inventoryHandler.GetSeatMap)
			r.Post("/holds", inventoryHandler.HoldSeats)
			r.Delete("/holds/{holdId}", inventoryHandler.ReleaseHold)
		}

		// Order routes
		if orderHandler != nil {
			r.Post("/orders", orderHandler.CreateOrder)
			r.Get("/orders", orderHandler.ListOrders)
			r.Get("/orders/{orderId}", orderHandler.GetOrder)
			r.Post("/orders/{orderId}/cancel", orderHandler.CancelOrder)
		}
	})

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("API Gateway starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	logger.Info("Server stopped")
}

// proxyTo creates a simple reverse proxy handler
func proxyTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy - in production use httputil.ReverseProxy
		client := &http.Client{Timeout: 10 * time.Second}

		proxyURL := target + r.URL.Path
		if r.URL.RawQuery != "" {
			proxyURL += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
		if err != nil {
			http.Error(w, "Proxy error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, v := range values {
				proxyReq.Header.Add(key, v)
			}
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Upstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}
}

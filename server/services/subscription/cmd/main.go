package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger first
	logger.Init("subscription-service")

	// 1. Load Config
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found")
	}

	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "travio_subscription"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50060" // Default port for Subscription
	}

	// 2. Connect to DB
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to connect to DB", "error", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping DB", "error", err)
	}
	logger.Info("Connected to PostgreSQL", "db", dbName)

	// 3. Init Layers
	repo := repository.NewPostgresRepository(db)
	svc := service.NewSubscriptionService(repo)

	// Seed Market-Fit Plans (Bangladesh Context)
	ctx := context.Background()
	plans := []struct {
		ID       string
		Name     string
		Desc     string
		Price    int64
		Features map[string]string
		IsActive bool
	}{
		{
			ID:    "plan_free",
			Name:  "Shuru (Starter)",
			Desc:  "Perfect for single bus owners. 3 Round Trips/Day included.",
			Price: 0,
			Features: map[string]string{
				"max_trips_per_month": "180", // 6 trips/day * 30 = 180 (3 Round Trips)
				"max_schedule_days":   "10",  // 10 days in advance
				"commission_rate":     "5.0", // 5% platform fee on online sales
				"counter_sales":       "unlimited",
				"analytics":           "basic",
			},
		},
		{
			ID:    "plan_pro",
			Name:  "Goti (Growth)",
			Desc:  "For growing fleets (up to 10 buses). Lower fees & SMS alerts.",
			Price: 250000, // 2,500 BDT
			Features: map[string]string{
				"max_trips_per_month": "1200", // ~40 trips/day (20 Round Trips)
				"max_schedule_days":   "30",   // 30 days in advance
				"commission_rate":     "2.5",  // 2.5% platform fee
				"counter_sales":       "unlimited",
				"analytics":           "advanced",
				"sms_alerts":          "true",
			},
		},
		{
			ID:    "plan_enterprise",
			Name:  "Bishal (Enterprise)",
			Desc:  "Unlimited scale for national carriers. Dedicated support.",
			Price: 1500000, // 15,000 BDT
			Features: map[string]string{
				"max_trips_per_month": "-1",  // Unlimited
				"max_schedule_days":   "90",  // 90 days in advance
				"commission_rate":     "1.0", // 1% platform fee
				"counter_sales":       "unlimited",
				"analytics":           "enterprise",
				"sms_alerts":          "true",
				"api_access":          "true",
				"custom_branding":     "true",
			},
		},
	}

	for _, p := range plans {
		if existing, _ := svc.GetPlan(ctx, p.ID); existing == nil {
			_, err := svc.CreatePlan(ctx, p.ID, p.Name, p.Desc, p.Price, "month", p.Features, 0)
			if err != nil {
				logger.Error("Failed to seed plan", "id", p.ID, "error", err)
			} else {
				logger.Info("Seeded plan", "id", p.ID, "name", p.Name)
			}
		}
	}

	// 4. Start HTTP health server for Docker healthchecks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		httpPort := os.Getenv("HTTP_PORT")
		if httpPort == "" {
			httpPort = "8060"
		}
		logger.Info("Starting HTTP health endpoint", "port", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Warn("HTTP health server failed", "error", err)
		}
	}()

	// 5. Start Server
	go func() {
		if err := handler.StartGRPCServer(grpcPort, svc); err != nil {
			logger.Fatal("Failed to start gRPC server", "error", err)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Subscription Service...")
}

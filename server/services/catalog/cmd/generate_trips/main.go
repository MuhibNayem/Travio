package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/catalog/config"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/events"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Init("generate-trips")
	cfg := config.Load()

	// Database Setup
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Redis Setup
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Repositories
	stationRepo := repository.NewStationRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	publisher := events.NewPublisher(db)
	tripRepo := repository.NewTripRepository(db, publisher)
	scheduleRepo := repository.NewScheduleRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Mock Entitlement & Fleet (not needed for simple generation if we bypass checks or use valid data)
	entChecker := entitlement.NewCachedChecker(rdb, nil, entitlement.DefaultConfig())
	fleetClient, _ := clients.NewFleetClient(cfg.FleetURL)

	catalogService := service.NewCatalogService(stationRepo, routeRepo, tripRepo, scheduleRepo, entChecker, fleetClient, auditRepo)

	// 1. List all schedules
	ctx := context.Background()
	// Using empty filters to get all schedules
	// Note: Pagination hack, get first 1000
	schedules, _, _, err := catalogService.ListSchedules(ctx, "04303fc7-19f1-4ab6-8961-102b9c836429", "", "", 1000, "")
	if err != nil {
		log.Fatalf("Failed to list schedules: %v", err)
	}

	fmt.Printf("Found %d schedules to process\n", len(schedules))

	// 2. Generate trips for next 30 days
	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	for _, sched := range schedules {
		fmt.Printf("Generating trips for schedule %s (%s - %s)\n", sched.ID, startDate, endDate)
		_, count, err := catalogService.GenerateTripInstances(ctx, sched.ID, sched.OrganizationID, startDate, endDate)
		if err != nil {
			log.Printf("Error generating trips for schedule %s: %v\n", sched.ID, err)
			continue
		}
		fmt.Printf("Generated %d trips\n", count)

		// Force manual outbox publishing if the service doesn't run the relay?
		// No, the relay runs in the main service container.
		// We just inserted events into DB. The main container (if running) will pick them up.
	}

	// Start a temporary relay just to be sure, or rely on the running service
	log.Println("Trips generated. Events inserted into outbox.")
}

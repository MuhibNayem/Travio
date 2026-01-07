package main

import (
	"os"
	"strings"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/consumer"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/provider"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/service"
)

func main() {
	logger.Init("notification-service")
	logger.Info("Notification service starting")

	// Initialize providers (use console for dev, real providers in production)
	emailProvider := provider.NewConsoleEmailProvider()
	smsProvider := provider.NewConsoleSMSProvider()

	// Create notification service
	notificationSvc := service.NewNotificationService(emailProvider, smsProvider)

	// Get Kafka brokers from environment
	kafkaBrokers := getKafkaBrokers()
	if len(kafkaBrokers) == 0 {
		logger.Error("KAFKA_BROKERS not configured")
		return
	}

	// Create and start event consumer
	eventConsumer, err := consumer.NewEventConsumer(kafkaBrokers, notificationSvc)
	if err != nil {
		logger.Error("Failed to create event consumer", "error", err)
		return
	}

	if err := eventConsumer.Start(); err != nil {
		logger.Error("Failed to start event consumer", "error", err)
		return
	}

	logger.Info("Notification service started, listening for events")

	// Block forever (or until signal)
	select {}
}

// getKafkaBrokers reads Kafka broker addresses from environment
func getKafkaBrokers() []string {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return nil
	}
	return strings.Split(brokersEnv, ",")
}

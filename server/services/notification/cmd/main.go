package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/consumer"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/provider"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/service"
)

const (
	EmailRatePerSecond = 10 // SES default is ~14/s
	SMSRatePerSecond   = 5  // Twilio default is ~10/s
)

func main() {
	logger.Init("notification-service")
	logger.Info("Notification service starting")

	// Initialize base providers (use console for dev, real providers in production)
	baseEmailProvider := provider.NewConsoleEmailProvider()
	baseSMSProvider := provider.NewConsoleSMSProvider()

	// Wrap with rate limiters to prevent provider suspension
	emailProvider := provider.NewRateLimitedEmailProvider(baseEmailProvider, EmailRatePerSecond)
	smsProvider := provider.NewRateLimitedSMSProvider(baseSMSProvider, SMSRatePerSecond)

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

	logger.Info("Notification service started", "email_rps", EmailRatePerSecond, "sms_rps", SMSRatePerSecond)

	// Start HTTP health server for Docker healthchecks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		httpPort := os.Getenv("HTTP_PORT")
		if httpPort == "" {
			httpPort = "8090"
		}
		logger.Info("Starting HTTP health endpoint", "port", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			logger.Warn("HTTP health server failed", "error", err)
		}
	}()

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

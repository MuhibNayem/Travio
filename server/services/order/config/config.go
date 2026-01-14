package config

import (
	"os"
	"strconv"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server   server.Config
	Database DatabaseConfig
	Services ServicesConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServicesConfig struct {
	InventoryAddr    string
	PaymentAddr      string
	NIDAddr          string
	NotificationAddr string
	SubscriptionAddr string
	CatalogAddr      string
	PricingAddr      string
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:        getEnvInt("GRPC_PORT", 9084),
			HTTPPort:        getEnvInt("HTTP_PORT", 8084),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "travio_order"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Services: ServicesConfig{
			InventoryAddr:    getEnv("INVENTORY_URL", "localhost:9083"),
			PaymentAddr:      getEnv("PAYMENT_URL", "localhost:9085"),
			NIDAddr:          getEnv("IDENTITY_URL", "localhost:9081"),
			NotificationAddr: getEnv("NOTIFICATION_URL", "localhost:9090"),
			SubscriptionAddr: getEnv("SUBSCRIPTION_URL", "localhost:50060"),
			CatalogAddr:      getEnv("CATALOG_URL", "localhost:9082"),
			PricingAddr:      getEnv("PRICING_URL", "localhost:9095"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

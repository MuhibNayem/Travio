package config

import (
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
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:        9085,
			HTTPPort:        8085,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "travio_order",
			SSLMode:  "disable",
		},
		Services: ServicesConfig{
			InventoryAddr:    "localhost:9083",
			PaymentAddr:      "localhost:9084",
			NIDAddr:          "localhost:9086",
			NotificationAddr: "localhost:9087",
		},
	}
}

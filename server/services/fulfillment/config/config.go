package config

import (
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server      server.Config
	Database    DatabaseConfig
	QRSecretKey string
	CompanyName string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:        9088,
			HTTPPort:        8088,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host: "localhost", Port: 5432, User: "postgres",
			Password: "postgres", DBName: "travio_fulfillment", SSLMode: "disable",
		},
		QRSecretKey: "your-super-secret-key-for-qr-signing",
		CompanyName: "Travio",
	}
}

package config

import (
	"os"
	"strconv"

	"github.com/MuhibNayem/Travio/server/pkg/database/postgres"
	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server   server.Config
	Database postgres.Config
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:    getEnvInt("GRPC_PORT", 50059),
			HTTPPort:    getEnvInt("HTTP_PORT", 8059),
			TLSCertFile: os.Getenv("TLS_CERT_FILE"),
			TLSKeyFile:  os.Getenv("TLS_KEY_FILE"),
			TLSCAFile:   os.Getenv("TLS_CA_FILE"),
		},
		Database: postgres.Config{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   getEnv("POSTGRES_DB", "travio_operator"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

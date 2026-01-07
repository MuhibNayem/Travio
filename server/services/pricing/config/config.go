package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort int
	DBHost   string
	DBPort   int
	DBUser   string
	DBPass   string
	DBName   string
}

func Load() *Config {
	return &Config{
		GRPCPort: getEnvInt("GRPC_PORT", 50058),
		DBHost:   getEnv("DB_HOST", "localhost"),
		DBPort:   getEnvInt("DB_PORT", 5432),
		DBUser:   getEnv("DB_USER", "postgres"),
		DBPass:   getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "travio_pricing"),
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

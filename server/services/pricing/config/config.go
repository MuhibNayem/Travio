package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort  int
	HTTPPort  int
	DBHost    string
	DBPort    int
	DBUser    string
	DBPass    string
	DBName    string
	RedisAddr string
}

func Load() *Config {
	return &Config{
		GRPCPort:  getEnvInt("GRPC_PORT", 50058),
		HTTPPort:  getEnvInt("HTTP_PORT", 8058),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnvInt("DB_PORT", 5432),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "travio_pricing"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
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

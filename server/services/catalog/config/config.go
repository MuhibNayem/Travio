package config

import (
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/utils"
)

type Config struct {
	Server   server.Config
	Database DatabaseConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:        utils.GetEnvAsInt("GRPC_PORT", 9082),
			HTTPPort:        utils.GetEnvAsInt("HTTP_PORT", 8082),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnvAsInt("DB_PORT", 5432),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			DBName:   utils.GetEnv("DB_NAME", "travio_catalog"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6379"),
			Password: utils.GetEnv("REDIS_PASSWORD", ""),
			DB:       utils.GetEnvAsInt("REDIS_DB", 0),
		},
	}
}

package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort        int
	HTTPPort        int
	Database        DatabaseConfig
	RetentionDays   int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return "postgres://" + d.User + ":" + d.Password + "@" + d.Host + ":" + strconv.Itoa(d.Port) + "/" + d.DBName + "?sslmode=" + d.SSLMode
}

func Load() *Config {
	return &Config{
		GRPCPort: getEnvInt("GRPC_PORT", 50092),
		HTTPPort: getEnvInt("HTTP_PORT", 8092),
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "travio_audit"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RetentionDays: getEnvInt("AUDIT_RETENTION_DAYS", 90),
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

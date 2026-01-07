package config

import (
	"os"
	"strconv"
)

// Config holds queue service configuration
type Config struct {
	HTTPPort    int
	GRPCPort    int
	RedisAddr   string
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string
}

// Load loads configuration from environment
func Load() *Config {
	httpPort, _ := strconv.Atoi(getEnv("HTTP_PORT", "8087"))
	grpcPort, _ := strconv.Atoi(getEnv("GRPC_PORT", "9087"))

	return &Config{
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		TLSCertFile: getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:  getEnv("TLS_KEY_FILE", ""),
		TLSCAFile:   getEnv("TLS_CA_FILE", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

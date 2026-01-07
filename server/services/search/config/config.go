package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	OpenSearchURL string
	KafkaBrokers  []string
	GroupID       string
	GRPCPort      int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	return &Config{
		OpenSearchURL: getEnv("OPENSEARCH_URL", "http://localhost:9200"),
		KafkaBrokers:  strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		GroupID:       getEnv("KAFKA_GROUP_ID", "search-service"),
		GRPCPort:      getEnvInt("GRPC_PORT", 9085),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		// handle error or just assume valid for now
		// using atoi implementation simplified
		// ...
		// actually let's implement basic atoi
		var res int
		fmt.Sscanf(v, "%d", &res)
		return res
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

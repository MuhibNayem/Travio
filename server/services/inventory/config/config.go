package config

import (
	"os"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server   server.Config
	ScyllaDB ScyllaDBConfig
	Redis    RedisConfig
}

type ScyllaDBConfig struct {
	Hosts       []string
	Keyspace    string
	Consistency string
	Timeout     time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	scyllaHosts := []string{"localhost:9042"}
	if env := os.Getenv("SCYLLA_HOSTS"); env != "" {
		scyllaHosts = strings.Split(env, ",")
	}

	return &Config{
		Server: server.Config{
			GRPCPort:        9083,
			HTTPPort:        8083,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		ScyllaDB: ScyllaDBConfig{
			Hosts:       scyllaHosts,
			Keyspace:    "travio_inventory",
			Consistency: "QUORUM",
			Timeout:     5 * time.Second,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       1,
		},
	}
}

package config

import (
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
	return &Config{
		Server: server.Config{
			GRPCPort:        9083,
			HTTPPort:        8083,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		ScyllaDB: ScyllaDBConfig{
			Hosts:       []string{"localhost:9042"},
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

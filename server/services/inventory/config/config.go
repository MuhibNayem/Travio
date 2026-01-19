package config

import (
	"os"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server       server.Config
	ScyllaDB     ScyllaDBConfig
	Redis        RedisConfig
	KafkaBrokers []string
	FleetURL     string
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

	scyllaConsistency := "ONE" // Default for single-node dev environments
	if env := os.Getenv("SCYLLA_CONSISTENCY"); env != "" {
		scyllaConsistency = env
	}

	kafkaBrokers := []string{"localhost:9092"}
	if env := os.Getenv("KAFKA_BROKERS"); env != "" {
		kafkaBrokers = strings.Split(env, ",")
	}

	fleetURL := "localhost:50053"
	if env := os.Getenv("FLEET_URL"); env != "" {
		fleetURL = env
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
			Consistency: scyllaConsistency,
			Timeout:     5 * time.Second,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       1,
		},
		KafkaBrokers: kafkaBrokers,
		FleetURL:     fleetURL,
	}
}

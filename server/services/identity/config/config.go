package config

import (
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/database/postgres"
	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server   server.Config
	Database postgres.Config
}

func Load() Config {
	// In reality this would load from env vars (e.g. using kelseyhightower/envconfig)
	return Config{
		Server: server.Config{
			HTTPPort:        8081,
			GRPCPort:        9081, // gRPC port
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Database: postgres.Config{
			Host:   "localhost",
			Port:   5432,
			User:   "postgres",
			DBName: "identity_db",
		},
	}
}

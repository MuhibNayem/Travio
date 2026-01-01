package config

import (
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server server.Config
}

func Load() Config {
	return Config{
		Server: server.Config{
			HTTPPort:        8093,
			GRPCPort:        9093,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
	}
}

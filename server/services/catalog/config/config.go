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
			HTTPPort:        8082,
			GRPCPort:        9082,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
	}
}

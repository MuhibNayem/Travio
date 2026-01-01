package scylladb

import (
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	// "github.com/gocql/gocql"
)

type Config struct {
	Hosts    []string
	Keyspace string
}

func Connect(cfg Config) (any, error) {
	logger.Info("Connecting to ScyllaDB", "hosts", cfg.Hosts, "keyspace", cfg.Keyspace)
	// cluster := gocql.NewCluster(cfg.Hosts...)
	// cluster.Keyspace = cfg.Keyspace
	// session, err := cluster.CreateSession()
	return nil, nil
}

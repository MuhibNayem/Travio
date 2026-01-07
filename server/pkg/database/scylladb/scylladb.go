package scylladb

import (
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/gocql/gocql"
)

// Config holds ScyllaDB configuration
type Config struct {
	Hosts          []string
	Keyspace       string
	Consistency    string
	Timeout        time.Duration
	ConnectTimeout time.Duration
}

// NewSession creates a new ScyllaDB session
func NewSession(cfg Config) (*gocql.Session, error) {
	logger.Info("Connecting to ScyllaDB", "hosts", cfg.Hosts, "keyspace", cfg.Keyspace)

	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Timeout = cfg.Timeout
	cluster.ConnectTimeout = cfg.ConnectTimeout
	cluster.Consistency = parseConsistency(cfg.Consistency)

	// Production hardening
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{NumRetries: 3, Min: 100 * time.Millisecond, Max: 1 * time.Second}
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.MaxWaitSchemaAgreement = 2 * time.Minute // Wait for schema agreement
	cluster.ReconnectInterval = 5 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func parseConsistency(c string) gocql.Consistency {
	switch c {
	case "ANY":
		return gocql.Any
	case "ONE":
		return gocql.One
	case "TWO":
		return gocql.Two
	case "THREE":
		return gocql.Three
	case "QUORUM":
		return gocql.Quorum
	case "ALL":
		return gocql.All
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum
	case "EACH_QUORUM":
		return gocql.EachQuorum
	case "LOCAL_ONE":
		return gocql.LocalOne
	default:
		return gocql.Quorum
	}
}

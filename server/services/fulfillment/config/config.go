package config

import (
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server      server.Config
	Database    DatabaseConfig
	MinIO       MinIOConfig
	QRSecretKey string
	CompanyName string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:        9088,
			HTTPPort:        8088,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			// mTLS: Set via environment or use defaults for development
			TLSCertFile: getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:  getEnv("TLS_KEY_FILE", ""),
			TLSCAFile:   getEnv("TLS_CA_FILE", ""),
		},
		Database: DatabaseConfig{
			Host: "localhost", Port: 5432, User: "postgres",
			Password: "postgres", DBName: "travio_fulfillment", SSLMode: "disable",
		},
		MinIO: MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			UseSSL:          false,
			BucketName:      "tickets",
		},
		QRSecretKey: "your-super-secret-key-for-qr-signing",
		CompanyName: "Travio",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

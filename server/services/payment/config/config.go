package config

import (
	"os"
	"strconv"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server     server.Config
	Database   DatabaseConfig
	SSLCommerz SSLCommerzConfig
	BKash      BKashConfig
	Nagad      NagadConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type SSLCommerzConfig struct {
	StoreID     string
	StorePasswd string
	IsSandbox   bool
}

type BKashConfig struct {
	AppKey    string
	AppSecret string
	Username  string
	Password  string
	IsSandbox bool
}

type NagadConfig struct {
	MerchantID     string
	MerchantNumber string
	PublicKey      string
	PrivateKey     string
	IsSandbox      bool
}

func Load() *Config {
	return &Config{
		Server: server.Config{
			GRPCPort:    getEnvInt("GRPC_PORT", 9085),
			HTTPPort:    getEnvInt("HTTP_PORT", 8085),
			ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			// mTLS: Set via environment or use defaults for development
			TLSCertFile: getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:  getEnv("TLS_KEY_FILE", ""),
			TLSCAFile:   getEnv("TLS_CA_FILE", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "travio_payment"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		SSLCommerz: SSLCommerzConfig{
			StoreID:     getEnv("SSLCOMMERZ_STORE_ID", "your_store_id"),
			StorePasswd: getEnv("SSLCOMMERZ_STORE_PASSWD", "your_store_passwd"),
			IsSandbox:   getEnvBool("SSLCOMMERZ_SANDBOX", true),
		},
		BKash: BKashConfig{
			AppKey: "your_app_key", AppSecret: "your_app_secret",
			Username: "your_username", Password: "your_password", IsSandbox: true,
		},
		Nagad: NagadConfig{
			MerchantID: "your_merchant_id", MerchantNumber: "your_number",
			PublicKey: "nagad_public_key", PrivateKey: "your_private_key", IsSandbox: true,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

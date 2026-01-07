package config

import (
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/server"
)

type Config struct {
	Server     server.Config
	SSLCommerz SSLCommerzConfig
	BKash      BKashConfig
	Nagad      NagadConfig
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
			GRPCPort: 9084, HTTPPort: 8084,
			ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			// mTLS: Set via environment or use defaults for development
			TLSCertFile: getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:  getEnv("TLS_KEY_FILE", ""),
			TLSCAFile:   getEnv("TLS_CA_FILE", ""),
		},
		SSLCommerz: SSLCommerzConfig{
			StoreID: "your_store_id", StorePasswd: "your_store_passwd", IsSandbox: true,
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

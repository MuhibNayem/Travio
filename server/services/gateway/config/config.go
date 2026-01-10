package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort        int
	IdentityURL     string
	CatalogURL      string
	InventoryURL    string
	OrderURL        string
	PaymentURL      string
	FulfillmentURL  string
	SearchURL       string
	PricingURL      string
	OperatorURL     string
	SubscriptionURL string
	QueueURL        string
	RedisURL        string
	JWTSecret       string
	AllowedOrigins  []string
	// mTLS for gRPC clients
	TLSCertFile  string
	TLSKeyFile   string
	TLSCAFile    string
	FraudURL     string
	ReportingURL string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	return &Config{
		HTTPPort:        port,
		IdentityURL:     getEnv("IDENTITY_URL", "localhost:8081"),
		CatalogURL:      getEnv("CATALOG_URL", "localhost:9082"),
		InventoryURL:    getEnv("INVENTORY_URL", "localhost:9083"),
		OrderURL:        getEnv("ORDER_URL", "localhost:9084"),
		PaymentURL:      getEnv("PAYMENT_URL", "localhost:9085"),
		FulfillmentURL:  getEnv("FULFILLMENT_URL", "localhost:9086"),
		SearchURL:       getEnv("SEARCH_URL", "localhost:9088"),
		PricingURL:      getEnv("PRICING_URL", "localhost:50058"),
		OperatorURL:     getEnv("OPERATOR_URL", "localhost:50059"),
		SubscriptionURL: getEnv("SUBSCRIPTION_URL", "localhost:50060"),
		QueueURL:        getEnv("QUEUE_URL", "localhost:9087"),
		RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
		FraudURL:        getEnv("FRAUD_URL", "localhost:50090"),
		ReportingURL:    getEnv("REPORTING_URL", "localhost:50091"),
		JWTSecret:       getEnv("JWT_SECRET", "travio-secret-key-change-in-production"),
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
		},
		TLSCertFile: getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:  getEnv("TLS_KEY_FILE", ""),
		TLSCAFile:   getEnv("TLS_CA_FILE", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

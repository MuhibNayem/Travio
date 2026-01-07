package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort       int
	IdentityURL    string
	CatalogURL     string
	InventoryURL   string
	OrderURL       string
	PaymentURL     string
	FulfillmentURL string
	RedisURL       string
	AllowedOrigins []string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	return &Config{
		HTTPPort:       port,
		IdentityURL:    getEnv("IDENTITY_URL", "localhost:8081"),
		CatalogURL:     getEnv("CATALOG_URL", "localhost:9082"),
		InventoryURL:   getEnv("INVENTORY_URL", "localhost:9083"),
		OrderURL:       getEnv("ORDER_URL", "localhost:9084"),
		PaymentURL:     getEnv("PAYMENT_URL", "localhost:9085"),
		FulfillmentURL: getEnv("FULFILLMENT_URL", "localhost:9086"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

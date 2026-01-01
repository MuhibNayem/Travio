package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("fraud-service")
	logger.Info("Starting Fraud Service", "port", 8087)
	http.ListenAndServe(":8087", nil)
}

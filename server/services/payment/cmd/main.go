package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("payment-service")
	logger.Info("Starting Payment Service", "port", 8085)
	http.ListenAndServe(":8085", nil)
}

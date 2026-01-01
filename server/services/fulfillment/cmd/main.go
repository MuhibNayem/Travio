package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("fulfillment-service")
	logger.Info("Starting Fulfillment Service", "port", 8088)
	http.ListenAndServe(":8088", nil)
}

package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("inventory-service")
	logger.Info("Starting Inventory Service", "port", 8083)

	http.ListenAndServe(":8083", nil)
}

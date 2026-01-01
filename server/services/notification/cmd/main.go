package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("notification-service")
	logger.Info("Starting Notification Service", "port", 8086)
	http.ListenAndServe(":8086", nil)
}

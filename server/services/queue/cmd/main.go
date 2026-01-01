package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("queue-service")
	logger.Info("Starting Queue Service", "port", 8091)
	http.ListenAndServe(":8091", nil)
}

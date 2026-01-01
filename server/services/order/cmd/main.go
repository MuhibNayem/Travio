package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("order-service")
	logger.Info("Starting Order Service", "port", 8084)

	http.ListenAndServe(":8084", nil)
}

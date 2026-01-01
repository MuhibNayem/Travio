package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("reporting-service")
	logger.Info("Starting Reporting Service", "port", 8090)
	http.ListenAndServe(":8090", nil)
}

package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("audit-service")
	logger.Info("Starting Audit Service", "port", 8092)
	http.ListenAndServe(":8092", nil)
}

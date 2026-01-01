package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("search-service")
	logger.Info("Starting Search Service", "port", 8089)
	http.ListenAndServe(":8089", nil)
}

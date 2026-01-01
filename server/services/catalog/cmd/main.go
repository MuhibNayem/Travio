package main

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/pkg/server"
	"github.com/MuhibNayem/Travio/server/services/catalog/config"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/handler"
)

func main() {
	logger.Init("catalog-service")
	cfg := config.Load()

	h := handler.NewCatalogHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", h.ListEvents)
	mux.HandleFunc("/v1/trips", h.CreateTrip)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := server.New(cfg.Server)
	srv.Start(mux)
}

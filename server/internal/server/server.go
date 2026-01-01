package server

import (
	"context"
	"log"
	"net/http"

	"github.com/amnayem/Travio/server/internal/config"
	"github.com/amnayem/Travio/server/internal/handler"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Health)

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("shutting down server")
	return s.httpServer.Shutdown(ctx)
}

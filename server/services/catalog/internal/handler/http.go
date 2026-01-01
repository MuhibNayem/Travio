package handler

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

type CatalogHandler struct{}

func NewCatalogHandler() *CatalogHandler {
	return &CatalogHandler{}
}

func (h *CatalogHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	logger.Info("ListEvents endpoint called")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"events":[]}`))
}

func (h *CatalogHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	logger.Info("CreateTrip endpoint called")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"trip_id":"trip-123"}`))
}

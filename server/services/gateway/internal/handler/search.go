package handler

import (
	"net/http"
	"strconv"

	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/render"
)

type SearchHandler struct {
	client *client.SearchClient
}

func NewSearchHandler(client *client.SearchClient) *SearchHandler {
	return &SearchHandler{client: client}
}

func (h *SearchHandler) SearchTrips(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")
	date := r.URL.Query().Get("date")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}

	resp, err := h.client.SearchTrips(r.Context(), query, fromID, toID, date, limit, offset)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, resp)
}

func (h *SearchHandler) SearchStations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit <= 0 {
		limit = 10
	}

	resp, err := h.client.SearchStations(r.Context(), query, limit)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, resp)
}

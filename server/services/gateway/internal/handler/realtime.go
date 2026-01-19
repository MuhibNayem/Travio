package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MuhibNayem/Travio/server/services/gateway/internal/realtime"
	"github.com/go-chi/chi/v5"
)

type RealtimeHandler struct {
	manager *realtime.Manager
}

func NewRealtimeHandler(manager *realtime.Manager) *RealtimeHandler {
	return &RealtimeHandler{manager: manager}
}

// SubscribeTripUpdates handles SSE connections for trip updates
func (h *RealtimeHandler) SubscribeTripUpdates(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "tripId")
	if tripID == "" {
		http.Error(w, "Trip ID is required", http.StatusBadRequest)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Flush headers immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Subscribe to updates
	msgChan := h.manager.Subscribe(tripID)

	// Send initial connection message (optional, good for debugging)
	// fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	// flusher.Flush()

	defer func() {
		h.manager.Unsubscribe(tripID, msgChan)
	}()

	// Heartbeat ticker to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-msgChan:
			// msg is expected to be a full JSON string or data
			// The format "data: <payload>\n\n" is standard SSE
			// Our consumer code formats it as: {"type":"...", "data": ...} in the payload string?
			// Let's check consumer.go:
			// sseMessage := fmt.Sprintf(`{"type":"%s","data":%s}`, event.Type, string(payloadBytes))
			// So we just need to wrap it in "data: ... \n\n"

			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()

		case <-ticker.C:
			// Send comment to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

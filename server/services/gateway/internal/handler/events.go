package handler

import (
	"encoding/json"
	"net/http"

	eventsv1 "github.com/MuhibNayem/Travio/server/api/proto/events/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

type EventsHandler struct {
	client *client.EventsClient
}

func NewEventsHandler(c *client.EventsClient) *EventsHandler {
	return &EventsHandler{client: c}
}

// --- Venues ---

func (h *EventsHandler) CreateVenue(w http.ResponseWriter, r *http.Request) {
	var req eventsv1.CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	venue, err := h.client.CreateVenue(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venue)
}

func (h *EventsHandler) GetVenue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	venue, err := h.client.GetVenue(r.Context(), &eventsv1.GetVenueRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venue)
}

func (h *EventsHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	resp, err := h.client.ListVenues(r.Context(), &eventsv1.ListVenuesRequest{OrganizationId: orgID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *EventsHandler) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req eventsv1.UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Id = id

	venue, err := h.client.UpdateVenue(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venue)
}

// --- Events ---

func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req eventsv1.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.client.CreateEvent(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	event, err := h.client.GetEvent(r.Context(), &eventsv1.GetEventRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (h *EventsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	resp, err := h.client.ListEvents(r.Context(), &eventsv1.ListEventsRequest{OrganizationId: orgID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req eventsv1.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Id = id

	e, err := h.client.UpdateEvent(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *EventsHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	e, err := h.client.PublishEvent(r.Context(), &eventsv1.PublishEventRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *EventsHandler) SearchEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.client.SearchEvents(r.Context(), &eventsv1.SearchEventsRequest{
		Query:     q.Get("q"),
		City:      q.Get("city"),
		Category:  q.Get("category"),
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
		PageToken: q.Get("page_token"),
		PageSize:  10, // Default
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Tickets ---

func (h *EventsHandler) CreateTicketType(w http.ResponseWriter, r *http.Request) {
	var req eventsv1.CreateTicketTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tt, err := h.client.CreateTicketType(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tt)
}

func (h *EventsHandler) ListTicketTypes(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	resp, err := h.client.ListTicketTypes(r.Context(), &eventsv1.ListTicketTypesRequest{EventId: eventID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

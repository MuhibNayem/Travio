package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/go-chi/chi/v5"
)

// FulfillmentHandler handles ticket/fulfillment requests via gRPC
type FulfillmentHandler struct {
	client *client.FulfillmentClient
}

// NewFulfillmentHandler creates a new fulfillment handler with gRPC client
func NewFulfillmentHandler(fulfillmentClient *client.FulfillmentClient) *FulfillmentHandler {
	return &FulfillmentHandler{client: fulfillmentClient}
}

// GetTicket returns ticket details via gRPC
func (h *FulfillmentHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.client.GetTicket(r.Context(), ticketID)
	if err != nil {
		logger.Error("Failed to get ticket", "error", err)
		http.Error(w, `{"error": "fulfillment service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// DownloadTicket returns the ticket PDF via gRPC
func (h *FulfillmentHandler) DownloadTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	pdfResp, err := h.client.GetTicketPDF(r.Context(), ticketID)
	if err != nil {
		logger.Error("Failed to download ticket", "error", err)
		http.Error(w, `{"error": "fulfillment service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", pdfResp.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+pdfResp.Filename)
	w.Write(pdfResp.PdfData)
}

// GetOrderTickets returns all tickets for an order via gRPC
func (h *FulfillmentHandler) GetOrderTickets(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")

	tickets, err := h.client.ListTickets(r.Context(), orderID)
	if err != nil {
		logger.Error("Failed to get order tickets", "error", err)
		http.Error(w, `{"error": "fulfillment service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tickets": tickets})
}

package qr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"

	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/domain"
	"github.com/skip2/go-qrcode"
)

// Generator creates and validates QR codes for tickets
type Generator struct {
	secretKey []byte
}

func NewGenerator(secretKey string) *Generator {
	return &Generator{secretKey: []byte(secretKey)}
}

// Generate creates a QR code for a ticket
func (g *Generator) Generate(ticket *domain.Ticket) ([]byte, string, error) {
	payload := domain.QRPayload{
		Version:      1,
		TicketID:     ticket.ID,
		BookingID:    ticket.BookingID,
		PassengerNID: ticket.PassengerNID,
		SeatNumber:   ticket.SeatNumber,
		TripID:       ticket.TripID,
		Departure:    ticket.DepartureTime.Unix(),
	}

	// Sign the payload
	payload.Signature = g.sign(payload)

	// Encode to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal QR payload: %w", err)
	}

	// Base64 encode for QR
	qrData := base64.StdEncoding.EncodeToString(data)

	// Generate QR code PNG
	qr, err := qrcode.New(qrData, qrcode.Medium)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create QR code: %w", err)
	}

	pngData, err := qr.PNG(256)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PNG: %w", err)
	}

	return pngData, qrData, nil
}

// WritePNG writes QR code to a writer
func (g *Generator) WritePNG(w io.Writer, ticket *domain.Ticket, size int) error {
	payload := domain.QRPayload{
		Version:      1,
		TicketID:     ticket.ID,
		BookingID:    ticket.BookingID,
		PassengerNID: ticket.PassengerNID,
		SeatNumber:   ticket.SeatNumber,
		TripID:       ticket.TripID,
		Departure:    ticket.DepartureTime.Unix(),
	}
	payload.Signature = g.sign(payload)

	data, _ := json.Marshal(payload)
	qrData := base64.StdEncoding.EncodeToString(data)

	qr, err := qrcode.New(qrData, qrcode.Medium)
	if err != nil {
		return err
	}

	return png.Encode(w, qr.Image(size))
}

// Validate verifies a QR code and returns the payload
func (g *Generator) Validate(qrData string) (*domain.QRPayload, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(qrData)
	if err != nil {
		return nil, fmt.Errorf("invalid QR format: %w", err)
	}

	// Parse JSON
	var payload domain.QRPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid QR data: %w", err)
	}

	// Verify signature
	providedSig := payload.Signature
	payload.Signature = ""
	expectedSig := g.sign(payload)

	if providedSig != expectedSig {
		return nil, fmt.Errorf("invalid QR signature")
	}

	payload.Signature = providedSig
	return &payload, nil
}

// sign creates HMAC signature for payload
func (g *Generator) sign(payload domain.QRPayload) string {
	// Clear signature for signing
	payload.Signature = ""

	data, _ := json.Marshal(payload)
	h := hmac.New(sha256.New, g.secretKey)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

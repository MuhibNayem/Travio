package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/pdf"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/qr"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/repository"
)

type FulfillmentService struct {
	ticketRepo   *repository.TicketRepository
	qrGenerator  *qr.Generator
	pdfGenerator *pdf.Generator
}

func NewFulfillmentService(
	ticketRepo *repository.TicketRepository,
	qrGen *qr.Generator,
	pdfGen *pdf.Generator,
) *FulfillmentService {
	return &FulfillmentService{
		ticketRepo:   ticketRepo,
		qrGenerator:  qrGen,
		pdfGenerator: pdfGen,
	}
}

type GenerateTicketsReq struct {
	BookingID      string
	OrderID        string
	OrganizationID string
	TripID         string
	RouteName      string
	FromStation    string
	ToStation      string
	DepartureTime  time.Time
	ArrivalTime    time.Time
	Passengers     []PassengerSeat
	ContactEmail   string
	ContactPhone   string
}

type PassengerSeat struct {
	NID        string
	Name       string
	SeatID     string
	SeatNumber string
	SeatClass  string
	PricePaisa int64
}

type GenerateTicketsResp struct {
	Tickets []*domain.Ticket
	PDFURL  string
	PDFData []byte
}

func (s *FulfillmentService) GenerateTickets(ctx context.Context, req *GenerateTicketsReq) (*GenerateTicketsResp, error) {
	var tickets []*domain.Ticket
	qrPNGs := make(map[string][]byte)

	for _, p := range req.Passengers {
		ticket := &domain.Ticket{
			BookingID:      req.BookingID,
			OrderID:        req.OrderID,
			OrganizationID: req.OrganizationID,
			TripID:         req.TripID,
			RouteName:      req.RouteName,
			FromStation:    req.FromStation,
			ToStation:      req.ToStation,
			DepartureTime:  req.DepartureTime,
			ArrivalTime:    req.ArrivalTime,
			PassengerNID:   p.NID,
			PassengerName:  p.Name,
			SeatNumber:     p.SeatNumber,
			SeatClass:      p.SeatClass,
			PricePaisa:     p.PricePaisa,
			Currency:       "BDT",
			Status:         domain.TicketStatusActive,
			ValidUntil:     req.DepartureTime.Add(24 * time.Hour), // Valid until 24h after departure
		}

		tickets = append(tickets, ticket)
	}

	// Create tickets in DB first to get IDs
	if err := s.ticketRepo.CreateBatch(ctx, tickets); err != nil {
		return nil, fmt.Errorf("failed to create tickets: %w", err)
	}

	// Generate QR codes for each ticket
	for _, ticket := range tickets {
		pngData, qrData, err := s.qrGenerator.Generate(ticket)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR for ticket %s: %w", ticket.ID, err)
		}
		ticket.QRCodeData = qrData
		qrPNGs[ticket.ID] = pngData
	}

	// Generate combined PDF
	pdfData, err := s.pdfGenerator.GenerateMultiTicketPDF(tickets, qrPNGs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &GenerateTicketsResp{
		Tickets: tickets,
		PDFData: pdfData,
	}, nil
}

func (s *FulfillmentService) GetTicket(ctx context.Context, ticketID string) (*domain.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, ticketID)
}

func (s *FulfillmentService) ListTickets(ctx context.Context, orderID string) ([]*domain.Ticket, error) {
	return s.ticketRepo.ListByOrder(ctx, orderID)
}

func (s *FulfillmentService) ValidateTicket(ctx context.Context, qrData, validatorID string) (*domain.Ticket, error) {
	// Validate QR code
	payload, err := s.qrGenerator.Validate(qrData)
	if err != nil {
		return nil, fmt.Errorf("invalid QR: %w", err)
	}

	// Get ticket from DB
	ticket, err := s.ticketRepo.GetByID(ctx, payload.TicketID)
	if err != nil {
		return nil, err
	}

	// Check status
	if ticket.Status != domain.TicketStatusActive {
		return ticket, fmt.Errorf("ticket is %s", ticket.Status)
	}

	if ticket.IsBoarded {
		return ticket, fmt.Errorf("ticket already used at %s", ticket.BoardedAt.Format(time.RFC3339))
	}

	// Check validity
	if time.Now().After(ticket.ValidUntil) {
		return ticket, fmt.Errorf("ticket expired")
	}

	// Mark as boarded
	if err := s.ticketRepo.MarkAsBoarded(ctx, ticket.ID, validatorID); err != nil {
		return nil, err
	}

	ticket.IsBoarded = true
	ticket.BoardedAt = time.Now()
	ticket.BoardedBy = validatorID
	ticket.Status = domain.TicketStatusUsed

	return ticket, nil
}

func (s *FulfillmentService) CancelTicket(ctx context.Context, ticketID, reason string) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status != domain.TicketStatusActive {
		return nil, fmt.Errorf("cannot cancel ticket in status: %s", ticket.Status)
	}

	if err := s.ticketRepo.UpdateStatus(ctx, ticketID, domain.TicketStatusCancelled); err != nil {
		return nil, err
	}

	ticket.Status = domain.TicketStatusCancelled
	return ticket, nil
}

func (s *FulfillmentService) GetTicketPDF(ctx context.Context, ticketID string) ([]byte, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Generate QR for this ticket
	pngData, _, err := s.qrGenerator.Generate(ticket)
	if err != nil {
		return nil, err
	}

	return s.pdfGenerator.GenerateTicketPDF(ticket, pngData)
}

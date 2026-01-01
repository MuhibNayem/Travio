package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/domain"
	"github.com/jung-kurt/gofpdf"
)

// Generator creates PDF tickets
type Generator struct {
	companyName string
	companyLogo string
}

func NewGenerator(companyName, logoPath string) *Generator {
	return &Generator{
		companyName: companyName,
		companyLogo: logoPath,
	}
}

// GenerateTicketPDF creates a PDF for a single ticket
func (g *Generator) GenerateTicketPDF(ticket *domain.Ticket, qrPNG []byte) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 102, 204)
	pdf.CellFormat(0, 15, g.companyName, "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Ticket Title
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 10, "E-TICKET", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// QR Code (if provided)
	if len(qrPNG) > 0 {
		opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		pdf.RegisterImageOptionsReader("qr", opt, bytes.NewReader(qrPNG))
		pdf.ImageOptions("qr", 75, pdf.GetY(), 60, 60, false, opt, 0, "")
		pdf.Ln(65)
	}

	// Ticket ID
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 6, fmt.Sprintf("Ticket ID: %s", ticket.ID), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Journey Details Box
	pdf.SetFillColor(245, 245, 245)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Rect(15, pdf.GetY(), 180, 50, "FD")

	startY := pdf.GetY() + 5
	pdf.SetXY(20, startY)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(80, 8, "FROM")
	pdf.Cell(20, 8, "")
	pdf.Cell(80, 8, "TO")
	pdf.Ln(8)

	pdf.SetXY(20, startY+8)
	pdf.SetFont("Arial", "", 14)
	pdf.Cell(80, 10, ticket.FromStation)
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(20, 10, "→")
	pdf.SetFont("Arial", "", 14)
	pdf.Cell(80, 10, ticket.ToStation)
	pdf.Ln(15)

	pdf.SetXY(20, startY+25)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(80, 6, fmt.Sprintf("Departure: %s", ticket.DepartureTime.Format("02 Jan 2006, 03:04 PM")))
	pdf.Ln(15)

	// Passenger Details
	pdf.SetY(pdf.GetY() + 15)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 8, "PASSENGER DETAILS", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(50, 7, "Name:", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 7, ticket.PassengerName, "", 1, "L", false, 0, "")

	pdf.CellFormat(50, 7, "NID:", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 7, maskNID(ticket.PassengerNID), "", 1, "L", false, 0, "")

	pdf.CellFormat(50, 7, "Seat:", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("%s (%s)", ticket.SeatNumber, ticket.SeatClass), "", 1, "L", false, 0, "")

	pdf.CellFormat(50, 7, "Route:", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 7, ticket.RouteName, "", 1, "L", false, 0, "")

	// Price
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(50, 10, "FARE:", "", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(0, 10, fmt.Sprintf("৳%.2f", float64(ticket.PricePaisa)/100), "", 1, "L", false, 0, "")

	// Footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated on %s", time.Now().Format("02 Jan 2006, 03:04 PM")), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, "Please show this ticket at boarding. Valid only with matching NID.", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateMultiTicketPDF creates a single PDF for multiple tickets
func (g *Generator) GenerateMultiTicketPDF(tickets []*domain.Ticket, qrPNGs map[string][]byte) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)

	for i, ticket := range tickets {
		pdf.AddPage()

		// Header
		pdf.SetFont("Arial", "B", 20)
		pdf.SetTextColor(0, 102, 204)
		pdf.CellFormat(0, 12, g.companyName, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 6, fmt.Sprintf("Ticket %d of %d", i+1, len(tickets)), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		// QR
		if qr, ok := qrPNGs[ticket.ID]; ok && len(qr) > 0 {
			opt := gofpdf.ImageOptions{ImageType: "PNG"}
			pdf.RegisterImageOptionsReader(ticket.ID, opt, bytes.NewReader(qr))
			pdf.ImageOptions(ticket.ID, 75, pdf.GetY(), 50, 50, false, opt, 0, "")
			pdf.Ln(55)
		}

		// Details
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(40, 7, "Passenger:", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 7, ticket.PassengerName, "", 1, "L", false, 0, "")

		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(40, 7, "Seat:", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 7, fmt.Sprintf("%s (%s)", ticket.SeatNumber, ticket.SeatClass), "", 1, "L", false, 0, "")

		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(40, 7, "Journey:", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 7, fmt.Sprintf("%s → %s", ticket.FromStation, ticket.ToStation), "", 1, "L", false, 0, "")

		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(40, 7, "Departure:", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(0, 7, ticket.DepartureTime.Format("02 Jan 2006, 03:04 PM"), "", 1, "L", false, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func maskNID(nid string) string {
	if len(nid) <= 4 {
		return nid
	}
	return nid[:2] + "****" + nid[len(nid)-4:]
}

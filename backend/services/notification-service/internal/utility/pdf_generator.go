package utility

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// TicketPDFData represents data for a single ticket in PDF
type TicketPDFData struct {
	TicketID       string
	TicketNumber   string
	TierName       string
	Price          float64
	QRCodeBase64   string
	EventName      string
	EventLocation  string
	EventStartTime string
	OrderID        string
}

// GenerateTicketPDF generates a professional e-ticket PDF with QR code
func GenerateTicketPDF(ticket *TicketPDFData) ([]byte, error) {
	// Create new PDF - A4 portrait
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Colors
	primaryColor := gofpdf.RGBType{R: 102, G: 126, B: 234} // Purple
	grayColor := gofpdf.RGBType{R: 108, G: 117, B: 125}    // Gray

	// Header background
	pdf.SetFillColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.Rect(0, 0, 210, 40, "F")

	// Company name
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetY(15)
	pdf.CellFormat(0, 10, "EVENT TICKETING PLATFORM", "", 1, "C", false, 0, "")

	// E-Ticket title
	pdf.SetFont("Arial", "", 12)
	pdf.SetY(28)
	pdf.CellFormat(0, 8, "E-TICKET", "", 1, "C", false, 0, "")

	// Reset text color
	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(50)

	// Event details section
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.CellFormat(0, 10, "Event Details", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)

	// Event info box
	pdf.SetFillColor(248, 249, 250)
	pdf.SetDrawColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.SetLineWidth(0.5)

	y := pdf.GetY()
	pdf.Rect(15, y, 180, 35, "FD")

	pdf.SetY(y + 5)
	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 6, "Event:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, ticket.EventName)
	pdf.Ln(7)

	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 6, "Location:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, ticket.EventLocation)
	pdf.Ln(7)

	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 6, "Date & Time:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, ticket.EventStartTime)
	pdf.Ln(12)

	// Ticket details section
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.CellFormat(0, 10, "Ticket Information", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)

	// Ticket info
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 7, "Ticket Type:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 7, ticket.TierName)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 7, "Price:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 7, fmt.Sprintf("Rp %s", formatCurrency(ticket.Price)))
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 7, "Ticket ID:")
	pdf.SetFont("Arial", "", 12)
	pdf.SetFont("Courier", "", 10)
	pdf.Cell(0, 7, ticket.TicketID)
	pdf.Ln(12)

	// QR Code section
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.CellFormat(0, 10, "QR Code", "", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)

	// Decode base64 QR code
	qrData, err := decodeBase64Image(ticket.QRCodeBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QR code: %w", err)
	}

	// Save QR code to temporary file for PDF
	tmpFile := fmt.Sprintf("/tmp/qr_%s.png", ticket.TicketID)
	pdf.RegisterImageReader(tmpFile, "png", strings.NewReader(qrData))

	// Center QR code
	qrSize := 60.0
	pageWidth := 210.0
	qrX := (pageWidth - qrSize) / 2

	// Draw QR code border
	pdf.SetDrawColor(primaryColor.R, primaryColor.G, primaryColor.B)
	pdf.SetLineWidth(1)
	pdf.Rect(qrX-2, pdf.GetY()-2, qrSize+4, qrSize+4, "D")

	// Insert QR code image
	pdf.ImageOptions(tmpFile, qrX, pdf.GetY(), qrSize, qrSize, false, gofpdf.ImageOptions{ImageType: "png"}, 0, "")
	pdf.Ln(qrSize + 8)

	// Ticket number below QR
	pdf.SetFont("Courier", "", 10)
	pdf.SetTextColor(grayColor.R, grayColor.G, grayColor.B)
	pdf.CellFormat(0, 6, ticket.TicketNumber, "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Instructions section
	pdf.SetFillColor(255, 243, 205)
	pdf.SetDrawColor(255, 193, 7)
	pdf.SetLineWidth(0.5)

	y = pdf.GetY()
	pdf.Rect(15, y, 180, 40, "FD")

	pdf.SetY(y + 5)
	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(133, 100, 4)
	pdf.Cell(0, 6, "IMPORTANT INSTRUCTIONS")
	pdf.Ln(8)

	pdf.SetX(20)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(160, 5,
		"• Show this QR code at the entrance\n"+
			"• One-time use only - cannot be used after scanned\n"+
			"• Arrive at least 30 minutes before event starts\n"+
			"• This ticket is non-transferable and non-refundable",
		"", "L", false)

	pdf.Ln(5)

	// Footer
	pdf.SetY(270)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(grayColor.R, grayColor.G, grayColor.B)
	pdf.CellFormat(0, 5, "Order ID: "+ticket.OrderID, "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, "Generated on: "+time.Now().Format("2 Jan 2006 15:04 MST"), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, "Event Ticketing Platform - www.eventticket.com", "", 1, "C", false, 0, "")

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// decodeBase64Image decodes base64 image (with or without data URI prefix)
func decodeBase64Image(base64Str string) (string, error) {
	// Remove data URI prefix if exists
	if strings.HasPrefix(base64Str, "data:image") {
		parts := strings.Split(base64Str, ",")
		if len(parts) > 1 {
			base64Str = parts[1]
		}
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	return string(decoded), nil
}

// formatCurrency formats amount to Indonesian Rupiah format
func formatCurrency(amount float64) string {
	str := fmt.Sprintf("%.0f", amount)

	var result []rune
	count := 0

	for i := len(str) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = append([]rune{'.'}, result...)
		}
		result = append([]rune{rune(str[i])}, result...)
		count++
	}

	return string(result)
}

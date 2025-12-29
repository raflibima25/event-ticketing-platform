package utility

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a QR code as base64 encoded string
func GenerateQRCode(data string) (string, error) {
	// Generate QR code with medium error correction level
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert to PNG bytes (256x256 pixels)
	pngBytes, err := qr.PNG(256)
	if err != nil {
		return "", fmt.Errorf("failed to convert QR to PNG: %w", err)
	}

	// Encode to base64 for easy transmission
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	return encoded, nil
}

// GenerateTicketQRData creates the data string for ticket QR code
func GenerateTicketQRData(ticketID, eventID string) string {
	// Format: TICKET|{ticket_id}|{event_id}
	// This can be scanned and validated at event entrance
	return fmt.Sprintf("TICKET|%s|%s", ticketID, eventID)
}

// ParseTicketQRData parses QR data and extracts ticket ID and event ID
func ParseTicketQRData(qrData string) (ticketID, eventID string, err error) {
	// Expected format: TICKET|{ticket_id}|{event_id}
	parts := strings.Split(qrData, "|")

	if len(parts) != 3 {
		return "", "", errors.New("invalid QR data format")
	}

	if parts[0] != "TICKET" {
		return "", "", errors.New("invalid QR data prefix")
	}

	ticketID = parts[1]
	eventID = parts[2]

	// Basic validation - ensure they're not empty
	if ticketID == "" || eventID == "" {
		return "", "", errors.New("invalid ticket or event ID in QR data")
	}

	return ticketID, eventID, nil
}

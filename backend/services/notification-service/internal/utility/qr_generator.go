package utility

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCodeBase64 generates QR code as base64 string for embedding in HTML
func GenerateQRCodeBase64(data string, size int) (string, error) {
	// Generate QR code as PNG bytes
	qrCode, err := qrcode.Encode(data, qrcode.Medium, size)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert to base64 string
	base64Str := base64.StdEncoding.EncodeToString(qrCode)

	// Return data URI format for HTML img src
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// GenerateQRCodeBytes generates QR code as raw bytes
func GenerateQRCodeBytes(data string, size int) ([]byte, error) {
	qrCode, err := qrcode.Encode(data, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return qrCode, nil
}

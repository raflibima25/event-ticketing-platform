package utility

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// VerifyXenditSignature verifies Xendit webhook signature
// Xendit uses HMAC-SHA256 to sign webhook payloads
func VerifyXenditSignature(payload []byte, signature, webhookToken string) error {
	// Create HMAC with SHA256
	h := hmac.New(sha256.New, []byte(webhookToken))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures (constant time comparison to prevent timing attacks)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// VerifyCallbackToken verifies Xendit callback token from header
// This is the simpler verification method using x-callback-token header
func VerifyCallbackToken(receivedToken, expectedToken string) error {
	if receivedToken != expectedToken {
		return fmt.Errorf("invalid callback token")
	}
	return nil
}

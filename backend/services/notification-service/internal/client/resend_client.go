package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResendClient handles communication with Resend API
type ResendClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewResendClient creates new Resend client instance
func NewResendClient(apiKey string) *ResendClient {
	return &ResendClient{
		apiKey:  apiKey,
		baseURL: "https://api.resend.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EmailAttachment represents email attachment
type EmailAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // Base64 encoded
}

// EmailRequest represents request to send email
type EmailRequest struct {
	From        string             `json:"from"`
	To          string             `json:"to"`
	Subject     string             `json:"subject"`
	HTML        string             `json:"html"`
	Attachments []EmailAttachment  `json:"attachments,omitempty"`
}

// EmailResponse represents Resend API response
type EmailResponse struct {
	ID string `json:"id"`
}

// SendEmail sends email via Resend API
func (c *ResendClient) SendEmail(req *EmailRequest) (*EmailResponse, error) {
	url := fmt.Sprintf("%s/emails", c.baseURL)

	// Marshal request body
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("resend API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var emailResp EmailResponse
	if err := json.Unmarshal(body, &emailResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &emailResp, nil
}

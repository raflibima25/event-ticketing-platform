package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/response"
)

// XenditClient handles communication with Xendit API
type XenditClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewXenditClient creates new Xendit client instance
func NewXenditClient(cfg *config.XenditConfig) *XenditClient {
	return &XenditClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateInvoice creates a new invoice in Xendit
func (c *XenditClient) CreateInvoice(req *request.XenditCreateInvoiceRequest) (*response.XenditInvoiceResponse, error) {
	url := fmt.Sprintf("%s/v2/invoices", c.baseURL)

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
	httpReq.Header.Set("Authorization", c.getAuthHeader())

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
		return nil, fmt.Errorf("xendit API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var invoiceResp response.XenditInvoiceResponse
	if err := json.Unmarshal(body, &invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &invoiceResp, nil
}

// GetInvoice retrieves invoice by ID from Xendit
func (c *XenditClient) GetInvoice(invoiceID string) (*response.XenditInvoiceResponse, error) {
	url := fmt.Sprintf("%s/v2/invoices/%s", c.baseURL, invoiceID)

	// Create HTTP request
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", c.getAuthHeader())

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
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("invoice not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xendit API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var invoiceResp response.XenditInvoiceResponse
	if err := json.Unmarshal(body, &invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &invoiceResp, nil
}

// getAuthHeader returns Basic Auth header for Xendit API
func (c *XenditClient) getAuthHeader() string {
	// Xendit uses Basic Auth with API key as username and empty password
	auth := c.apiKey + ":"
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encoded
}

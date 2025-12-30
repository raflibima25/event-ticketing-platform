package request

// CreateInvoiceRequest represents request to create payment invoice
type CreateInvoiceRequest struct {
	OrderID       string  `json:"order_id" binding:"required,uuid"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
	PayerEmail    string  `json:"payer_email" binding:"required,email"`
	Description   string  `json:"description" binding:"required"`
	SuccessRedirectURL string `json:"success_redirect_url,omitempty"`
	FailureRedirectURL string `json:"failure_redirect_url,omitempty"`
}

// XenditCreateInvoiceRequest represents Xendit API create invoice request
type XenditCreateInvoiceRequest struct {
	ExternalID         string   `json:"external_id"`
	Amount             float64  `json:"amount"`
	PayerEmail         string   `json:"payer_email"`
	Description        string   `json:"description"`
	InvoiceDuration    int      `json:"invoice_duration"` // in seconds
	SuccessRedirectURL string   `json:"success_redirect_url,omitempty"`
	FailureRedirectURL string   `json:"failure_redirect_url,omitempty"`
	Currency           string   `json:"currency"`
	Items              []XenditInvoiceItem `json:"items,omitempty"`
}

// XenditInvoiceItem represents an item in Xendit invoice
type XenditInvoiceItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Category string  `json:"category,omitempty"`
}

// RefundRequest represents request to create refund
type RefundRequest struct {
	OrderID string `json:"order_id" binding:"required,uuid"`
	Reason  string `json:"reason"`
}

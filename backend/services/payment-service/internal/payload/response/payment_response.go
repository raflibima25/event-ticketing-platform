package response

import (
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
)

// InvoiceResponse represents invoice response to client
type InvoiceResponse struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	ExternalID    string     `json:"external_id"`
	InvoiceURL    string     `json:"invoice_url"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// XenditInvoiceResponse represents Xendit API invoice response
type XenditInvoiceResponse struct {
	ID                     string       `json:"id"`
	ExternalID             string       `json:"external_id"`
	UserID                 string       `json:"user_id"`
	Status                 string       `json:"status"`
	MerchantName           string       `json:"merchant_name"`
	Amount                 float64      `json:"amount"`
	PayerEmail             string       `json:"payer_email"`
	Description            string       `json:"description"`
	ExpiryDate             time.Time    `json:"expiry_date"`
	InvoiceURL             string       `json:"invoice_url"`
	AvailableBanks         []XenditBank `json:"available_banks"`
	AvailableRetailOutlets interface{}  `json:"available_retail_outlets"`
	AvailableEwallets      interface{}  `json:"available_ewallets"`
	ShouldExcludeCreditCard bool        `json:"should_exclude_credit_card"`
	ShouldSendEmail        bool         `json:"should_send_email"`
	Created                time.Time    `json:"created"`
	Updated                time.Time    `json:"updated"`
	Currency               string       `json:"currency"`
}

// XenditBank represents bank in Xendit response
type XenditBank struct {
	BankCode          string `json:"bank_code"`
	CollectionType    string `json:"collection_type"`
	BankBranch        string `json:"bank_branch"`
	AccountHolderName string `json:"account_holder_name"`
	TransferAmount    int    `json:"transfer_amount"`
}

// XenditWebhookPayload represents Xendit webhook payload
type XenditWebhookPayload struct {
	ID                string    `json:"id"`
	ExternalID        string    `json:"external_id"`
	UserID            string    `json:"user_id"`
	Status            string    `json:"status"`
	Amount            float64   `json:"amount"`
	PaidAmount        float64   `json:"paid_amount,omitempty"`
	PayerEmail        string    `json:"payer_email"`
	Description       string    `json:"description"`
	PaymentMethod     string    `json:"payment_method,omitempty"`
	PaymentChannel    string    `json:"payment_channel,omitempty"`
	PaymentDestination string   `json:"payment_destination,omitempty"`
	PaidAt            time.Time `json:"paid_at,omitempty"`
	Updated           time.Time `json:"updated"`
	Created           time.Time `json:"created"`
}

// ToInvoiceResponse converts PaymentTransaction entity to response
func ToInvoiceResponse(payment *entity.PaymentTransaction) *InvoiceResponse {
	invoiceURL := ""
	if payment.InvoiceURL != nil {
		invoiceURL = *payment.InvoiceURL
	}

	return &InvoiceResponse{
		ID:         payment.ID,
		OrderID:    payment.OrderID,
		ExternalID: payment.ExternalID,
		InvoiceURL: invoiceURL,
		Amount:     payment.Amount,
		Status:     payment.Status,
		ExpiresAt:  payment.ExpiresAt,
		CreatedAt:  payment.CreatedAt,
	}
}

// RefundResponse represents refund response
type RefundResponse struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToRefundResponse converts Refund entity to response
func ToRefundResponse(refund *entity.Refund) *RefundResponse {
	return &RefundResponse{
		ID:        refund.ID,
		OrderID:   refund.OrderID,
		Amount:    refund.Amount,
		Status:    refund.Status,
		CreatedAt: refund.CreatedAt,
	}
}

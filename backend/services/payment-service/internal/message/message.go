package message

// Success messages
const (
	MsgInvoiceCreated     = "Invoice created successfully"
	MsgInvoiceRetrieved   = "Invoice retrieved successfully"
	MsgWebhookProcessed   = "Webhook processed successfully"
	MsgRefundRequested    = "Refund requested successfully"
	MsgRefundCompleted    = "Refund completed successfully"
)

// Error messages
const (
	ErrInvalidRequest      = "Invalid request payload"
	ErrUnauthorized        = "Unauthorized access"
	ErrInternalServer      = "Internal server error"
	ErrPaymentNotFound     = "Payment transaction not found"
	ErrInvoiceNotFound     = "Invoice not found"
	ErrWebhookNotFound     = "Webhook event not found"
	ErrInvalidSignature    = "Invalid webhook signature"
	ErrDuplicateWebhook    = "Webhook already processed"
	ErrPaymentAlreadyPaid  = "Payment already completed"
	ErrPaymentExpired      = "Payment has expired"
	ErrRefundNotAllowed    = "Refund not allowed for this order"
	ErrXenditAPIError      = "Xendit API error"
)

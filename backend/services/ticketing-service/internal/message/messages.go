package message

// Success messages
const (
	MsgCartItemAdded      = "Item added to cart successfully"
	MsgCartItemUpdated    = "Cart item updated successfully"
	MsgCartItemRemoved    = "Cart item removed successfully"
	MsgCartCleared        = "Cart cleared successfully"
	MsgCartRetrieved      = "Cart retrieved successfully"
	MsgOrderCreated       = "Order created successfully"
	MsgOrderRetrieved     = "Order retrieved successfully"
	MsgOrdersRetrieved    = "Orders retrieved successfully"
	MsgOrderCancelled     = "Order cancelled successfully"
	MsgOrderConfirmed     = "Order confirmed successfully"
	MsgTicketRetrieved    = "Ticket retrieved successfully"
	MsgTicketsRetrieved   = "Tickets retrieved successfully"
	MsgTicketValidated    = "Ticket validated successfully"
	MsgAvailabilityChecked = "Availability checked successfully"
)

// Error messages
const (
	ErrInvalidRequest        = "Invalid request payload"
	ErrUnauthorized          = "Unauthorized access"
	ErrForbidden             = "You don't have permission to perform this action"
	ErrInternalServer        = "Internal server error"
	ErrCartNotFound          = "Cart not found"
	ErrCartItemNotFound      = "Cart item not found"
	ErrOrderNotFound         = "Order not found"
	ErrTicketNotFound        = "Ticket not found"
	ErrTicketTierNotFound    = "Ticket tier not found"
	ErrInsufficientQuota     = "Insufficient ticket quota available"
	ErrInvalidQuantity       = "Invalid quantity"
	ErrMaxPerOrderExceeded   = "Maximum tickets per order exceeded"
	ErrOrderExpired          = "Order has expired"
	ErrOrderAlreadyPaid      = "Order has already been paid"
	ErrOrderAlreadyCancelled = "Order has already been cancelled"
	ErrCannotCancelOrder     = "Cannot cancel order at this stage"
	ErrTicketAlreadyUsed     = "Ticket has already been used"
	ErrTicketInvalid         = "Ticket is invalid"
	ErrLockAcquisitionFailed = "Failed to acquire lock, please try again"
	ErrEventNotFound         = "Event not found"
)

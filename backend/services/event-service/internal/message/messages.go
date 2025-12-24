package message

// Success messages
const (
	MsgEventCreated      = "Event created successfully"
	MsgEventUpdated      = "Event updated successfully"
	MsgEventDeleted      = "Event deleted successfully"
	MsgEventRetrieved    = "Event retrieved successfully"
	MsgEventsRetrieved   = "Events retrieved successfully"
	MsgTicketTierCreated = "Ticket tier created successfully"
	MsgTicketTierUpdated = "Ticket tier updated successfully"
	MsgTicketTierDeleted = "Ticket tier deleted successfully"
)

// Error messages
const (
	ErrInvalidRequest           = "Invalid request payload"
	ErrEventNotFound            = "Event not found"
	ErrTicketTierNotFound       = "Ticket tier not found"
	ErrUnauthorized             = "Unauthorized access"
	ErrForbidden                = "You don't have permission to perform this action"
	ErrInternalServer           = "Internal server error"
	ErrInvalidDateRange         = "End date must be after start date"
	ErrEventSlugExists          = "Event with this slug already exists"
	ErrInvalidStatus            = "Invalid event status"
	ErrInvalidCategory          = "Invalid event category"
	ErrQuotaBelowSoldCount      = "Quota cannot be less than sold count"
	ErrInvalidEarlyBirdSettings = "Early bird end date must be set when early bird price is provided"
	ErrInvalidEarlyBirdPrice    = "Early bird price must be less than regular price"
	ErrInvalidEarlyBirdEndDate  = "Early bird end date must be in the future"
)

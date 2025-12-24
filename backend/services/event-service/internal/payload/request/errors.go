package request

import "errors"

var (
	// Ticket tier validation errors
	ErrInvalidEarlyBirdSettings = errors.New("early bird end date must be set when early bird price is provided")
	ErrInvalidEarlyBirdPrice    = errors.New("early bird price must be less than regular price")
	ErrInvalidEarlyBirdEndDate  = errors.New("early bird end date must be in the future")
)

package response

import "time"

// ============================================
// Success Response
// ============================================

type ApiResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(message string, data interface{}) ApiResponse {
	return ApiResponse{
		Status:  true,
		Message: message,
		Data:    data,
	}
}

// ============================================
// Paginated Response
// ============================================

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

type PaginatedResponse struct {
	Status  bool           `json:"status"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

func SuccessWithPagination(message string, data interface{}, meta PaginationMeta) PaginatedResponse {
	return PaginatedResponse{
		Status:  true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

// ============================================
// Error Response
// ============================================

type ErrorResponse struct {
	Status    bool        `json:"status"`
	Message   string      `json:"message"`
	Errors    interface{} `json:"errors,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func Error(message string, errors interface{}) ErrorResponse {
	return ErrorResponse{
		Status:    false,
		Message:   message,
		Errors:    errors,
		Timestamp: time.Now(),
	}
}

func ErrorWithCode(message string, errorCode string, errors interface{}) ErrorResponse {
	return ErrorResponse{
		Status:    false,
		Message:   message,
		ErrorCode: errorCode,
		Errors:    errors,
		Timestamp: time.Now(),
	}
}

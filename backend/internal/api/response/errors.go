package response

import (
	"net/http"
	"time"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	// General errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Domain-specific errors
	ErrCodeInvalidState       ErrorCode = "INVALID_STATE_TRANSITION"
	ErrCodeAlreadyAllocated   ErrorCode = "ALREADY_ALLOCATED"
	ErrCodeNotSubscribed      ErrorCode = "NOT_SUBSCRIBED_TO_INBOX"
	ErrCodeOperatorOffline    ErrorCode = "OPERATOR_OFFLINE"
	ErrCodeConversationLocked ErrorCode = "CONVERSATION_LOCKED"
	ErrCodeTenantRequired     ErrorCode = "TENANT_REQUIRED"
	ErrCodeOperatorRequired   ErrorCode = "OPERATOR_REQUIRED"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorBody `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorBody contains error details
type ErrorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details []string  `json:"details,omitempty"`
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, code ErrorCode, message string, details ...string) {
	response := ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}
	writeJSON(w, status, response)
}

// BadRequest sends a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string, details ...string) {
	Error(w, http.StatusBadRequest, ErrCodeBadRequest, message, details...)
}

// ValidationError sends a 400 error for validation failures
func ValidationError(w http.ResponseWriter, message string, details ...string) {
	Error(w, http.StatusBadRequest, ErrCodeValidation, message, details...)
}

// NotFound sends a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrCodeNotFound, message)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, ErrCodeForbidden, message)
}

// Conflict sends a 409 Conflict error
func Conflict(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusConflict, code, message)
}

// InternalError sends a 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, ErrCodeInternal, message)
}

// TooManyRequests sends a 429 Too Many Requests error
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, ErrCodeTooManyRequests, message)
}

// ServiceUnavailable sends a 503 Service Unavailable error
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, ErrCodeInternal, message)
}

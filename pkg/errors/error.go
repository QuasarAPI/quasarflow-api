package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents a category of application errors.
// It helps in identifying and handling different types of errors consistently.
type ErrorType string

const (
	// Domain error types
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR" // Represents validation failures
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"        // Represents resource not found errors
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"     // Represents authentication/authorization failures
	ErrorTypeConflict     ErrorType = "CONFLICT"         // Represents resource conflict errors

	// Technical error types
	ErrorTypeInternal   ErrorType = "INTERNAL_ERROR"         // Represents unexpected internal errors
	ErrorTypeDatabase   ErrorType = "DATABASE_ERROR"         // Represents database operation failures
	ErrorTypeExternal   ErrorType = "EXTERNAL_SERVICE_ERROR" // Represents external service failures
	ErrorTypeCrypto     ErrorType = "CRYPTO_ERROR"           // Represents cryptographic operation failures
	ErrorTypeBlockchain ErrorType = "BLOCKCHAIN_ERROR"       // Represents blockchain interaction failures
)

// AppError represents a structured error type that includes categorization,
// HTTP status codes, and additional context for API responses.
type AppError struct {
	Type       ErrorType // Type categorizes the error
	Message    string    // Message provides a human-readable error description
	Detail     string    // Detail provides additional context about the error
	StatusCode int       // StatusCode maps to HTTP status codes
	Err        error     // Err contains the underlying error, if any
}

// Error implements the error interface.
// It returns a formatted string containing the error type and message,
// with optional detail if present.
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Wrapper interface.
// It returns the underlying error if one exists.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error with the specified message and detail.
// It sets the appropriate HTTP status code to 400 Bad Request.
func NewValidationError(message string, detail string) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		Detail:     detail,
		StatusCode: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a new not found error with the specified message.
// It sets the appropriate HTTP status code to 404 Not Found.
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewInternalError creates a new internal error with the specified message and underlying error.
// It sets the appropriate HTTP status code to 500 Internal Server Error.
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewCryptoError creates a new cryptographic error with the specified message and underlying error.
// It sets the appropriate HTTP status code to 500 Internal Server Error.
func NewCryptoError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeCrypto,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewBlockchainError creates a new blockchain-related error with the specified message and underlying error.
// It sets the appropriate HTTP status code to 502 Bad Gateway.
func NewBlockchainError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeBlockchain,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Err:        err,
	}
}

// IsErrorType checks if the provided error is of the specified ErrorType.
func IsErrorType(err error, errorType ErrorType) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		return false
	}
	return appErr.Type == errorType
}

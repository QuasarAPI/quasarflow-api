package response

import (
	"encoding/json"
	"net/http"

	"quasarflow-api/pkg/errors"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides structured error information in API responses
type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Success writes a successful JSON response
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, try to send a plain error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Type:    "ERROR",
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, try to send a plain error
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// AppError writes an AppError as a JSON response
// It extracts status code and detailed information from the AppError
func AppError(w http.ResponseWriter, err *errors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Type:    string(err.Type),
			Message: err.Message,
			Detail:  err.Detail,
		},
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

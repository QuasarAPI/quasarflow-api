package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"quasarflow-api/internal/interface/http/response"
	"quasarflow-api/pkg/logger"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Context key type to avoid collisions
type contextKey string

const (
	// Context keys
	ctxKeyValidatedData contextKey = "validated_data"

	// Stellar validation constants
	stellarAddressLength = 56
	stellarAddressPrefix = "G"
	stellarSeedPrefix    = "S"
	base32Charset        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	// Content type
	contentTypeJSON = "application/json"

	// Validation tags
	tagRequired       = "required"
	tagMin            = "min"
	tagMax            = "max"
	tagEmail          = "email"
	tagOneOf          = "oneof"
	tagStellarAddress = "stellar_address"
	tagStellarSeed    = "stellar_seed"
	tagSafeString     = "safe_string"
)

var (
	// Dangerous patterns for XSS prevention
	dangerousPatterns = []string{
		"<script", "</script>",
		"javascript:",
		"data:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"<iframe",
		"<object",
		"<embed",
	}

	// Event handlers to sanitize
	eventHandlers = []string{
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"onfocus=", "onblur=", "onchange=", "onsubmit=",
	}

	// Common validation errors
	errInvalidContentType = errors.New("content-Type must be application/json")
	errInvalidJSON        = errors.New("invalid JSON format")
)

// ValidationMiddleware handles input validation for HTTP requests.
type ValidationMiddleware struct {
	validator *validator.Validate
	logger    logger.Logger
}

// NewValidationMiddleware creates a new validation middleware instance.
func NewValidationMiddleware(log logger.Logger) *ValidationMiddleware {
	v := validator.New()

	// Register custom validators
	_ = v.RegisterValidation(tagStellarAddress, validateStellarAddress)
	_ = v.RegisterValidation(tagStellarSeed, validateStellarSeed)
	_ = v.RegisterValidation(tagSafeString, validateSafeString)

	return &ValidationMiddleware{
		validator: v,
		logger:    log,
	}
}

// ValidateJSON validates JSON request body against the provided struct type.
// The validated data is stored in the request context under the key "validated_data".
func (vm *ValidationMiddleware) ValidateJSON(target interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := vm.validateRequest(w, r, target); err != nil {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// validateRequest performs the actual validation logic.
func (vm *ValidationMiddleware) validateRequest(w http.ResponseWriter, r *http.Request, target interface{}) error {
	// Validate content type
	if err := vm.checkContentType(r); err != nil {
		vm.logger.Warn("invalid content type",
			zap.String("content_type", r.Header.Get("Content-Type")),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, err.Error())
		return err
	}

	// Create instance and decode JSON
	instance, err := vm.decodeJSON(r.Body, target)
	if err != nil {
		vm.logger.Warn("failed to decode JSON",
			zap.Error(err),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, errInvalidJSON.Error())
		return err
	}

	// Validate struct
	if err := vm.validator.Struct(instance); err != nil {
		errMsg := vm.formatValidationErrors(err)
		vm.logger.Warn("validation failed",
			zap.String("errors", errMsg),
			zap.String("ip", r.RemoteAddr))
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Validation failed: %s", errMsg))
		return err
	}

	// Store validated data in context
	ctx := context.WithValue(r.Context(), ctxKeyValidatedData, instance)
	*r = *r.WithContext(ctx)

	return nil
}

// checkContentType validates the Content-Type header.
func (vm *ValidationMiddleware) checkContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, contentTypeJSON) {
		return errInvalidContentType
	}
	return nil
}

// decodeJSON decodes JSON from reader into a new instance of target type.
func (vm *ValidationMiddleware) decodeJSON(body io.Reader, target interface{}) (interface{}, error) {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	instance := reflect.New(targetType).Interface()

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields() // Reject unknown fields for stricter validation

	if err := decoder.Decode(instance); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return instance, nil
}

// formatValidationErrors converts validator errors into user-friendly messages.
func (vm *ValidationMiddleware) formatValidationErrors(err error) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err.Error()
	}

	messages := make([]string, 0, len(validationErrs))
	for _, e := range validationErrs {
		messages = append(messages, vm.formatSingleError(e))
	}

	return strings.Join(messages, "; ")
}

// formatSingleError formats a single validation error.
func (vm *ValidationMiddleware) formatSingleError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case tagRequired:
		return fmt.Sprintf("%s is required", field)
	case tagMin:
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case tagMax:
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case tagEmail:
		return fmt.Sprintf("%s must be a valid email address", field)
	case tagOneOf:
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case tagStellarAddress:
		return fmt.Sprintf("%s must be a valid Stellar address", field)
	case tagStellarSeed:
		return fmt.Sprintf("%s must be a valid Stellar seed", field)
	case tagSafeString:
		return fmt.Sprintf("%s contains invalid characters", field)
	default:
		return fmt.Sprintf("%s is invalid (tag: %s)", field, e.Tag())
	}
}

// validateStellarAddress validates Stellar public key format.
// Format: 56 characters, starts with 'G', Base32 encoded.
func validateStellarAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	return isValidStellarKey(address, stellarAddressPrefix)
}

// validateStellarSeed validates Stellar seed format.
// Format: 56 characters, starts with 'S', Base32 encoded.
func validateStellarSeed(fl validator.FieldLevel) bool {
	seed := fl.Field().String()
	return isValidStellarKey(seed, stellarSeedPrefix)
}

// isValidStellarKey validates the format of a Stellar key (address or seed).
func isValidStellarKey(key, prefix string) bool {
	if len(key) != stellarAddressLength || !strings.HasPrefix(key, prefix) {
		return false
	}

	return isBase32(key)
}

// isBase32 checks if a string contains only valid Base32 characters.
func isBase32(s string) bool {
	for _, char := range s {
		if !strings.ContainsRune(base32Charset, char) {
			return false
		}
	}
	return true
}

// validateSafeString validates that a string doesn't contain dangerous patterns.
func validateSafeString(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	for _, pattern := range dangerousPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return true
}

// SanitizeString removes potentially dangerous characters and patterns from input.
// This provides defense-in-depth against XSS attacks.
func SanitizeString(input string) string {
	// HTML entity encoding for angle brackets
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")

	// Remove dangerous protocols
	protocols := []string{"javascript:", "data:", "vbscript:"}
	for _, protocol := range protocols {
		input = strings.ReplaceAll(input, protocol, "")
	}

	// Remove event handlers
	for _, handler := range eventHandlers {
		input = strings.ReplaceAll(input, handler, "")
	}

	return strings.TrimSpace(input)
}

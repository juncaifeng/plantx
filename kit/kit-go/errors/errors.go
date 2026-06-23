// Package errors defines standard error types and codes for PlantX services.
package errors

import "fmt"

// KitError is the standard error type for PlantX services.
type KitError struct {
	Code    string
	Message string
	Cause   error
}

func (e *KitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new KitError.
func New(code, message string) *KitError {
	return &KitError{Code: code, Message: message}
}

// Wrap wraps an existing error.
func Wrap(code, message string, cause error) *KitError {
	return &KitError{Code: code, Message: message, Cause: cause}
}

// Common error codes.
const (
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeInternal     = "INTERNAL_ERROR"
	CodeInvalidInput = "INVALID_INPUT"
)

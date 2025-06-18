// Package validation provides command structure and metadata validation utilities.
package validation

import "fmt"

// Error represents a validation error with context
type Error struct {
	Type    string
	Details string
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Type, e.Details)
}

// NewValidationError creates a new validation error
func NewValidationError(errType, details string) *Error {
	return &Error{
		Type:    errType,
		Details: details,
	}
}

// IsValidationError checks if an error is a validation.Error
func IsValidationError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

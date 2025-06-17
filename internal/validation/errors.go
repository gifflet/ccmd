package validation

import "fmt"

// ValidationError represents a validation error with context
type ValidationError struct {
	Type    string
	Details string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Type, e.Details)
}

// NewValidationError creates a new validation error
func NewValidationError(errType, details string) *ValidationError {
	return &ValidationError{
		Type:    errType,
		Details: details,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

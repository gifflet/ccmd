package output

import (
	"fmt"
)

// UserError represents an error that should be displayed to the user
// with a friendly message
type UserError struct {
	Message string
	Details string
	Err     error
}

func (e *UserError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a new user-friendly error
func NewUserError(message string, err error) *UserError {
	return &UserError{
		Message: message,
		Err:     err,
	}
}

// NewUserErrorf creates a new user-friendly error with formatted message
func NewUserErrorf(format string, a ...interface{}) *UserError {
	return &UserError{
		Message: fmt.Sprintf(format, a...),
	}
}

// PrintUserError prints a UserError in a friendly format
func PrintUserError(err error) {
	if err == nil {
		return
	}

	if ue, ok := err.(*UserError); ok {
		PrintErrorf(ue.Message)
		if ue.Details != "" {
			PrintErrorf("Details: %s", ue.Details)
		}
		if ue.Err != nil {
			Debugf("Underlying error: %v", ue.Err)
		}
	} else {
		PrintErrorf("Error: %v", err)
	}
}

// WrapError wraps an error with a user-friendly message
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return &UserError{
		Message: message,
		Err:     err,
	}
}

// IsUserError checks if an error is a UserError
func IsUserError(err error) bool {
	_, ok := err.(*UserError)
	return ok
}

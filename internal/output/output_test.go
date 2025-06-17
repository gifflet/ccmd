package output

import (
	"errors"
	"testing"
)

func TestUserError(t *testing.T) {
	tests := []struct {
		name     string
		err      *UserError
		expected string
	}{
		{
			name: "simple message",
			err: &UserError{
				Message: "Operation failed",
			},
			expected: "Operation failed",
		},
		{
			name: "message with details",
			err: &UserError{
				Message: "Operation failed",
				Details: "file not found",
			},
			expected: "Operation failed: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("UserError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewUserError(t *testing.T) {
	originalErr := errors.New("original error")
	userErr := NewUserError("User friendly message", originalErr)

	if userErr.Message != "User friendly message" {
		t.Errorf("Expected message 'User friendly message', got '%s'", userErr.Message)
	}

	if userErr.Err != originalErr {
		t.Errorf("Expected wrapped error to be the original error")
	}
}

func TestNewUserErrorf(t *testing.T) {
	userErr := NewUserErrorf("Failed to process %d items", 5)

	expected := "Failed to process 5 items"
	if userErr.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, userErr.Message)
	}
}

func TestWrapError(t *testing.T) {
	// Test with nil error
	if err := WrapError(nil, "message"); err != nil {
		t.Errorf("WrapError with nil should return nil")
	}

	// Test with actual error
	originalErr := errors.New("original error")
	wrapped := WrapError(originalErr, "friendly message")

	userErr, ok := wrapped.(*UserError)
	if !ok {
		t.Errorf("Expected UserError type")
	}

	if userErr.Message != "friendly message" {
		t.Errorf("Expected message 'friendly message', got '%s'", userErr.Message)
	}
}

func TestIsUserError(t *testing.T) {
	// Test with UserError
	userErr := NewUserError("test", nil)
	if !IsUserError(userErr) {
		t.Errorf("Expected IsUserError to return true for UserError")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	if IsUserError(regularErr) {
		t.Errorf("Expected IsUserError to return false for regular error")
	}
}

package validation

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test-type", "test details")

	if err.Type != "test-type" {
		t.Errorf("expected Type = 'test-type', got %s", err.Type)
	}

	if err.Details != "test details" {
		t.Errorf("expected Details = 'test details', got %s", err.Details)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Type:    "missing-file",
		Details: "ccmd.yaml not found",
	}

	expected := "validation error [missing-file]: ccmd.yaml not found"
	if err.Error() != expected {
		t.Errorf("expected error string = %s, got %s", expected, err.Error())
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error",
			err:  NewValidationError("test", "details"),
			want: true,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

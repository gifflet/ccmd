// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeNotFound, "resource not found")

	if err.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, err.Code)
	}

	if err.Message != "resource not found" {
		t.Errorf("expected message 'resource not found', got '%s'", err.Message)
	}

	expected := "[NOT_FOUND] resource not found"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestNewf(t *testing.T) {
	err := Newf(CodeInvalidArgument, "invalid %s: %d", "count", 42)

	expected := "invalid count: 42"
	if err.Message != expected {
		t.Errorf("expected message '%s', got '%s'", expected, err.Message)
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, CodeInternal, "operation failed")

	if wrapped.Code != CodeInternal {
		t.Errorf("expected code %s, got %s", CodeInternal, wrapped.Code)
	}

	if wrapped.Cause != original {
		t.Errorf("expected cause to be original error")
	}

	if !errors.Is(wrapped, wrapped) {
		t.Error("wrapped error should match itself with errors.Is")
	}
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil, CodeInternal, "operation failed")
	if wrapped != nil {
		t.Error("wrapping nil error should return nil")
	}
}

func TestWrapf(t *testing.T) {
	original := errors.New("disk full")
	wrapped := Wrapf(original, CodeFileIO, "failed to write %s", "config.yaml")

	expected := "failed to write config.yaml"
	if wrapped.Message != expected {
		t.Errorf("expected message '%s', got '%s'", expected, wrapped.Message)
	}
}

func TestWithDetail(t *testing.T) {
	err := New(CodeNotFound, "command not found").
		WithDetail("command", "mycmd").
		WithDetail("version", "1.0.0")

	if err.Details["command"] != "mycmd" {
		t.Errorf("expected detail 'command' to be 'mycmd', got '%v'", err.Details["command"])
	}

	if err.Details["version"] != "1.0.0" {
		t.Errorf("expected detail 'version' to be '1.0.0', got '%v'", err.Details["version"])
	}
}

func TestWithDetails(t *testing.T) {
	err := New(CodeConfigInvalid, "invalid config").
		WithDetails(map[string]interface{}{
			"file":   "ccmd.yaml",
			"line":   42,
			"column": 10,
		})

	if err.Details["file"] != "ccmd.yaml" {
		t.Errorf("expected detail 'file' to be 'ccmd.yaml', got '%v'", err.Details["file"])
	}

	if err.Details["line"] != 42 {
		t.Errorf("expected detail 'line' to be 42, got '%v'", err.Details["line"])
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "ccmd error",
			err:      New(CodeGitClone, "clone failed"),
			expected: CodeGitClone,
		},
		{
			name:     "wrapped ccmd error",
			err:      fmt.Errorf("outer: %w", New(CodeCommandNotFound, "not found")),
			expected: CodeCommandNotFound,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: CodeUnknown,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: CodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetCode(tt.err)
			if code != tt.expected {
				t.Errorf("expected code %s, got %s", tt.expected, code)
			}
		})
	}
}

func TestIsCode(t *testing.T) {
	err := New(CodeFileNotFound, "file not found")

	if !IsCode(err, CodeFileNotFound) {
		t.Error("expected IsCode to return true for matching code")
	}

	if IsCode(err, CodeFileExists) {
		t.Error("expected IsCode to return false for non-matching code")
	}
}

func TestCategoryCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with CodeNotFound",
			err:      New(CodeNotFound, "not found"),
			checker:  IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with CodeCommandNotFound",
			err:      New(CodeCommandNotFound, "command not found"),
			checker:  IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with other code",
			err:      New(CodeInternal, "internal error"),
			checker:  IsNotFound,
			expected: false,
		},
		{
			name:     "IsAlreadyExists with CodeAlreadyExists",
			err:      New(CodeAlreadyExists, "already exists"),
			checker:  IsAlreadyExists,
			expected: true,
		},
		{
			name:     "IsPermissionDenied with CodeFilePermission",
			err:      New(CodeFilePermission, "permission denied"),
			checker:  IsPermissionDenied,
			expected: true,
		},
		{
			name:     "IsValidationError with CodeValidationFailed",
			err:      New(CodeValidationFailed, "validation failed"),
			checker:  IsValidationError,
			expected: true,
		},
		{
			name:     "IsGitError with CodeGitClone",
			err:      New(CodeGitClone, "clone failed"),
			checker:  IsGitError,
			expected: true,
		},
		{
			name:     "IsNetworkError with CodeNetworkTimeout",
			err:      New(CodeNetworkTimeout, "timeout"),
			checker:  IsNetworkError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	err1 := New(CodeNotFound, "not found")
	err2 := New(CodeNotFound, "also not found")
	err3 := New(CodeInternal, "internal")

	if !err1.Is(err2) {
		t.Error("errors with same code should match")
	}

	if err1.Is(err3) {
		t.Error("errors with different codes should not match")
	}

	if err1.Is(errors.New("standard error")) {
		t.Error("ccmd error should not match standard error")
	}
}

func TestUnwrap(t *testing.T) {
	original := errors.New("original")
	wrapped := Wrap(original, CodeInternal, "wrapped")

	unwrapped := wrapped.Unwrap()
	if unwrapped != original {
		t.Error("Unwrap should return the original error")
	}

	// Test with errors.Unwrap
	if errors.Unwrap(wrapped) != original {
		t.Error("errors.Unwrap should also work")
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package integration_test

import (
	"testing"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

func TestErrorHandlingIntegration(t *testing.T) {
	// Create logger (output goes to stderr by default)
	log := logger.New()

	// Create error handler with logger
	handler := errors.NewHandler(log)

	// Test different error types
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "not found error",
			err: errors.New(errors.CodeCommandNotFound, "command not found").
				WithDetail("command", "test-cmd"),
		},
		{
			name: "git clone error",
			err: errors.Wrap(
				errors.New(errors.CodeGitClone, "clone failed"),
				errors.CodeGitClone,
				"failed to install command",
			).WithDetail("repository", "https://github.com/user/repo"),
		},
		{
			name: "validation error",
			err: errors.New(errors.CodeValidationFailed, "invalid configuration").
				WithDetail("file", "ccmd.yaml").
				WithDetail("line", 42),
		},
		{
			name: "wrapped error",
			err: errors.Wrap(
				errors.New(errors.CodePermissionDenied, "access denied"),
				errors.CodeInternal,
				"operation failed",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that handling doesn't panic
			// Since we can't easily capture slog output, we just ensure it works
			handler.Handle(tt.err)
		})
	}
}

func TestLoggingWithContext(t *testing.T) {
	// Create logger
	log := logger.New()

	// Create a command-specific logger
	cmdLog := log.WithField("command", "install")

	// Log with additional context - just ensure it doesn't panic
	cmdLog.WithFields(logger.Fields{
		"repository": "github.com/user/repo",
		"version":    "v1.0.0",
	}).Info("installing command")

	// Test chaining
	cmdLog.
		WithField("package", "test-package").
		WithError(errors.New(errors.CodeNotFound, "package not found")).
		Error("installation failed")
}

func TestErrorIntegrationWithOutput(t *testing.T) {
	// Initialize error handler with output functions
	output.InitializeErrorHandler()

	// Create and handle an error
	err := errors.New(errors.CodeCommandNotFound, "command 'foo' not found").
		WithDetail("command", "foo")

	// This would normally print to stderr, but we can't easily capture that in tests
	// The important part is that it doesn't panic and integrates properly
	errors.Handle(err)
}

func TestLoggerIntegrationWithErrors(t *testing.T) {
	log := logger.New()

	// Test various error logging scenarios
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "log with error field",
			fn: func() {
				err := errors.New(errors.CodeNotFound, "not found")
				log.WithError(err).Error("operation failed")
			},
		},
		{
			name: "log with multiple fields and error",
			fn: func() {
				err := errors.New(errors.CodeValidationFailed, "validation failed")
				log.WithFields(logger.Fields{
					"file": "config.yaml",
					"line": 10,
				}).WithError(err).Error("config validation failed")
			},
		},
		{
			name: "contextual logger with error",
			fn: func() {
				contextLog := log.WithField("component", "installer")
				err := errors.New(errors.CodeGitClone, "clone failed")
				contextLog.WithError(err).Error("installation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure functions don't panic
			tt.fn()
		})
	}
}

func TestOutputMigrationIntegration(t *testing.T) {
	// Test the output migration functions with logger
	tests := []struct {
		name  string
		level string
		msg   string
	}{
		{"success", "success", "Operation completed"},
		{"error", "error", "Operation failed"},
		{"warning", "warning", "Operation has warnings"},
		{"info", "info", "Operation info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This logs to both logger and output
			output.LogAndPrintf(tt.level, "%s: %s", tt.name, tt.msg)
		})
	}
}

func TestDebugErrorIntegration(t *testing.T) {
	// Test debug error logging
	err := errors.New(errors.CodeInternal, "internal error")
	output.DebugError(err, "processing request")

	// Test with nil error
	output.DebugError(nil, "no error")
}

func TestErrorToOutputIntegration(t *testing.T) {
	// Test various error types with output conversion
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "not found error",
			err:  errors.New(errors.CodeNotFound, "resource not found"),
		},
		{
			name: "already exists error",
			err:  errors.New(errors.CodeAlreadyExists, "resource already exists"),
		},
		{
			name: "validation error",
			err:  errors.New(errors.CodeValidationFailed, "validation failed"),
		},
		{
			name: "permission denied error",
			err:  errors.New(errors.CodePermissionDenied, "permission denied"),
		},
		{
			name: "standard error",
			err:  errors.New(errors.CodeInternal, "internal error"),
		},
		{
			name: "nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would normally print to stdout/stderr
			// We're just ensuring it doesn't panic
			output.ErrorToOutput(tt.err)
		})
	}
}

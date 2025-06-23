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
	"bytes"
	"testing"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

func TestErrorHandlingIntegration(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf bytes.Buffer
	log := logger.New(&logBuf, logger.DebugLevel)

	// Create error handler with custom logger
	handler := errors.NewHandler(log)

	// Test different error types
	tests := []struct {
		name        string
		err         error
		expectLog   string
		expectLevel string
	}{
		{
			name: "not found error",
			err: errors.New(errors.CodeCommandNotFound, "command not found").
				WithDetail("command", "test-cmd"),
			expectLog:   "command not found",
			expectLevel: "[WARN]",
		},
		{
			name: "git clone error",
			err: errors.Wrap(
				errors.New(errors.CodeGitClone, "clone failed"),
				errors.CodeGitClone,
				"failed to install command",
			).WithDetail("repository", "https://github.com/user/repo"),
			expectLog:   "failed to install command",
			expectLevel: "[ERROR]",
		},
		{
			name: "validation error",
			err: errors.New(errors.CodeValidationFailed, "invalid configuration").
				WithDetail("file", "ccmd.yaml").
				WithDetail("line", 42),
			expectLog:   "invalid configuration",
			expectLevel: "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			logBuf.Reset()

			// Handle error
			handled := handler.Handle(tt.err)
			if !handled {
				t.Error("expected error to be handled")
			}

			// Check log output
			logOutput := logBuf.String()
			if !contains(logOutput, tt.expectLog) {
				t.Errorf("expected log to contain %q, got %q", tt.expectLog, logOutput)
			}
			if !contains(logOutput, tt.expectLevel) {
				t.Errorf("expected log level %q, got %q", tt.expectLevel, logOutput)
			}
		})
	}
}

func TestLoggingWithContext(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.InfoLevel)

	// Create a command-specific logger
	cmdLog := log.WithField("command", "install")

	// Log with additional context
	cmdLog.WithFields(logger.Fields{
		"repository": "github.com/user/repo",
		"version":    "v1.0.0",
	}).Info("installing command")

	output := buf.String()
	if !contains(output, "command=install") {
		t.Error("expected command field in output")
	}
	if !contains(output, "repository=github.com/user/repo") {
		t.Error("expected repository field in output")
	}
	if !contains(output, "version=v1.0.0") {
		t.Error("expected version field in output")
	}
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

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

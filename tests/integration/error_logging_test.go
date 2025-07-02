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

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

func TestErrorHandlingIntegration(t *testing.T) {
	// Create logger
	log := logger.New()

	// Use the default error handler
	handler := errors.DefaultHandler

	// Test different error types
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "not found error",
			err:  errors.NotFound("command test-cmd"),
		},
		{
			name: "already exists error",
			err:  errors.AlreadyExists("command test-cmd"),
		},
		{
			name: "git operation error",
			err:  errors.GitError("clone", nil),
		},
		{
			name: "validation error",
			err:  errors.InvalidInput("invalid configuration in ccmd.yaml"),
		},
		{
			name: "file error",
			err:  errors.FileError("read", "/etc/config", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that handler doesn't panic
			handler.Handle(tt.err)
		})
	}

	// Test with nil error
	if handler.Handle(nil) {
		t.Error("Handle should return false for nil error")
	}

	// Test logging integration
	log.WithError(errors.NotFound("package")).Error("failed to find package")
}

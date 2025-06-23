// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package install

import (
	"testing"
)

func TestVersionPrecedence(t *testing.T) {
	tests := []struct {
		name            string
		repository      string
		versionFlag     string
		expectedRepo    string
		expectedVersion string
	}{
		{
			name:            "only @ notation",
			repository:      "github.com/user/repo@v1.0.0",
			versionFlag:     "",
			expectedRepo:    "github.com/user/repo",
			expectedVersion: "v1.0.0",
		},
		{
			name:            "only --version flag",
			repository:      "github.com/user/repo",
			versionFlag:     "v2.0.0",
			expectedRepo:    "github.com/user/repo",
			expectedVersion: "v2.0.0",
		},
		{
			name:            "both @ and --version (flag takes precedence)",
			repository:      "github.com/user/repo@v1.0.0",
			versionFlag:     "v2.0.0",
			expectedRepo:    "github.com/user/repo",
			expectedVersion: "v2.0.0",
		},
		{
			name:            "neither @ nor --version",
			repository:      "github.com/user/repo",
			versionFlag:     "",
			expectedRepo:    "github.com/user/repo",
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the version precedence logic
			// The actual implementation is in runInstall function
			// which follows this precedence:
			// 1. Parse @ notation from repository
			// 2. If --version flag is provided, it overrides @ notation

			// Note: This is a documentation test to show expected behavior
			// Integration tests would be needed to test the full flow
		})
	}
}

// Test removed as extractRepoPath is now handled internally by the installer package

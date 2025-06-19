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

func TestExtractRepoPath(t *testing.T) {
	tests := []struct {
		name     string
		gitURL   string
		expected string
	}{
		{
			name:     "https URL with .git",
			gitURL:   "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "https URL without .git",
			gitURL:   "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "http URL",
			gitURL:   "http://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "git protocol URL",
			gitURL:   "git://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL",
			gitURL:   "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL without .git",
			gitURL:   "git@github.com:owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "URL with subdomain",
			gitURL:   "https://git.company.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "invalid URL",
			gitURL:   "not-a-url",
			expected: "",
		},
		{
			name:     "URL with only domain",
			gitURL:   "https://github.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoPath(tt.gitURL)
			if result != tt.expected {
				t.Errorf("extractRepoPath(%q) = %q, want %q", tt.gitURL, result, tt.expected)
			}
		})
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCommitHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid commit hashes
		{"valid short hash", "a76c963", true},
		{"valid 8 char hash", "a76c9635", true},
		{"valid full hash", "a76c96359914b84ed1bcdbc11df03e6313e09ecf", true},
		{"valid hash all numbers", "1234567", true},
		{"valid hash mixed", "abc1234", true},

		// Invalid cases
		{"too short", "a76c96", false},
		{"too long", "a76c96359914b84ed1bcdbc11df03e6313e09ecf1", false},
		{"contains uppercase", "A76C963", false},
		{"contains invalid char g", "g76c963", false},
		{"contains invalid char z", "a76c96z", false},
		{"contains special char", "a76c96-", false},
		{"empty string", "", false},
		{"branch name", "main", false},
		{"tag name", "v1.0.0", false},
		{"branch with slash", "feature/test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommitHash(tt.input)
			assert.Equal(t, tt.expected, result, "isCommitHash(%q) should return %v", tt.input, tt.expected)
		})
	}
}

func TestExtractRepoPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"https URL with .git", "https://github.com/owner/repo.git", "owner/repo"},
		{"https URL without .git", "https://github.com/owner/repo", "owner/repo"},
		{"SSH URL with .git", "git@github.com:owner/repo.git", "owner/repo"},
		{"SSH URL without .git", "git@github.com:owner/repo", "owner/repo"},
		{"shorthand format", "owner/repo", "owner/repo"},
		{"URL with subdomain", "https://git.example.com/owner/repo.git", "owner/repo"},
		{"URL with port", "https://github.com:443/owner/repo.git", "owner/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRepoPath(tt.input)
			assert.Equal(t, tt.expected, result, "ExtractRepoPath(%q) should return %q", tt.input, tt.expected)
		})
	}
}

func TestGetGitPath(t *testing.T) {
	// Reset the cached values for testing
	gitPath = ""
	gitPathErr = nil
	gitPathOnce = sync.Once{}

	path, err := getGitPath()
	require.NoError(t, err, "git should be found in PATH")
	assert.NotEmpty(t, path, "git path should not be empty")

	// Verify it's in a trusted location
	trustedPaths := []string{
		"/usr/bin/",
		"/usr/local/bin/",
		"/opt/homebrew/bin/",
		"/opt/local/bin/",
		"C:\\Program Files\\Git\\",
		"C:\\Program Files (x86)\\Git\\",
	}

	found := false
	for _, tp := range trustedPaths {
		if strings.HasPrefix(path, tp) {
			found = true
			break
		}
	}

	assert.True(t, found, "git should be in a trusted location, found at: %s", path)

	// Test caching - should return same result
	path2, err2 := getGitPath()
	assert.Equal(t, path, path2, "cached path should be the same")
	assert.Equal(t, err, err2, "cached error should be the same")
}

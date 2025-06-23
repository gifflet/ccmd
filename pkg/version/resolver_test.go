// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ResolveVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		tags        []string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:     "exact tag match",
			version:  "v1.2.3",
			tags:     []string{"v1.0.0", "v1.2.3", "v2.0.0"},
			expected: "v1.2.3",
		},
		{
			name:     "exact tag without v prefix",
			version:  "1.2.3",
			tags:     []string{"v1.0.0", "v1.2.3", "v2.0.0"},
			expected: "v1.2.3",
		},
		{
			name:     "exact tag with v when tag has no v",
			version:  "v1.2.3",
			tags:     []string{"1.0.0", "1.2.3", "2.0.0"},
			expected: "1.2.3",
		},
		{
			name:     "latest keyword",
			version:  "latest",
			tags:     []string{"v1.0.0", "v1.2.3", "v2.0.0", "v2.1.0"},
			expected: "v2.1.0",
		},
		{
			name:        "latest with no tags",
			version:     "latest",
			tags:        []string{},
			expectError: true,
			errorMsg:    "no tags found",
		},
		{
			name:        "latest with no semver tags",
			version:     "latest",
			tags:        []string{"release-1", "feature-xyz"},
			expectError: true,
			errorMsg:    "no semantic version tags found",
		},
		{
			name:     "caret constraint",
			version:  "^1.0.0",
			tags:     []string{"v0.9.0", "v1.0.0", "v1.2.3", "v1.9.0", "v2.0.0"},
			expected: "v1.9.0",
		},
		{
			name:     "tilde constraint",
			version:  "~1.2.0",
			tags:     []string{"v1.0.0", "v1.2.0", "v1.2.3", "v1.3.0", "v2.0.0"},
			expected: "v1.2.3",
		},
		{
			name:     "greater than constraint",
			version:  ">1.5.0",
			tags:     []string{"v1.0.0", "v1.5.0", "v1.6.0", "v2.0.0"},
			expected: "v2.0.0",
		},
		{
			name:     "range constraint",
			version:  ">=1.0.0 <2.0.0",
			tags:     []string{"v0.9.0", "v1.0.0", "v1.5.0", "v2.0.0", "v2.1.0"},
			expected: "v1.5.0",
		},
		{
			name:        "constraint with no matching version",
			version:     "^3.0.0",
			tags:        []string{"v1.0.0", "v2.0.0", "v2.5.0"},
			expectError: true,
			errorMsg:    "no version found matching constraint",
		},
		{
			name:     "branch name",
			version:  "main",
			tags:     []string{"v1.0.0", "v2.0.0"},
			expected: "main",
		},
		{
			name:     "commit hash",
			version:  "abc123def",
			tags:     []string{"v1.0.0", "v2.0.0"},
			expected: "abc123def",
		},
		{
			name:        "empty version",
			version:     "",
			tags:        []string{"v1.0.0"},
			expectError: true,
			errorMsg:    "version cannot be empty",
		},
		{
			name:     "mixed tag formats",
			version:  "latest",
			tags:     []string{"1.0.0", "v1.2.0", "1.5.0", "v2.0.0", "release-1"},
			expected: "v2.0.0",
		},
		{
			name:     "pre-release versions",
			version:  "^1.0.0",
			tags:     []string{"v1.0.0", "v1.1.0-beta", "v1.2.0", "v2.0.0-alpha"},
			expected: "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(func(string) ([]string, error) {
				return tt.tags, nil
			})

			result, err := resolver.ResolveVersion("/fake/repo", tt.version)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestResolver_ResolveVersion_GetTagsError(t *testing.T) {
	resolver := NewResolver(func(string) ([]string, error) {
		return nil, fmt.Errorf("failed to access repository")
	})

	_, err := resolver.ResolveVersion("/fake/repo", "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tags")
}

func TestParseSemverTag(t *testing.T) {
	tests := []struct {
		tag         string
		expectError bool
	}{
		{"v1.2.3", false},
		{"1.2.3", false},
		{"v1.0.0-alpha", false},
		{"1.0.0-beta+build123", false},
		{"release-1", true},
		{"feature/xyz", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			v, err := parseSemverTag(tt.tag)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, v)
			}
		})
	}
}

func TestIsSemverConstraint(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"^1.0.0", true},
		{"~1.2.0", true},
		{">1.0.0", true},
		{">=1.0.0", true},
		{"<2.0.0", true},
		{"<=2.0.0", true},
		{"=1.0.0", true},
		{"1.0.0 - 2.0.0", true},
		{"^1.0.0 || ^2.0.0", true},
		{"1.x", true},
		{"1.*", true},
		{"1.X.0", true},
		{"v1.0.0", false},
		{"1.0.0", false},
		{"main", false},
		{"feature/xyz", false},
		{"abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isSemverConstraint(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindExactTag(t *testing.T) {
	resolver := &Resolver{}
	tags := []string{"v1.0.0", "v1.2.3", "2.0.0", "release-1"}

	tests := []struct {
		version  string
		expected string
	}{
		{"v1.0.0", "v1.0.0"},
		{"v1.2.3", "v1.2.3"},
		{"2.0.0", "2.0.0"},
		{"v2.0.0", "2.0.0"},
		{"1.0.0", "v1.0.0"},
		{"release-1", "release-1"},
		{"v3.0.0", ""},
		{"main", ""},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := resolver.findExactTag(tt.version, tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

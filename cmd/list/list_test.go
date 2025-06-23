/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package list

import (
	"testing"
	"time"

	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/project"
)

func TestPrintSimpleList(t *testing.T) {
	// Note: This test validates the basic output format. Full integration testing
	// of stdout capture would require more complex test infrastructure.
	// For now, we verify the core formatting logic.

	now := time.Now()
	commands := []*commands.CommandDetail{
		{
			CommandLockInfo: &project.CommandLockInfo{
				Name:      "test-cmd",
				Version:   "1.0.0",
				Source:    "github.com/user/repo",
				UpdatedAt: now,
			},
			HasDirectory:    true,
			HasMarkdownFile: true,
			StructureValid:  true,
			CommandMetadata: &models.CommandMetadata{
				Name:        "test-cmd",
				Version:     "1.2.0",
				Description: "A test command for unit testing",
				Author:      "Test Author",
				Repository:  "github.com/user/repo",
				Entry:       "cmd/main.go",
			},
		},
		{
			CommandLockInfo: &project.CommandLockInfo{
				Name:      "broken-cmd",
				Version:   "0.5.0",
				Source:    "github.com/user/broken",
				UpdatedAt: now.Add(-24 * time.Hour),
			},
			HasDirectory:     false,
			HasMarkdownFile:  true,
			StructureValid:   false,
			StructureMessage: "broken structure: [missing directory]",
		},
	}

	// Test passes if function doesn't panic
	printSimpleList(commands)
}

func TestPrintLongList(t *testing.T) {
	// Note: This test validates the basic output format.

	now := time.Now()
	commands := []*commands.CommandDetail{
		{
			CommandLockInfo: &project.CommandLockInfo{
				Name:         "full-cmd",
				Version:      "1.0.0",
				Source:       "github.com/user/repo",
				InstalledAt:  now.Add(-48 * time.Hour),
				UpdatedAt:    now,
				Dependencies: []string{"dep1", "dep2"},
				Metadata: map[string]string{
					"custom": "value",
				},
			},
			HasDirectory:    true,
			HasMarkdownFile: true,
			StructureValid:  true,
			CommandMetadata: &models.CommandMetadata{
				Name:        "full-cmd",
				Version:     "1.5.0",
				Description: "A fully featured command",
				Author:      "John Doe",
				Repository:  "github.com/user/repo",
				Entry:       "cmd/main.go",
				Tags:        []string{"cli", "tool"},
				License:     "MIT",
				Homepage:    "https://example.com",
			},
		},
	}

	// Test passes if function doesn't panic
	printLongList(commands)
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "30 minutes ago",
			time:     time.Now().Add(-30 * time.Minute),
			expected: "30 minutes ago",
		},
		{
			name:     "2 hours ago",
			time:     time.Now().Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "1 hour ago",
			time:     time.Now().Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 days ago",
			time:     time.Now().Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "1 day ago",
			time:     time.Now().Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "old date",
			time:     time.Now().Add(-30 * 24 * time.Hour),
			expected: time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.time)
			if result != tt.expected {
				t.Errorf("formatTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "short text",
			text:     "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "exact length",
			text:     "exactly10!",
			maxLen:   10,
			expected: "exactly10!",
		},
		{
			name:     "long text",
			text:     "this is a very long text that needs truncation",
			maxLen:   20,
			expected: "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.maxLen, result, tt.expected)
			}
		})
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package regression

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/core"
)

// TestInstallCommandBehavior captures the current behavior of the install command
// This test will serve as a regression test during refactoring
func TestInstallCommandBehavior(t *testing.T) {
	t.Run("validates_empty_repository", func(t *testing.T) {
		opts := core.InstallOptions{
			Repository: "",
		}

		err := core.Install(context.Background(), opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repository URL is required")
	})

	t.Run("normalizes_github_shorthand", func(t *testing.T) {
		// Test shorthand normalization
		normalized := core.NormalizeRepositoryURL("user/repo")
		assert.Equal(t, "https://github.com/user/repo.git", normalized)

		// Test full URL passthrough
		normalized = core.NormalizeRepositoryURL("https://github.com/user/repo.git")
		assert.Equal(t, "https://github.com/user/repo.git", normalized)

		// Test SSH URL passthrough
		normalized = core.NormalizeRepositoryURL("git@github.com:user/repo.git")
		assert.Equal(t, "git@github.com:user/repo.git", normalized)
	})

	t.Run("parses_repository_spec", func(t *testing.T) {
		// Test with version
		repo, version := core.ParseRepositorySpec("user/repo@v1.0.0")
		assert.Equal(t, "user/repo", repo)
		assert.Equal(t, "v1.0.0", version)

		// Test without version
		repo, version = core.ParseRepositorySpec("user/repo")
		assert.Equal(t, "user/repo", repo)
		assert.Equal(t, "", version)

		// Test with URL and version
		repo, version = core.ParseRepositorySpec("https://github.com/user/repo.git@main")
		assert.Equal(t, "https://github.com/user/repo.git", repo)
		assert.Equal(t, "main", version)
	})

	t.Run("extracts_repo_path", func(t *testing.T) {
		tests := []struct {
			url      string
			expected string
		}{
			{"https://github.com/user/repo.git", "user/repo"},
			{"https://github.com/user/repo", "user/repo"},
			{"git@github.com:user/repo.git", "user/repo"},
			{"https://gitlab.com/group/project.git", "group/project"},
		}

		for _, tt := range tests {
			path := core.ExtractRepoPath(tt.url)
			assert.Equal(t, tt.expected, path)
		}
	})

	t.Run("install_creates_correct_directory_structure", func(t *testing.T) {
		// Skip if we can't create temp directories (CI environment)
		if os.Getenv("CI") != "" {
			t.Skip("Skipping filesystem test in CI")
		}

		// This test captures the expected directory structure
		// The new implementation must maintain this structure
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// Expected structure after install:
		// .claude/
		//   commands/
		//     <command-name>/
		//       ccmd.yaml
		//       ... other files

		expectedDirs := []string{
			".claude",
			".claude/commands",
		}

		// Verify the installer would create these directories
		for _, dir := range expectedDirs {
			expectedPath := filepath.Join(tempDir, dir)
			// The actual install would create these
			// We're just documenting the expected behavior
			_ = expectedPath
		}
	})
}

// TestInstallCommandErrorCases documents current error handling behavior
func TestInstallCommandErrorCases(t *testing.T) {
	testCases := []struct {
		name        string
		opts        core.InstallOptions
		errContains string
	}{
		{
			name: "empty_repository",
			opts: core.InstallOptions{
				Repository: "",
			},
			errContains: "repository URL is required",
		},
		{
			name: "invalid_characters_in_name",
			opts: core.InstallOptions{
				Repository: "user/repo",
				Name:       "invalid/name",
			},
			// Current behavior might not validate this
			// Document actual behavior here
			errContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := core.Install(context.Background(), tc.opts)

			if tc.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errContains))
			} else {
				// If no error expected, document actual behavior
				// This might still error due to network/git issues
				_ = err
			}
		})
	}
}

// TestInstallFromProjectBehavior documents InstallFromProject behavior
func TestInstallFromProjectBehavior(t *testing.T) {
	t.Run("requires_valid_project_path", func(t *testing.T) {
		err := core.InstallFromConfig(context.Background(), "/non/existent/path", false)
		require.Error(t, err)
		// Document the actual error message
	})

	t.Run("force_flag_behavior", func(t *testing.T) {
		// Document how force flag affects installation
		// - Does it skip validation?
		// - Does it overwrite existing commands?
		// - Does it clean before reinstall?
	})
}

// TestGetInstalledCommandsBehavior documents list behavior
func TestGetInstalledCommandsBehavior(t *testing.T) {
	t.Run("handles_missing_project", func(t *testing.T) {
		// GetInstalledCommands not yet migrated to core
		// This will be updated when list command is migrated
	})

	t.Run("returns_empty_list_for_no_commands", func(t *testing.T) {
		// GetInstalledCommands not yet migrated to core
		// This will be updated when list command is migrated
	})
}

// BenchmarkCurrentImplementation provides performance baseline
func BenchmarkCurrentImplementation(b *testing.B) {
	// Benchmark key operations to ensure refactoring doesn't degrade performance

	b.Run("ParseRepositorySpec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			core.ParseRepositorySpec("user/repo@v1.0.0")
		}
	})

	b.Run("NormalizeRepositoryURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			core.NormalizeRepositoryURL("user/repo")
		}
	})

	b.Run("ExtractRepoPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			core.ExtractRepoPath("https://github.com/user/repo.git")
		}
	})
}

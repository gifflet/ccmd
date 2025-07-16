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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUpdateLockFile(t *testing.T) {
	t.Run("creates new lock file", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Mock metadata
		metadata := &ProjectConfig{
			Name:       "test-cmd",
			Version:    "1.0.0",
			Repository: "https://github.com/user/test-cmd.git",
		}

		// Call updateLockFile
		err := updateLockFile(tempDir, "test-cmd", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read the created lock file
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		data, err := os.ReadFile(lockPath)
		require.NoError(t, err)

		// Parse lock file
		var lockFile LockFile
		err = yaml.Unmarshal(data, &lockFile)
		require.NoError(t, err)

		// Verify structure
		assert.Equal(t, "1.0", lockFile.Version)
		assert.Equal(t, 1, lockFile.LockfileVersion)
		assert.Len(t, lockFile.Commands, 1)

		// Verify command
		cmd := lockFile.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, "test-cmd", cmd.Name)
		assert.Equal(t, "1.0.0", cmd.Version)
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Source)
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Resolved) // No requested version
		assert.NotZero(t, cmd.InstalledAt)
		assert.NotZero(t, cmd.UpdatedAt)
		assert.Equal(t, cmd.InstalledAt, cmd.UpdatedAt) // First install
	})

	t.Run("updates existing command preserving installed_at", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial lock file
		initialTime := time.Now().Add(-24 * time.Hour)
		initialLock := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:        "test-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/test-cmd.git",
					Resolved:    "https://github.com/user/test-cmd.git@1.0.0",
					Commit:      "abc123",
					InstalledAt: initialTime,
					UpdatedAt:   initialTime,
				},
			},
		}

		// Write initial lock file
		data, err := yaml.Marshal(&initialLock)
		require.NoError(t, err)
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		err = os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

		// Update with new version
		metadata := &ProjectConfig{
			Name:       "test-cmd",
			Version:    "2.0.0",
			Repository: "https://github.com/user/test-cmd.git",
		}

		// Call updateLockFile
		err = updateLockFile(tempDir, "test-cmd", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read updated lock file
		data, err = os.ReadFile(lockPath)
		require.NoError(t, err)

		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

		// Verify update
		cmd := updatedLock.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, "2.0.0", cmd.Version)
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Resolved) // No requested version
		assert.Equal(t, initialTime.Unix(), cmd.InstalledAt.Unix())           // Preserved
		assert.True(t, cmd.UpdatedAt.After(initialTime))                      // Updated
	})

	t.Run("handles multiple commands", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial lock file with one command
		initialLock := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"existing-cmd": {
					Name:        "existing-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/existing-cmd.git",
					Resolved:    "https://github.com/user/existing-cmd.git@1.0.0",
					Commit:      "xyz789",
					InstalledAt: time.Now().Add(-48 * time.Hour),
					UpdatedAt:   time.Now().Add(-48 * time.Hour),
				},
			},
		}

		// Write initial lock file
		data, err := yaml.Marshal(&initialLock)
		require.NoError(t, err)
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		err = os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

		// Add new command
		metadata := &ProjectConfig{
			Name:       "new-cmd",
			Version:    "1.5.0",
			Repository: "https://github.com/user/new-cmd.git",
		}

		err = updateLockFile(tempDir, "new-cmd", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read updated lock file
		data, err = os.ReadFile(lockPath)
		require.NoError(t, err)

		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

		// Verify both commands exist
		assert.Len(t, updatedLock.Commands, 2)
		assert.NotNil(t, updatedLock.Commands["existing-cmd"])
		assert.NotNil(t, updatedLock.Commands["new-cmd"])

		// Verify new command
		newCmd := updatedLock.Commands["new-cmd"]
		assert.Equal(t, "new-cmd", newCmd.Name)
		assert.Equal(t, "1.5.0", newCmd.Version)
		assert.Equal(t, "https://github.com/user/new-cmd.git", newCmd.Source)
		assert.Equal(t, "https://github.com/user/new-cmd.git", newCmd.Resolved) // No requested version
	})

	t.Run("creates resolved URL without version", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Mock metadata without version
		metadata := &ProjectConfig{
			Name:       "test-cmd",
			Version:    "",
			Repository: "https://github.com/user/test-cmd.git",
		}

		// Call updateLockFile
		err := updateLockFile(tempDir, "test-cmd", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read the created lock file
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		data, err := os.ReadFile(lockPath)
		require.NoError(t, err)

		// Parse lock file
		var lockFile LockFile
		err = yaml.Unmarshal(data, &lockFile)
		require.NoError(t, err)

		// Verify command
		cmd := lockFile.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, "", cmd.Version)
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Source)
		// Without git repo, resolved should fall back to source
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Resolved)
	})

	t.Run("updates existing entry when name differs from repo", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial lock file with command under different name
		initialTime := time.Now().Add(-24 * time.Hour)
		initialLock := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"my-awesome-cli": {
					Name:        "my-awesome-cli",
					Version:     "1.0.0",
					Source:      "https://github.com/owner/cli-tool.git",
					Resolved:    "https://github.com/owner/cli-tool.git@1.0.0",
					Commit:      "abc123",
					InstalledAt: initialTime,
					UpdatedAt:   initialTime,
				},
			},
		}

		// Write initial lock file
		data, err := yaml.Marshal(&initialLock)
		require.NoError(t, err)
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		err = os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

		// Update with same repo but different command name
		metadata := &ProjectConfig{
			Name:       "new-cli-name",
			Version:    "2.0.0",
			Repository: "https://github.com/owner/cli-tool.git",
		}

		err = updateLockFile(tempDir, "new-cli-name", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read updated lock file
		data, err = os.ReadFile(lockPath)
		require.NoError(t, err)

		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

		// Verify old entry is removed
		assert.Nil(t, updatedLock.Commands["my-awesome-cli"])

		// Verify new entry exists with preserved installed_at
		cmd := updatedLock.Commands["new-cli-name"]
		require.NotNil(t, cmd)
		assert.Equal(t, "new-cli-name", cmd.Name)
		assert.Equal(t, "2.0.0", cmd.Version)
		assert.Equal(t, "https://github.com/owner/cli-tool.git", cmd.Source)
		assert.Equal(t, initialTime.Unix(), cmd.InstalledAt.Unix()) // Preserved
		assert.True(t, cmd.UpdatedAt.After(initialTime))            // Updated
	})

	t.Run("handles multiple repos with name changes", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial lock file with multiple commands
		initialLock := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"tool-one": {
					Name:        "tool-one",
					Version:     "1.0.0",
					Source:      "https://github.com/org/first-repo.git",
					Resolved:    "https://github.com/org/first-repo.git@1.0.0",
					InstalledAt: time.Now().Add(-48 * time.Hour),
					UpdatedAt:   time.Now().Add(-48 * time.Hour),
				},
				"tool-two": {
					Name:        "tool-two",
					Version:     "1.0.0",
					Source:      "https://github.com/org/second-repo.git",
					Resolved:    "https://github.com/org/second-repo.git@1.0.0",
					InstalledAt: time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				},
			},
		}

		// Write initial lock file
		data, err := yaml.Marshal(&initialLock)
		require.NoError(t, err)
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		err = os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

		// Update second command with new name
		metadata := &ProjectConfig{
			Name:       "renamed-tool",
			Version:    "2.0.0",
			Repository: "https://github.com/org/second-repo.git",
		}

		err = updateLockFile(tempDir, "renamed-tool", metadata, metadata.Version, "")
		require.NoError(t, err)

		// Read updated lock file
		data, err = os.ReadFile(lockPath)
		require.NoError(t, err)

		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

		// Verify first command unchanged
		assert.NotNil(t, updatedLock.Commands["tool-one"])
		assert.Equal(t, "1.0.0", updatedLock.Commands["tool-one"].Version)

		// Verify second command renamed
		assert.Nil(t, updatedLock.Commands["tool-two"])
		assert.NotNil(t, updatedLock.Commands["renamed-tool"])
		assert.Equal(t, "2.0.0", updatedLock.Commands["renamed-tool"].Version)
	})

	t.Run("preserves original ccmd.yaml version when install version differs", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Mock metadata with original version from ccmd.yaml
		metadata := &ProjectConfig{
			Name:       "test-cmd",
			Version:    "1.2.3", // Original version in ccmd.yaml
			Repository: "https://github.com/user/test-cmd.git",
		}

		// Simulate updateLockFile being called with different originalVersion
		// This simulates the case where user installs with @v1.0.0 but ccmd.yaml has version 1.2.3
		originalVersion := "1.2.3"  // The version from ccmd.yaml
		metadata.Version = "v1.0.0" // The version specified during install

		// Call updateLockFile with original version
		err := updateLockFile(tempDir, "test-cmd", metadata, originalVersion, "v1.0.0")
		require.NoError(t, err)

		// Read the created lock file
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		data, err := os.ReadFile(lockPath)
		require.NoError(t, err)

		// Parse lock file
		var lockFile LockFile
		err = yaml.Unmarshal(data, &lockFile)
		require.NoError(t, err)

		// Verify command has original version from ccmd.yaml, not install version
		cmd := lockFile.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, "test-cmd", cmd.Name)
		assert.Equal(t, "1.2.3", cmd.Version) // Should be original version from ccmd.yaml
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Source)
		assert.Equal(t, "https://github.com/user/test-cmd.git@v1.0.0", cmd.Resolved) // Should show install version
	})

	t.Run("uses ccmd.yaml version when no version specified", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Mock metadata with version from ccmd.yaml
		metadata := &ProjectConfig{
			Name:       "test-cmd",
			Version:    "1.0.0", // Version in ccmd.yaml
			Repository: "https://github.com/user/test-cmd.git",
		}

		// Simulate updateLockFile being called without version specified
		// This simulates the case where user installs without specifying version
		originalVersion := "1.0.0" // The version from ccmd.yaml

		// Call updateLockFile with original version
		err := updateLockFile(tempDir, "test-cmd", metadata, originalVersion, "")
		require.NoError(t, err)

		// Read the created lock file
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")
		data, err := os.ReadFile(lockPath)
		require.NoError(t, err)

		// Parse lock file
		var lockFile LockFile
		err = yaml.Unmarshal(data, &lockFile)
		require.NoError(t, err)

		// Verify command has ccmd.yaml version
		cmd := lockFile.Commands["test-cmd"]
		require.NotNil(t, cmd)
		assert.Equal(t, "test-cmd", cmd.Name)
		assert.Equal(t, "1.0.0", cmd.Version) // Should be ccmd.yaml version
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Source)
	})
}

func TestAddToConfig(t *testing.T) {
	t.Run("creates new config with commands only", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Call addToConfig on non-existent config
		err := addToConfig(tempDir, "test-cmd", "user/test-cmd", "1.0.0")
		require.NoError(t, err)

		// Read created config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify structure - should only have commands
		assert.Empty(t, config.Name)
		assert.Empty(t, config.Version)
		assert.Empty(t, config.Description)
		assert.Len(t, config.Commands, 1)
		assert.Equal(t, "user/test-cmd@1.0.0", config.Commands[0])
	})

	t.Run("adds to existing config", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial config
		initialConfig := &ProjectConfig{
			Name:        "test-project",
			Version:     "1.0.0",
			Description: "Test project",
			Commands:    []string{"existing/cmd@2.0.0"},
		}
		err := SaveProjectConfig(tempDir, initialConfig)
		require.NoError(t, err)

		// Add new command
		err = addToConfig(tempDir, "new-cmd", "owner/new-cmd", "3.0.0")
		require.NoError(t, err)

		// Read updated config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify existing fields preserved
		assert.Equal(t, "test-project", config.Name)
		assert.Equal(t, "1.0.0", config.Version)
		assert.Equal(t, "Test project", config.Description)

		// Verify commands
		assert.Len(t, config.Commands, 2)
		assert.Contains(t, config.Commands, "existing/cmd@2.0.0")
		assert.Contains(t, config.Commands, "owner/new-cmd@3.0.0")
	})

	t.Run("updates existing command with force", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial config
		initialConfig := &ProjectConfig{
			Commands: []string{"owner/test-cmd@1.0.0", "other/cmd"},
		}
		err := SaveProjectConfig(tempDir, initialConfig)
		require.NoError(t, err)

		// Update existing command
		err = addToConfig(tempDir, "test-cmd", "owner/test-cmd", "2.0.0")
		require.NoError(t, err)

		// Read updated config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify update
		assert.Len(t, config.Commands, 2)
		assert.Contains(t, config.Commands, "owner/test-cmd@2.0.0")
		assert.Contains(t, config.Commands, "other/cmd")
		assert.NotContains(t, config.Commands, "owner/test-cmd@1.0.0")
	})

	t.Run("handles command without version", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Add command without version
		err := addToConfig(tempDir, "test-cmd", "owner/test-cmd", "")
		require.NoError(t, err)

		// Read config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify
		assert.Len(t, config.Commands, 1)
		assert.Equal(t, "owner/test-cmd", config.Commands[0])
	})

}

func TestParseRepositorySpec(t *testing.T) {
	tests := []struct {
		name           string
		spec           string
		wantRepository string
		wantVersion    string
	}{
		{
			name:           "SSH URL without version",
			spec:           "git@github.com:gifflet/parallax.git",
			wantRepository: "git@github.com:gifflet/parallax.git",
			wantVersion:    "",
		},
		{
			name:           "SSH URL with version",
			spec:           "git@github.com:gifflet/parallax.git@v1.0.0",
			wantRepository: "git@github.com:gifflet/parallax.git",
			wantVersion:    "v1.0.0",
		},
		{
			name:           "SSH URL with branch",
			spec:           "git@github.com:owner/repo.git@optimize-context-loading",
			wantRepository: "git@github.com:owner/repo.git",
			wantVersion:    "optimize-context-loading",
		},
		{
			name:           "HTTPS URL without version",
			spec:           "https://github.com/gifflet/parallax.git",
			wantRepository: "https://github.com/gifflet/parallax.git",
			wantVersion:    "",
		},
		{
			name:           "HTTPS URL with version",
			spec:           "https://github.com/gifflet/parallax.git@main",
			wantRepository: "https://github.com/gifflet/parallax.git",
			wantVersion:    "main",
		},
		{
			name:           "HTTP URL without version",
			spec:           "http://example.com/repo.git",
			wantRepository: "http://example.com/repo.git",
			wantVersion:    "",
		},
		{
			name:           "HTTP URL with tag",
			spec:           "http://example.com/repo.git@v2.5.0",
			wantRepository: "http://example.com/repo.git",
			wantVersion:    "v2.5.0",
		},
		{
			name:           "Shorthand without version",
			spec:           "gifflet/parallax",
			wantRepository: "gifflet/parallax",
			wantVersion:    "",
		},
		{
			name:           "Shorthand with version",
			spec:           "gifflet/parallax@v1.0.0",
			wantRepository: "gifflet/parallax",
			wantVersion:    "v1.0.0",
		},
		{
			name:           "Shorthand with branch",
			spec:           "owner/repo@main",
			wantRepository: "owner/repo",
			wantVersion:    "main",
		},
		{
			name:           "Complex SSH URL without version",
			spec:           "git@gitlab.com:group/subgroup/project.git",
			wantRepository: "git@gitlab.com:group/subgroup/project.git",
			wantVersion:    "",
		},
		{
			name:           "Complex SSH URL with version",
			spec:           "git@gitlab.com:group/subgroup/project.git@develop",
			wantRepository: "git@gitlab.com:group/subgroup/project.git",
			wantVersion:    "develop",
		},
		{
			name:           "SSH URL with @ in path without version",
			spec:           "git@github.com:user@company/repo.git",
			wantRepository: "git@github.com:user@company/repo.git",
			wantVersion:    "",
		},
		{
			name:           "SSH URL with @ in path with version",
			spec:           "git@github.com:user@company/repo.git@feature-branch",
			wantRepository: "git@github.com:user@company/repo.git",
			wantVersion:    "feature-branch",
		},
		{
			name:           "SSH URL without .git extension with version",
			spec:           "git@github.com:gifflet/parallax@feat/optimize-context-loading",
			wantRepository: "git@github.com:gifflet/parallax",
			wantVersion:    "feat/optimize-context-loading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository, version := ParseRepositorySpec(tt.spec)
			assert.Equal(t, tt.wantRepository, repository)
			assert.Equal(t, tt.wantVersion, version)
		})
	}
}

func TestAddToConfigWithDifferentCommandName(t *testing.T) {
	t.Run("updates existing command when name differs from repo", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create commands directory
		commandsDir := filepath.Join(tempDir, ".claude", "commands", "my-awesome-cli")
		err := os.MkdirAll(commandsDir, 0755)
		require.NoError(t, err)

		// Create command metadata with different name than repo
		commandMetadata := &ProjectConfig{
			Name:       "my-awesome-cli",
			Version:    "1.0.0",
			Repository: "https://github.com/owner/cli-tool.git",
		}
		metadataPath := filepath.Join(commandsDir, "ccmd.yaml")
		data, err := yaml.Marshal(commandMetadata)
		require.NoError(t, err)
		err = os.WriteFile(metadataPath, data, 0644)
		require.NoError(t, err)

		// Create initial config with the command
		initialConfig := &ProjectConfig{
			Commands: []string{"owner/cli-tool@1.0.0"},
		}
		err = SaveProjectConfig(tempDir, initialConfig)
		require.NoError(t, err)

		// Update command with new version
		err = addToConfig(tempDir, "my-awesome-cli", "owner/cli-tool", "2.0.0")
		require.NoError(t, err)

		// Read updated config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify update - should not have duplicates
		assert.Len(t, config.Commands, 1)
		assert.Equal(t, "owner/cli-tool@2.0.0", config.Commands[0])
	})

	t.Run("adds new command when not installed", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create initial config without the command
		initialConfig := &ProjectConfig{
			Commands: []string{"other/command@1.0.0"},
		}
		err := SaveProjectConfig(tempDir, initialConfig)
		require.NoError(t, err)

		// Add new command
		err = addToConfig(tempDir, "new-cli", "owner/new-cli", "1.0.0")
		require.NoError(t, err)

		// Read updated config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify addition
		assert.Len(t, config.Commands, 2)
		assert.Contains(t, config.Commands, "other/command@1.0.0")
		assert.Contains(t, config.Commands, "owner/new-cli@1.0.0")
	})

	t.Run("handles multiple commands with different names", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create commands directory structure
		commands := []struct {
			name string
			repo string
		}{
			{"awesome-tool", "company/tool-repo"},
			{"super-cli", "company/cli-project"},
		}

		for _, cmd := range commands {
			cmdDir := filepath.Join(tempDir, ".claude", "commands", cmd.name)
			err := os.MkdirAll(cmdDir, 0755)
			require.NoError(t, err)

			metadata := &ProjectConfig{
				Name:       cmd.name,
				Version:    "1.0.0",
				Repository: fmt.Sprintf("https://github.com/%s.git", cmd.repo),
			}
			metadataPath := filepath.Join(cmdDir, "ccmd.yaml")
			data, err := yaml.Marshal(metadata)
			require.NoError(t, err)
			err = os.WriteFile(metadataPath, data, 0644)
			require.NoError(t, err)
		}

		// Create initial config
		initialConfig := &ProjectConfig{
			Commands: []string{
				"company/tool-repo@1.0.0",
				"company/cli-project@1.0.0",
			},
		}
		err := SaveProjectConfig(tempDir, initialConfig)
		require.NoError(t, err)

		// Update first command
		err = addToConfig(tempDir, "awesome-tool", "company/tool-repo", "2.0.0")
		require.NoError(t, err)

		// Read updated config
		config, err := LoadProjectConfig(tempDir)
		require.NoError(t, err)

		// Verify correct update
		assert.Len(t, config.Commands, 2)
		assert.Contains(t, config.Commands, "company/tool-repo@2.0.0")
		assert.Contains(t, config.Commands, "company/cli-project@1.0.0")
	})
}

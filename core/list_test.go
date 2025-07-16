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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("returns empty list when no lock file exists", func(t *testing.T) {
		tempDir := t.TempDir()

		commands, err := List(ListOptions{ProjectPath: tempDir})
		require.NoError(t, err)
		assert.Empty(t, commands)
	})

	t.Run("returns commands from lock file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["test-cmd"] = &LockCommand{
			Name:        "test-cmd",
			Version:     "1.0.0",
			Source:      "https://github.com/user/test-cmd.git",
			Resolved:    "https://github.com/user/test-cmd.git@v1.0.0",
			Commit:      "abc123",
			InstalledAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		}
		lockFile.Commands["another-cmd"] = &LockCommand{
			Name:        "another-cmd",
			Version:     "2.0.0",
			Source:      "https://github.com/user/another-cmd.git",
			Resolved:    "https://github.com/user/another-cmd.git@v2.0.0",
			Commit:      "def456",
			InstalledAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create command structures
		oldDir, _ := os.Getwd()
		os.Chdir(tempDir)
		createCommandStructure(t, "test-cmd")
		createCommandStructure(t, "another-cmd")
		os.Chdir(oldDir)

		// List commands
		commands, err := List(ListOptions{ProjectPath: tempDir})
		require.NoError(t, err)
		assert.Len(t, commands, 2)

		// Check sorting by name
		assert.Equal(t, "another-cmd", commands[0].Name)
		assert.Equal(t, "test-cmd", commands[1].Name)

		// Verify command details
		cmd := commands[1] // test-cmd
		assert.Equal(t, "test-cmd", cmd.Name)
		assert.Equal(t, "1.0.0", cmd.Version)
		assert.Equal(t, "https://github.com/user/test-cmd.git", cmd.Repository)
		assert.Equal(t, "https://github.com/user/test-cmd.git@v1.0.0", cmd.Resolved)
		assert.False(t, cmd.BrokenStructure)
		assert.Equal(t, "", cmd.StructureError)
	})

	t.Run("detects broken structure - missing directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["broken-cmd"] = &LockCommand{
			Name:        "broken-cmd",
			Version:     "1.0.0",
			Source:      "https://github.com/user/broken-cmd.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create .md file but no command directory
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "commands", "broken-cmd.md"), []byte("# broken-cmd"), 0644)

		// List commands
		commands, err := List(ListOptions{ProjectPath: tempDir})
		require.NoError(t, err)
		assert.Len(t, commands, 1)

		cmd := commands[0]
		assert.True(t, cmd.BrokenStructure)
		assert.Equal(t, "command directory not found", cmd.StructureError)
	})

	t.Run("detects broken structure - missing md file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["broken-cmd"] = &LockCommand{
			Name:        "broken-cmd",
			Version:     "1.0.0",
			Source:      "https://github.com/user/broken-cmd.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create command directory but no .md file
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands", "broken-cmd"), 0755)

		// List commands
		commands, err := List(ListOptions{ProjectPath: tempDir})
		require.NoError(t, err)
		assert.Len(t, commands, 1)

		cmd := commands[0]
		assert.True(t, cmd.BrokenStructure)
		assert.Equal(t, "standalone .md file not found", cmd.StructureError)
	})

	t.Run("reads metadata from ccmd.yaml in command directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["meta-cmd"] = &LockCommand{
			Name:        "meta-cmd",
			Version:     "", // Empty in lock file
			Source:      "https://github.com/user/meta-cmd.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create command directory and files
		cmdDir := filepath.Join(tempDir, ".claude", "commands", "meta-cmd")
		os.MkdirAll(cmdDir, 0755)
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "commands", "meta-cmd.md"), []byte("# meta-cmd"), 0644)

		// Create metadata file
		metadata := ProjectConfig{
			Name:        "meta-cmd",
			Version:     "1.2.3",
			Description: "Test command with metadata",
			Author:      "Test Author",
			Repository:  "https://github.com/user/meta-cmd.git",
			Entry:       "main.go",
			Tags:        []string{"test", "example"},
			License:     "MIT",
			Homepage:    "https://example.com",
		}
		err := writeCommandMetadata(filepath.Join(cmdDir, "ccmd.yaml"), &metadata)
		require.NoError(t, err)

		// List commands
		commands, err := List(ListOptions{ProjectPath: tempDir})
		require.NoError(t, err)
		assert.Len(t, commands, 1)

		cmd := commands[0]
		assert.Equal(t, "meta-cmd", cmd.Name)
		assert.Equal(t, "1.2.3", cmd.Version) // From metadata
		assert.Equal(t, "Test command with metadata", cmd.Description)
		assert.Equal(t, "Test Author", cmd.Author)
		assert.Equal(t, []string{"test", "example"}, cmd.Tags)
		assert.Equal(t, "MIT", cmd.License)
		assert.Equal(t, "https://example.com", cmd.Homepage)
		assert.Equal(t, "main.go", cmd.Entry)
	})
}

func TestGetCommandInfo(t *testing.T) {
	t.Run("returns command info when exists", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file with one command
		lockFile := createBasicLockFile()
		lockFile.Commands["test-cmd"] = &LockCommand{
			Name:        "test-cmd",
			Version:     "1.0.0",
			Source:      "https://github.com/user/test-cmd.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create command structure
		oldDir, _ := os.Getwd()
		os.Chdir(tempDir)
		createCommandStructure(t, "test-cmd")
		os.Chdir(oldDir)

		// Get command info
		info, err := GetCommandInfo("test-cmd", tempDir)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "test-cmd", info.Name)
		assert.Equal(t, "1.0.0", info.Version)
	})

	t.Run("returns error when command not found", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create empty lock file
		lockFile := createBasicLockFile()

		// Write lock file
		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Try to get non-existent command
		info, err := GetCommandInfo("non-existent", tempDir)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), "non-existent")
	})
}

func TestReadLockFile(t *testing.T) {
	t.Run("reads valid lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["cmd1"] = &LockCommand{
			Name:        "cmd1",
			Version:     "1.0.0",
			Source:      "https://github.com/user/cmd1.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		writeLockFileToPath(t, lockPath, lockFile)

		// Read lock file
		result, err := ReadLockFile(lockPath)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "1.0", result.Version)
		assert.Equal(t, 1, result.LockfileVersion)
		assert.Len(t, result.Commands, 1)
		assert.NotNil(t, result.Commands["cmd1"])
	})

	t.Run("handles empty commands", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")

		// Create lock file without commands field
		data := []byte(`version: "1.0"
lockfileVersion: 1`)
		err := os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

		// Read lock file
		result := readLockFileFromPath(t, lockPath)
		assert.NotNil(t, result.Commands)
		assert.Len(t, result.Commands, 0)
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "ccmd-lock.yaml")

		// Write invalid YAML
		err := os.WriteFile(lockPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		// Try to read
		result, err := ReadLockFile(lockPath)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "parse lock file")
	})

	t.Run("returns error when file doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "non-existent.yaml")

		result, err := ReadLockFile(lockPath)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "read lock file")
	})
}

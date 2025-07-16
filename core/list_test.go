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
	"gopkg.in/yaml.v3"
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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:        "test-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/test-cmd.git",
					Resolved:    "https://github.com/user/test-cmd.git@v1.0.0",
					Commit:      "abc123",
					InstalledAt: time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now(),
				},
				"another-cmd": {
					Name:        "another-cmd",
					Version:     "2.0.0",
					Source:      "https://github.com/user/another-cmd.git",
					Resolved:    "https://github.com/user/another-cmd.git@v2.0.0",
					Commit:      "def456",
					InstalledAt: time.Now().Add(-48 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

		// Create directories for commands
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands", "test-cmd"), 0755)
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands", "another-cmd"), 0755)

		// Create .md files
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "commands", "test-cmd.md"), []byte("# test-cmd"), 0644)
		os.WriteFile(filepath.Join(tempDir, ".claude", "commands", "another-cmd.md"), []byte("# another-cmd"), 0644)

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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"broken-cmd": {
					Name:        "broken-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/broken-cmd.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

		// Create .md file but no command directory
		os.MkdirAll(filepath.Join(tempDir, ".claude"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "broken-cmd.md"), []byte("# broken-cmd"), 0644)

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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"broken-cmd": {
					Name:        "broken-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/broken-cmd.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"meta-cmd": {
					Name:        "meta-cmd",
					Version:     "", // Empty in lock file
					Source:      "https://github.com/user/meta-cmd.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

		// Create command directory and files
		cmdDir := filepath.Join(tempDir, ".claude", "commands", "meta-cmd")
		os.MkdirAll(cmdDir, 0755)
		os.MkdirAll(filepath.Join(tempDir, ".claude"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "meta-cmd.md"), []byte("# meta-cmd"), 0644)

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
		metaData, err := yaml.Marshal(&metadata)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), metaData, 0644)
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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:        "test-cmd",
					Version:     "1.0.0",
					Source:      "https://github.com/user/test-cmd.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

		// Create command structure
		os.MkdirAll(filepath.Join(tempDir, ".claude", "commands", "test-cmd"), 0755)
		os.MkdirAll(filepath.Join(tempDir, ".claude"), 0755)
		os.WriteFile(filepath.Join(tempDir, ".claude", "test-cmd.md"), []byte("# test-cmd"), 0644)

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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands:        make(map[string]*LockCommand),
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"cmd1": {
					Name:        "cmd1",
					Version:     "1.0.0",
					Source:      "https://github.com/user/cmd1.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(lockPath, data, 0644)
		require.NoError(t, err)

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
		result, err := ReadLockFile(lockPath)
		require.NoError(t, err)
		require.NotNil(t, result)
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

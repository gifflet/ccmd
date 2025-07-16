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

func TestRemove(t *testing.T) {
	t.Run("removes command successfully", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["test-cmd"] = createTestLockCommand("test-cmd", "1.0.0", "https://github.com/user/test-cmd.git")
		lockFile.Commands["test-cmd"].Resolved = "https://github.com/user/test-cmd.git@v1.0.0"
		lockFile.Commands["keep-cmd"] = &LockCommand{
			Name:        "keep-cmd",
			Version:     "2.0.0",
			Source:      "https://github.com/user/keep-cmd.git",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Write lock file
		writeLockFile(t, lockFile)

		// Create command directories and files
		createCommandStructure(t, "test-cmd")
		createCommandStructure(t, "keep-cmd")

		// Remove command
		err := Remove(RemoveOptions{
			Name:  "test-cmd",
			Force: true,
		})
		require.NoError(t, err)

		// Verify command was removed from lock file
		updatedLock := readLockFile(t)

		assert.Len(t, updatedLock.Commands, 1)
		assert.NotNil(t, updatedLock.Commands["keep-cmd"])
		assert.Nil(t, updatedLock.Commands["test-cmd"])

		// Verify files were removed
		assert.False(t, dirExists(filepath.Join(".claude", "commands", "test-cmd")))
		assert.False(t, fileExists(filepath.Join(".claude", "commands", "test-cmd.md")))

		// Verify other command was not touched
		assert.True(t, dirExists(filepath.Join(".claude", "commands", "keep-cmd")))
		assert.True(t, fileExists(filepath.Join(".claude", "commands", "keep-cmd.md")))
	})

	t.Run("returns error when command not found", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		// Create lock file without the command
		lockFile := createBasicLockFile()
		lockFile.Commands["other-cmd"] = &LockCommand{
			Name: "other-cmd",
		}

		writeLockFile(t, lockFile)

		// Try to remove non-existent command
		err := Remove(RemoveOptions{
			Name: "non-existent",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), "non-existent")
	})

	t.Run("returns error when no commands installed", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		// No lock file exists
		err := Remove(RemoveOptions{
			Name: "any-cmd",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no commands installed")
	})

	t.Run("handles missing command directory gracefully", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["test-cmd"] = &LockCommand{
			Name:   "test-cmd",
			Source: "https://github.com/user/test-cmd.git",
		}

		writeLockFile(t, lockFile)

		// Create only .md file, no command directory
		os.MkdirAll(filepath.Join(".claude", "commands"), 0755)
		os.WriteFile(filepath.Join(".claude", "commands", "test-cmd.md"), []byte("# test-cmd"), 0644)

		// Remove should still work
		err := Remove(RemoveOptions{
			Name:  "test-cmd",
			Force: true,
		})
		require.NoError(t, err)

		// Verify command was removed from lock file
		updatedLock := readLockFile(t)

		assert.Len(t, updatedLock.Commands, 0)
		assert.False(t, fileExists(filepath.Join(".claude", "commands", "test-cmd.md")))
	})

	t.Run("updates ccmd.yaml when requested", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["test-cmd"] = &LockCommand{
			Name:   "test-cmd",
			Source: "https://github.com/user/test-cmd.git",
		}

		writeLockFile(t, lockFile)

		// Create ccmd.yaml with the command
		writeConfig(t, []string{"user/test-cmd@v1.0.0", "user/keep-cmd"})

		// Create command structure
		createCommandStructure(t, "test-cmd")

		// Remove with UpdateFiles
		err := Remove(RemoveOptions{
			Name:        "test-cmd",
			Force:       true,
			UpdateFiles: true,
		})
		require.NoError(t, err)

		// Verify ccmd.yaml was updated
		configData, err := os.ReadFile("ccmd.yaml")
		require.NoError(t, err)
		var updatedConfig map[string]interface{}
		err = yaml.Unmarshal(configData, &updatedConfig)
		require.NoError(t, err)

		commands := updatedConfig["commands"].([]interface{})
		assert.Len(t, commands, 1)
		assert.Equal(t, "user/keep-cmd", commands[0])
	})

	t.Run("requires command name", func(t *testing.T) {
		err := Remove(RemoveOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command name is required")
	})
}

func TestRemoveFromConfig(t *testing.T) {
	t.Run("removes command by repository match", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create config file
		configPath := filepath.Join(tempDir, "ccmd.yaml")
		config := map[string]interface{}{
			"commands": []interface{}{
				"https://github.com/user/test-cmd.git@v1.0.0",
				"user/keep-cmd",
			},
		}
		data, err := yaml.Marshal(config)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Remove command
		err = removeFromConfig(tempDir, "test-cmd", "https://github.com/user/test-cmd.git")
		require.NoError(t, err)

		// Verify result
		data, err = os.ReadFile(configPath)
		require.NoError(t, err)
		var result map[string]interface{}
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err)

		commands := result["commands"].([]interface{})
		assert.Len(t, commands, 1)
		assert.Equal(t, "user/keep-cmd", commands[0])
	})

	t.Run("removes command by name extraction", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create config file
		config := map[string]interface{}{
			"commands": []interface{}{
				"user/my-cmd@v1.0.0",
				"user/keep-cmd",
			},
		}
		data, err := yaml.Marshal(config)
		require.NoError(t, err)
		configPath := filepath.Join(tempDir, "ccmd.yaml")
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Remove command
		err = removeFromConfig(tempDir, "my-cmd", "https://github.com/different/repo.git")
		require.NoError(t, err)

		// Verify result
		data, err = os.ReadFile(configPath)
		require.NoError(t, err)
		var result map[string]interface{}
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err)

		commands := result["commands"].([]interface{})
		assert.Len(t, commands, 1)
		assert.Equal(t, "user/keep-cmd", commands[0])
	})

	t.Run("handles missing config file", func(t *testing.T) {
		tempDir := t.TempDir()

		// No config file exists
		err := removeFromConfig(tempDir, "test-cmd", "repo")
		assert.NoError(t, err) // Should not error
	})

	t.Run("handles config without commands", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create config without commands
		config := map[string]interface{}{
			"other": "value",
		}
		data, err := yaml.Marshal(config)
		require.NoError(t, err)
		configPath := filepath.Join(tempDir, "ccmd.yaml")
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Try to remove
		err = removeFromConfig(tempDir, "test-cmd", "repo")
		assert.NoError(t, err)

		// Config should be unchanged
		data, err = os.ReadFile(configPath)
		require.NoError(t, err)
		var result map[string]interface{}
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "value", result["other"])
	})
}

func TestListCommands(t *testing.T) {
	t.Run("returns command names", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock file
		lockFile := createBasicLockFile()
		lockFile.Commands["cmd1"] = &LockCommand{Name: "cmd1"}
		lockFile.Commands["cmd2"] = &LockCommand{Name: "cmd2"}
		lockFile.Commands["cmd3"] = &LockCommand{Name: "cmd3"}

		writeLockFileToPath(t, filepath.Join(tempDir, "ccmd-lock.yaml"), lockFile)

		// Create command structures
		for _, cmd := range []string{"cmd1", "cmd2", "cmd3"} {
			os.MkdirAll(filepath.Join(tempDir, ".claude", "commands", cmd), 0755)
			os.WriteFile(filepath.Join(tempDir, ".claude", "commands", cmd+".md"), []byte("# "+cmd), 0644)
		}

		// Get command names
		names, err := ListCommands(tempDir)
		require.NoError(t, err)
		assert.Len(t, names, 3)

		// Should be sorted
		assert.Equal(t, []string{"cmd1", "cmd2", "cmd3"}, names)
	})

	t.Run("returns empty list when no commands", func(t *testing.T) {
		tempDir := t.TempDir()

		names, err := ListCommands(tempDir)
		require.NoError(t, err)
		assert.Empty(t, names)
	})
}

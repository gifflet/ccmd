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
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

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
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
				"keep-cmd": {
					Name:        "keep-cmd",
					Version:     "2.0.0",
					Source:      "https://github.com/user/keep-cmd.git",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}

		// Write lock file
		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile("ccmd-lock.yaml", data, 0644)
		require.NoError(t, err)

		// Create command directories and files
		os.MkdirAll(filepath.Join(".claude", "commands", "test-cmd"), 0755)
		os.MkdirAll(filepath.Join(".claude", "commands", "keep-cmd"), 0755)
		os.WriteFile(filepath.Join(".claude", "commands", "test-cmd.md"), []byte("# test-cmd"), 0644)
		os.WriteFile(filepath.Join(".claude", "commands", "keep-cmd.md"), []byte("# keep-cmd"), 0644)

		// Remove command
		err = Remove(RemoveOptions{
			Name:  "test-cmd",
			Force: true,
		})
		require.NoError(t, err)

		// Verify command was removed from lock file
		data, err = os.ReadFile("ccmd-lock.yaml")
		require.NoError(t, err)
		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

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
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// Create lock file without the command
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"other-cmd": {
					Name: "other-cmd",
				},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile("ccmd-lock.yaml", data, 0644)
		require.NoError(t, err)

		// Try to remove non-existent command
		err = Remove(RemoveOptions{
			Name: "non-existent",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), "non-existent")
	})

	t.Run("returns error when no commands installed", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// No lock file exists
		err := Remove(RemoveOptions{
			Name: "any-cmd",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no commands installed")
	})

	t.Run("handles missing command directory gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// Create lock file
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:   "test-cmd",
					Source: "https://github.com/user/test-cmd.git",
				},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile("ccmd-lock.yaml", data, 0644)
		require.NoError(t, err)

		// Create only .md file, no command directory
		os.MkdirAll(filepath.Join(".claude", "commands"), 0755)
		os.WriteFile(filepath.Join(".claude", "commands", "test-cmd.md"), []byte("# test-cmd"), 0644)

		// Remove should still work
		err = Remove(RemoveOptions{
			Name:  "test-cmd",
			Force: true,
		})
		require.NoError(t, err)

		// Verify command was removed from lock file
		data, err = os.ReadFile("ccmd-lock.yaml")
		require.NoError(t, err)
		var updatedLock LockFile
		err = yaml.Unmarshal(data, &updatedLock)
		require.NoError(t, err)

		assert.Len(t, updatedLock.Commands, 0)
		assert.False(t, fileExists(filepath.Join(".claude", "commands", "test-cmd.md")))
	})

	t.Run("updates ccmd.yaml when requested", func(t *testing.T) {
		tempDir := t.TempDir()
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// Create lock file
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"test-cmd": {
					Name:   "test-cmd",
					Source: "https://github.com/user/test-cmd.git",
				},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile("ccmd-lock.yaml", data, 0644)
		require.NoError(t, err)

		// Create ccmd.yaml with the command
		config := map[string]interface{}{
			"commands": []interface{}{
				"user/test-cmd@v1.0.0",
				"user/keep-cmd",
			},
		}
		configData, err := yaml.Marshal(config)
		require.NoError(t, err)
		err = os.WriteFile("ccmd.yaml", configData, 0644)
		require.NoError(t, err)

		// Create command structure
		os.MkdirAll(filepath.Join(".claude", "commands", "test-cmd"), 0755)
		os.WriteFile(filepath.Join(".claude", "commands", "test-cmd.md"), []byte("# test-cmd"), 0644)

		// Remove with UpdateFiles
		err = Remove(RemoveOptions{
			Name:        "test-cmd",
			Force:       true,
			UpdateFiles: true,
		})
		require.NoError(t, err)

		// Verify ccmd.yaml was updated
		configData, err = os.ReadFile("ccmd.yaml")
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
		config := map[string]interface{}{
			"commands": []interface{}{
				"https://github.com/user/test-cmd.git@v1.0.0",
				"user/keep-cmd",
			},
		}
		data, err := yaml.Marshal(config)
		require.NoError(t, err)
		configPath := filepath.Join(tempDir, "ccmd.yaml")
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
		lockFile := LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands: map[string]*LockCommand{
				"cmd1": {Name: "cmd1"},
				"cmd2": {Name: "cmd2"},
				"cmd3": {Name: "cmd3"},
			},
		}

		data, err := yaml.Marshal(&lockFile)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "ccmd-lock.yaml"), data, 0644)
		require.NoError(t, err)

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

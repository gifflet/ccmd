// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	t.Run("NewManager", func(t *testing.T) {
		dir := "/tmp/test"
		m := NewManager(dir)

		if m.rootDir != dir {
			t.Errorf("expected rootDir %s, got %s", dir, m.rootDir)
		}

		if m.ConfigPath() != filepath.Join(dir, ConfigFileName) {
			t.Errorf("unexpected config path: %s", m.ConfigPath())
		}

		if m.LockPath() != filepath.Join(dir, LockFileName) {
			t.Errorf("unexpected lock path: %s", m.LockPath())
		}
	})

	t.Run("InitializeConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Test initialization
		err := m.InitializeConfig()
		if err != nil {
			t.Fatalf("failed to initialize config: %v", err)
		}

		// Check file exists
		if !m.ConfigExists() {
			t.Error("config file should exist after initialization")
		}

		// Test double initialization
		err = m.InitializeConfig()
		if err == nil {
			t.Error("expected error when initializing existing config")
		}

		// Verify content
		config, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		commands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) != 0 {
			t.Errorf("expected empty commands, got %d", len(commands))
		}
	})

	t.Run("InitializeLock", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Test initialization
		err := m.InitializeLock()
		if err != nil {
			t.Fatalf("failed to initialize lock: %v", err)
		}

		// Check file exists
		if !m.LockExists() {
			t.Error("lock file should exist after initialization")
		}

		// Test double initialization
		err = m.InitializeLock()
		if err == nil {
			t.Error("expected error when initializing existing lock")
		}

		// Verify content
		lock, err := m.LoadLockFile()
		if err != nil {
			t.Fatalf("failed to load lock: %v", err)
		}

		if lock.Version != LockFileVersion {
			t.Errorf("expected version %s, got %s", LockFileVersion, lock.Version)
		}

		if len(lock.Commands) != 0 {
			t.Errorf("expected empty commands, got %d", len(lock.Commands))
		}
	})

	t.Run("AddCommand", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Initialize config
		if err := m.InitializeConfig(); err != nil {
			t.Fatalf("failed to initialize config: %v", err)
		}

		// Add command
		err := m.AddCommand("owner/repo", "v1.0.0")
		if err != nil {
			t.Fatalf("failed to add command: %v", err)
		}

		// Verify
		config, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		commands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(commands))
		}

		if commands[0].Repo != "owner/repo" {
			t.Errorf("expected repo owner/repo, got %s", commands[0].Repo)
		}

		if commands[0].Version != "v1.0.0" {
			t.Errorf("expected version v1.0.0, got %s", commands[0].Version)
		}

		// Test duplicate
		err = m.AddCommand("owner/repo", "v2.0.0")
		if err == nil {
			t.Error("expected error when adding duplicate command")
		}

		// Test invalid repo format
		err = m.AddCommand("invalid-repo", "v1.0.0")
		if err == nil {
			t.Error("expected error when adding invalid repo format")
		}
	})

	t.Run("RemoveCommand", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Initialize and add commands
		if err := m.InitializeConfig(); err != nil {
			t.Fatalf("failed to initialize config: %v", err)
		}

		if err := m.AddCommand("owner/repo1", "v1.0.0"); err != nil {
			t.Fatalf("failed to add command 1: %v", err)
		}

		if err := m.AddCommand("owner/repo2", "v2.0.0"); err != nil {
			t.Fatalf("failed to add command 2: %v", err)
		}

		// Remove command
		err := m.RemoveCommand("owner/repo1")
		if err != nil {
			t.Fatalf("failed to remove command: %v", err)
		}

		// Verify
		config, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		commands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(commands))
		}

		if commands[0].Repo != "owner/repo2" {
			t.Errorf("expected remaining repo to be owner/repo2, got %s", commands[0].Repo)
		}

		// Test removing non-existent
		err = m.RemoveCommand("owner/nonexistent")
		if err == nil {
			t.Error("expected error when removing non-existent command")
		}
	})

	t.Run("UpdateCommand", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Initialize and add command
		if err := m.InitializeConfig(); err != nil {
			t.Fatalf("failed to initialize config: %v", err)
		}

		if err := m.AddCommand("owner/repo", "v1.0.0"); err != nil {
			t.Fatalf("failed to add command: %v", err)
		}

		// Update command
		err := m.UpdateCommand("owner/repo", "v2.0.0")
		if err != nil {
			t.Fatalf("failed to update command: %v", err)
		}

		// Verify
		config, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		commands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) > 0 && commands[0].Version != "v2.0.0" {
			t.Errorf("expected version v2.0.0, got %s", commands[0].Version)
		}

		// Test updating non-existent
		err = m.UpdateCommand("owner/nonexistent", "v3.0.0")
		if err == nil {
			t.Error("expected error when updating non-existent command")
		}

		// Test empty version (which is valid - defaults to latest)
		err = m.UpdateCommand("owner/repo", "")
		if err != nil {
			t.Errorf("empty version should be valid: %v", err)
		}

		// Verify empty version was set
		config, err = m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		commands, err = config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) > 0 && commands[0].Version != "" {
			t.Errorf("expected empty version, got %s", commands[0].Version)
		}
	})

	t.Run("SaveAndLoadOperations", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Test Config operations
		config := &Config{
			Commands: []string{
				"owner/repo1@v1.0.0",
				"owner/repo2@latest",
			},
		}

		err := m.SaveConfig(config)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		loadedConfig, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		loadedCommands, err := loadedConfig.GetCommands()
		if err != nil {
			t.Fatalf("failed to get loaded commands: %v", err)
		}
		if len(loadedCommands) != 2 {
			t.Errorf("expected 2 commands, got %d", len(loadedCommands))
		}

		// Test LockFile operations
		lockFile := NewLockFile()
		cmd := &CommandLockInfo{
			Name:         "repo1",
			Source:       "owner/repo1",
			Version:      "v1.0.0",
			Resolved:     "owner/repo1@v1.0.0",
			Commit:       "1234567890abcdef1234567890abcdef12345678",
			InstalledAt:  time.Now(),
			UpdatedAt:    time.Now(),
			Dependencies: []string{},
			Metadata:     map[string]string{},
		}

		if err := lockFile.AddCommand(cmd); err != nil {
			t.Fatalf("failed to add command to lock: %v", err)
		}

		err = m.SaveLockFile(lockFile)
		if err != nil {
			t.Fatalf("failed to save lock file: %v", err)
		}

		loadedLock, err := m.LoadLockFile()
		if err != nil {
			t.Fatalf("failed to load lock file: %v", err)
		}

		if len(loadedLock.Commands) != 1 {
			t.Errorf("expected 1 command in lock, got %d", len(loadedLock.Commands))
		}

		if cmd, exists := loadedLock.GetCommand("repo1"); !exists {
			t.Error("expected to find repo1 in lock file")
		} else {
			if cmd.Source != "owner/repo1" {
				t.Errorf("expected source owner/repo1, got %s", cmd.Source)
			}
		}
	})

	t.Run("FileNotExist", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Test loading non-existent files
		_, err := m.LoadConfig()
		if err == nil {
			t.Error("expected error when loading non-existent config")
		}

		_, err = m.LoadLockFile()
		if err == nil {
			t.Error("expected error when loading non-existent lock file")
		}

		// Test operations on non-existent config
		// AddCommand should create a minimal config when it doesn't exist
		err = m.AddCommand("owner/repo", "v1.0.0")
		if err != nil {
			t.Errorf("unexpected error when adding to non-existent config: %v", err)
		}

		// Verify the config was created with minimal structure
		config, err := m.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load created config: %v", err)
		}

		// Check that only commands field is present
		commands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(commands))
		}

		// RemoveCommand should now work since config exists
		err = m.RemoveCommand("owner/repo")
		if err != nil {
			t.Errorf("unexpected error when removing from config: %v", err)
		}

		// Clean up for next test
		if err := m.fs.Remove(m.ConfigPath()); err != nil {
			t.Fatalf("failed to remove config: %v", err)
		}

		// Now test RemoveCommand on non-existent config
		err = m.RemoveCommand("owner/repo")
		if err == nil {
			t.Error("expected error when removing from non-existent config")
		}

		err = m.UpdateCommand("owner/repo", "v2.0.0")
		if err == nil {
			t.Error("expected error when updating non-existent config")
		}
	})

	t.Run("CommandExists", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		// Test command doesn't exist
		exists, err := m.CommandExists("test-cmd")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected command to not exist")
		}

		// Create command directory
		commandPath := filepath.Join(tmpDir, "commands", "test-cmd")
		if err := os.MkdirAll(commandPath, 0755); err != nil {
			t.Fatalf("failed to create command directory: %v", err)
		}

		// Test command exists
		exists, err = m.CommandExists("test-cmd")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected command to exist")
		}
	})
}

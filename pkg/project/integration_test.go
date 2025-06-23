/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
)

func TestIntegration(t *testing.T) {
	t.Run("CompleteWorkflow", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(tmpDir)

		// Step 1: Initialize project
		err := manager.InitializeConfig()
		if err != nil {
			t.Fatalf("failed to initialize config: %v", err)
		}

		err = manager.InitializeLock()
		if err != nil {
			t.Fatalf("failed to initialize lock: %v", err)
		}

		// Step 2: Add commands to config
		commands := []struct {
			repo    string
			version string
		}{
			{"github/cli", "v2.0.0"},
			{"junegunn/fzf", "latest"},
			{"evilmartians/lefthook", "v1.5.0"},
		}

		for _, cmd := range commands {
			err = manager.AddCommand(cmd.repo, cmd.version)
			if err != nil {
				t.Fatalf("failed to add command %s: %v", cmd.repo, err)
			}
		}

		// Step 3: Verify config
		config, err := manager.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		configCommands, err := config.GetCommands()
		if err != nil {
			t.Fatalf("failed to get commands: %v", err)
		}
		if len(configCommands) != 3 {
			t.Errorf("expected 3 commands, got %d", len(configCommands))
		}

		// Step 4: Add commands to lock file (simulating installation)
		lockFile, err := manager.LoadLockFile()
		if err != nil {
			t.Fatalf("failed to load lock file: %v", err)
		}

		installedCommands := []*CommandLockInfo{
			{
				Name:         "gh",
				Source:       "github/cli",
				Version:      "v2.0.0",
				Resolved:     "github/cli@v2.0.0",
				Commit:       "1234567890abcdef1234567890abcdef12345678",
				InstalledAt:  time.Now(),
				UpdatedAt:    time.Now(),
				Dependencies: []string{},
				Metadata: map[string]string{
					"description": "GitHub CLI",
				},
			},
			{
				Name:         "fzf",
				Source:       "junegunn/fzf",
				Version:      "latest",
				Resolved:     "junegunn/fzf@latest",
				Commit:       "2345678901bcdefg2345678901bcdefg23456789",
				InstalledAt:  time.Now(),
				UpdatedAt:    time.Now(),
				Dependencies: []string{},
				Metadata: map[string]string{
					"description": "Fuzzy finder",
				},
			},
		}

		for _, cmd := range installedCommands {
			err = lockFile.AddCommand(cmd)
			if err != nil {
				t.Fatalf("failed to add command to lock: %v", err)
			}
		}

		err = manager.SaveLockFile(lockFile)
		if err != nil {
			t.Fatalf("failed to save lock file: %v", err)
		}

		// Step 5: Update a command version
		err = manager.UpdateCommand("github/cli", "v2.1.0")
		if err != nil {
			t.Fatalf("failed to update command: %v", err)
		}

		// Step 6: Remove a command
		err = manager.RemoveCommand("evilmartians/lefthook")
		if err != nil {
			t.Fatalf("failed to remove command: %v", err)
		}

		// Step 7: Final verification
		finalConfig, err := manager.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load final config: %v", err)
		}

		finalCommands, err := finalConfig.GetCommands()
		if err != nil {
			t.Fatalf("failed to get final commands: %v", err)
		}
		if len(finalCommands) != 2 {
			t.Errorf("expected 2 commands after removal, got %d", len(finalCommands))
		}

		// Check updated version
		found := false
		for _, cmd := range finalCommands {
			if cmd.Repo == "github/cli" {
				found = true
				if cmd.Version != "v2.1.0" {
					t.Errorf("expected version v2.1.0, got %s", cmd.Version)
				}
				break
			}
		}

		if !found {
			t.Error("github/cli not found in final config")
		}

		// Verify lock file still has entries
		finalLock, err := manager.LoadLockFile()
		if err != nil {
			t.Fatalf("failed to load final lock: %v", err)
		}

		if len(finalLock.Commands) != 2 {
			t.Errorf("expected 2 commands in lock, got %d", len(finalLock.Commands))
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create initial config
		config := &Config{
			Commands: []string{"owner/repo1@v1.0.0"},
		}

		configPath := filepath.Join(tmpDir, ConfigFileName)
		fileSystem := fs.OS{}
		err := SaveConfig(config, configPath, fileSystem)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		// Simulate concurrent reads and writes
		done := make(chan bool, 3)
		errors := make(chan error, 3)

		// Reader 1
		go func() {
			defer func() { done <- true }()
			for i := 0; i < 10; i++ {
				_, err := LoadConfig(configPath, fileSystem)
				if err != nil {
					errors <- err
					return
				}
				time.Sleep(time.Millisecond)
			}
		}()

		// Reader 2
		go func() {
			defer func() { done <- true }()
			for i := 0; i < 10; i++ {
				_, err := LoadConfig(configPath, fileSystem)
				if err != nil {
					errors <- err
					return
				}
				time.Sleep(time.Millisecond)
			}
		}()

		// Writer
		go func() {
			defer func() { done <- true }()
			for i := 0; i < 5; i++ {
				config.Commands = []string{"owner/repo1@v1.0." + string(rune('0'+i))}
				err := SaveConfig(config, configPath, fileSystem)
				if err != nil {
					errors <- err
					return
				}
				time.Sleep(2 * time.Millisecond)
			}
		}()

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}

		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("concurrent operation failed: %v", err)
		}

		// Verify final state is valid
		finalConfig, err := LoadConfig(configPath, fileSystem)
		if err != nil {
			t.Fatalf("failed to load final config: %v", err)
		}

		finalCommands, err := finalConfig.GetCommands()
		if err != nil {
			t.Fatalf("failed to get final commands: %v", err)
		}
		if len(finalCommands) != 1 {
			t.Errorf("expected 1 command, got %d", len(finalCommands))
		}
	})

	t.Run("InvalidFileRecovery", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(tmpDir)

		// Create invalid YAML file
		configPath := manager.ConfigPath()
		err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
		if err != nil {
			t.Fatalf("failed to write invalid file: %v", err)
		}

		// Try to load - should fail
		_, err = manager.LoadConfig()
		if err == nil {
			t.Error("expected error when loading invalid YAML")
		}

		// Overwrite with valid config
		validConfig := &Config{
			Commands: []string{"owner/repo@v1.0.0"},
		}

		err = manager.SaveConfig(validConfig)
		if err != nil {
			t.Fatalf("failed to save valid config: %v", err)
		}

		// Should be able to load now
		loaded, err := manager.LoadConfig()
		if err != nil {
			t.Fatalf("failed to load after recovery: %v", err)
		}

		loadedCommands, err := loaded.GetCommands()
		if err != nil {
			t.Fatalf("failed to get loaded commands: %v", err)
		}
		if len(loadedCommands) != 1 {
			t.Errorf("expected 1 command after recovery, got %d", len(loadedCommands))
		}
	})
}

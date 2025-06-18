package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

		if len(config.Commands) != 3 {
			t.Errorf("expected 3 commands, got %d", len(config.Commands))
		}

		// Step 4: Add commands to lock file (simulating installation)
		lockFile, err := manager.LoadLockFile()
		if err != nil {
			t.Fatalf("failed to load lock file: %v", err)
		}

		installedCommands := []*Command{
			{
				Name:         "gh",
				Repository:   "github/cli",
				Version:      "v2.0.0",
				CommitHash:   "1234567890abcdef1234567890abcdef12345678",
				InstalledAt:  time.Now(),
				UpdatedAt:    time.Now(),
				FileSize:     10485760, // 10MB
				Checksum:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				Dependencies: []string{},
				Metadata: map[string]string{
					"description": "GitHub CLI",
				},
			},
			{
				Name:         "fzf",
				Repository:   "junegunn/fzf",
				Version:      "latest",
				CommitHash:   "2345678901bcdefg2345678901bcdefg23456789",
				InstalledAt:  time.Now(),
				UpdatedAt:    time.Now(),
				FileSize:     2097152, // 2MB
				Checksum:     "bcdefg2345678901bcdefg2345678901bcdefg2345678901bcdefg2345678901",
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

		if len(finalConfig.Commands) != 2 {
			t.Errorf("expected 2 commands after removal, got %d", len(finalConfig.Commands))
		}

		// Check updated version
		found := false
		for _, cmd := range finalConfig.Commands {
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
			Commands: []ConfigCommand{
				{Repo: "owner/repo1", Version: "v1.0.0"},
			},
		}

		configPath := filepath.Join(tmpDir, ConfigFileName)
		err := SaveConfig(config, configPath)
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
				_, err := LoadConfig(configPath)
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
				_, err := LoadConfig(configPath)
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
				config.Commands[0].Version = "v1.0." + string(rune('0'+i))
				err := SaveConfig(config, configPath)
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
		finalConfig, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load final config: %v", err)
		}

		if len(finalConfig.Commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(finalConfig.Commands))
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
			Commands: []ConfigCommand{
				{Repo: "owner/repo", Version: "v1.0.0"},
			},
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

		if len(loaded.Commands) != 1 {
			t.Errorf("expected 1 command after recovery, got %d", len(loaded.Commands))
		}
	})
}

package project

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ccmd.yaml")

		config := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo1", Version: "v1.0.0"},
				{Repo: "owner/repo2", Version: "latest"},
			},
		}

		err := SaveConfig(config, configPath)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Load and verify
		loaded, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}

		if len(loaded.Commands) != 2 {
			t.Errorf("expected 2 commands, got %d", len(loaded.Commands))
		}

		if loaded.Commands[0].Repo != "owner/repo1" {
			t.Errorf("expected first repo to be owner/repo1, got %s", loaded.Commands[0].Repo)
		}

		if loaded.Commands[1].Version != "latest" {
			t.Errorf("expected second version to be latest, got %s", loaded.Commands[1].Version)
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ccmd.yaml")

		// Empty commands is now valid
		config := &Config{
			Commands: []ConfigCommand{},
		}

		err := SaveConfig(config, configPath)
		if err != nil {
			t.Errorf("empty config should be valid: %v", err)
		}

		// Invalid command
		config = &Config{
			Commands: []ConfigCommand{
				{Repo: "", Version: "v1.0.0"},
			},
		}

		err = SaveConfig(config, configPath)
		if err == nil {
			t.Error("expected error when saving config with invalid command")
		}
	})

	t.Run("AtomicWrite", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ccmd.yaml")

		// Create initial config
		config1 := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo1", Version: "v1.0.0"},
			},
		}

		err := SaveConfig(config1, configPath)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		// Save new config
		config2 := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo2", Version: "v2.0.0"},
			},
		}

		err = SaveConfig(config2, configPath)
		if err != nil {
			t.Fatalf("failed to save updated config: %v", err)
		}

		// Verify temp file doesn't exist
		tempPath := configPath + ".tmp"
		if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
			t.Error("temporary file should not exist after save")
		}

		// Verify final content
		loaded, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if loaded.Commands[0].Repo != "owner/repo2" {
			t.Errorf("expected repo owner/repo2, got %s", loaded.Commands[0].Repo)
		}
	})

	t.Run("FilePermissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ccmd.yaml")

		config := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo", Version: "v1.0.0"},
			},
		}

		err := SaveConfig(config, configPath)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Check file permissions
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("failed to stat config file: %v", err)
		}

		perm := info.Mode().Perm()
		// Should be readable by all, writable by owner (0644)
		if perm != 0644 {
			t.Errorf("expected permissions 0644, got %#o", perm)
		}
	})
}

func TestWriteConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo1", Version: "v1.0.0"},
				{Repo: "owner/repo2", Version: "latest"},
			},
		}

		var buf bytes.Buffer
		err := WriteConfig(config, &buf)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "owner/repo1") {
			t.Error("output should contain owner/repo1")
		}
		if !strings.Contains(output, "v1.0.0") {
			t.Error("output should contain v1.0.0")
		}
		if !strings.Contains(output, "owner/repo2") {
			t.Error("output should contain owner/repo2")
		}
		if !strings.Contains(output, "latest") {
			t.Error("output should contain latest")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		// Empty commands is now valid
		config := &Config{
			Commands: []ConfigCommand{},
		}

		var buf bytes.Buffer
		err := WriteConfig(config, &buf)
		if err != nil {
			t.Errorf("empty config should be valid: %v", err)
		}

		// Invalid command
		config = &Config{
			Commands: []ConfigCommand{
				{Repo: "invalid", Version: "v1.0.0"},
			},
		}

		buf.Reset()
		err = WriteConfig(config, &buf)
		if err == nil {
			t.Error("expected error when writing config with invalid command")
		}
	})

	t.Run("YAMLFormat", func(t *testing.T) {
		config := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo", Version: "v1.0.0"},
			},
		}

		var buf bytes.Buffer
		err := WriteConfig(config, &buf)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Parse the output to ensure it's valid YAML
		parsedConfig, err := ParseConfig(&buf)
		if err != nil {
			t.Fatalf("failed to parse written config: %v", err)
		}

		if len(parsedConfig.Commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(parsedConfig.Commands))
		}

		if parsedConfig.Commands[0].Repo != "owner/repo" {
			t.Errorf("expected repo owner/repo, got %s", parsedConfig.Commands[0].Repo)
		}
	})

	t.Run("EmptyVersion", func(t *testing.T) {
		config := &Config{
			Commands: []ConfigCommand{
				{Repo: "owner/repo", Version: ""},
			},
		}

		var buf bytes.Buffer
		err := WriteConfig(config, &buf)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		output := buf.String()
		// Version field should be omitted when empty due to omitempty tag
		if strings.Contains(output, "version:") {
			t.Error("output should not contain version field when empty")
		}
	})
}

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
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
)

const (
	// ConfigFileName is the default name for the configuration file
	ConfigFileName = "ccmd.yaml"
	// LockFileName is the default name for the lock file
	LockFileName = "ccmd-lock.yaml"
)

// Manager provides high-level operations for project file management
type Manager struct {
	rootDir string
	fs      fs.FileSystem
}

// NewManager creates a new project manager for the given directory
func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir: rootDir,
		fs:      fs.OS{},
	}
}

// NewManagerWithFS creates a new project manager with a custom filesystem
func NewManagerWithFS(rootDir string, fileSystem fs.FileSystem) *Manager {
	return &Manager{
		rootDir: rootDir,
		fs:      fileSystem,
	}
}

// ConfigPath returns the full path to the config file
func (m *Manager) ConfigPath() string {
	return filepath.Join(m.rootDir, ConfigFileName)
}

// LockPath returns the full path to the lock file
func (m *Manager) LockPath() string {
	return filepath.Join(m.rootDir, LockFileName)
}

// LoadConfig loads the project configuration file
func (m *Manager) LoadConfig() (*Config, error) {
	return LoadConfig(m.ConfigPath(), m.fs)
}

// SaveConfig saves the project configuration file
func (m *Manager) SaveConfig(config *Config) error {
	return SaveConfig(config, m.ConfigPath(), m.fs)
}

// LoadLockFile loads the project lock file
func (m *Manager) LoadLockFile() (*LockFile, error) {
	return LoadFromFile(m.LockPath(), m.fs)
}

// SaveLockFile saves the project lock file
func (m *Manager) SaveLockFile(lockFile *LockFile) error {
	return lockFile.SaveToFile(m.LockPath(), m.fs)
}

// ConfigExists checks if the config file exists
func (m *Manager) ConfigExists() bool {
	_, err := m.fs.Stat(m.ConfigPath())
	return err == nil
}

// LockExists checks if the lock file exists
func (m *Manager) LockExists() bool {
	_, err := m.fs.Stat(m.LockPath())
	return err == nil
}

// InitializeConfig creates a new config file with default values
func (m *Manager) InitializeConfig() error {
	if m.ConfigExists() {
		return fmt.Errorf("config file already exists at %s", m.ConfigPath())
	}

	config := &Config{
		Commands: []ConfigCommand{},
	}

	return m.SaveConfig(config)
}

// InitializeLock creates a new lock file
func (m *Manager) InitializeLock() error {
	if m.LockExists() {
		return fmt.Errorf("lock file already exists at %s", m.LockPath())
	}

	lockFile := NewLockFile()
	return m.SaveLockFile(lockFile)
}

// AddCommand adds a new command to the configuration
func (m *Manager) AddCommand(repo, version string) error {
	// Validate repo format first
	if err := validateRepoFormat(repo); err != nil {
		return fmt.Errorf("invalid repo format: %w", err)
	}

	// If config doesn't exist, create minimal config with just commands
	if !m.ConfigExists() {
		var newCmd string
		if version != "" {
			newCmd = repo + "@" + version
		} else {
			newCmd = repo
		}

		config := &Config{
			Commands: []string{newCmd},
		}
		return m.SaveConfig(config)
	}

	// If config exists, preserve existing structure
	return m.preserveAndUpdateCommands(func(commands []ConfigCommand) ([]ConfigCommand, error) {
		// Check if command already exists
		for _, existing := range commands {
			if existing.Repo == repo {
				return nil, fmt.Errorf("command %s already exists in configuration", repo)
			}
		}

		// Add new command
		commands = append(commands, ConfigCommand{
			Repo:    repo,
			Version: version,
		})
		return commands, nil
	})
}

// RemoveCommand removes a command from the configuration
func (m *Manager) RemoveCommand(repo string) error {
	// Use preserve and update to maintain existing structure
	return m.preserveAndUpdateCommands(func(commands []ConfigCommand) ([]ConfigCommand, error) {
		found := false
		newCommands := make([]ConfigCommand, 0, len(commands))

		for _, cmd := range commands {
			if cmd.Repo != repo {
				newCommands = append(newCommands, cmd)
			} else {
				found = true
			}
		}

		if !found {
			return nil, fmt.Errorf("command %s not found in configuration", repo)
		}

		return newCommands, nil
	})
}

// UpdateCommandInLockFile updates command information in the lock file
func (m *Manager) UpdateCommandInLockFile(cmd *Command) error {
	lockFile, err := m.LoadLockFile()
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	if err := lockFile.AddCommand(cmd); err != nil {
		return fmt.Errorf("failed to add command to lock file: %w", err)
	}

	return m.SaveLockFile(lockFile)
}

// UpdateCommand updates a command's version in the configuration
func (m *Manager) UpdateCommand(repo, newVersion string) error {
	// Use preserve and update to maintain existing structure
	return m.preserveAndUpdateCommands(func(commands []ConfigCommand) ([]ConfigCommand, error) {
		found := false
		updatedCommands := make([]ConfigCommand, len(commands))

		for i, cmd := range commands {
			if cmd.Repo == repo {
				found = true
				updatedCommands[i] = ConfigCommand{
					Repo:    repo,
					Version: newVersion,
				}
			} else {
				updatedCommands[i] = cmd
			}
		}

		if !found {
			return nil, fmt.Errorf("command %s not found in configuration", repo)
		}

		return updatedCommands, nil
	})
}

// CommandExists checks if a command is installed in the project
func (m *Manager) CommandExists(name string) (bool, error) {
	commandPath := filepath.Join(m.rootDir, "commands", name)
	return m.fs.Exists(commandPath)
}

// Sync ensures the lock file is in sync with the configuration
// This is a placeholder for future implementation that would:
// 1. Check all commands in config are in lock file
// 2. Remove commands from lock that are not in config
// 3. Update versions as needed
func (m *Manager) Sync() error {
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var lockFile *LockFile
	if m.LockExists() {
		lockFile, err = m.LoadLockFile()
		if err != nil {
			return fmt.Errorf("failed to load lock file: %w", err)
		}
	} else {
		lockFile = NewLockFile()
	}

	// This is a placeholder - actual sync logic would be implemented
	// by the installation/update system
	_ = config
	_ = lockFile

	return nil
}

// LoadLegacyLockFile loads commands.lock JSON file for migration
func (m *Manager) LoadLegacyLockFile() (*LockFile, error) {
	lockPath := filepath.Join(filepath.Dir(m.rootDir), ".claude", "commands.lock")

	// Check if legacy file exists
	exists, err := m.fs.Exists(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check legacy lock file: %w", err)
	}
	if !exists {
		return nil, nil
	}

	// Read JSON file
	data, err := m.fs.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read legacy lock file: %w", err)
	}

	// Parse JSON structure
	var legacyLock struct {
		Version  string                 `json:"version"`
		Commands map[string]interface{} `json:"commands"`
	}

	if err := json.Unmarshal(data, &legacyLock); err != nil {
		return nil, fmt.Errorf("failed to parse legacy lock file: %w", err)
	}

	// Convert to new format
	lockFile := NewLockFile()
	for name, cmdData := range legacyLock.Commands {
		cmdMap, ok := cmdData.(map[string]interface{})
		if !ok {
			continue
		}

		cmd := &CommandLockInfo{
			Name:    name,
			Version: getString(cmdMap, "version"),
			Source:  getString(cmdMap, "source"),
		}

		// Parse timestamps
		if installedAt := getString(cmdMap, "installed_at"); installedAt != "" {
			if t, err := time.Parse(time.RFC3339, installedAt); err == nil {
				cmd.InstalledAt = t
			}
		}
		if updatedAt := getString(cmdMap, "updated_at"); updatedAt != "" {
			if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
				cmd.UpdatedAt = t
			}
		}

		// Generate resolved field
		if cmd.Version != "" && cmd.Source != "" {
			cmd.Resolved = fmt.Sprintf("%s@%s", cmd.Source, cmd.Version)
		}

		lockFile.Commands[name] = cmd
	}

	return lockFile, nil
}

// MigrateLegacyLockFile migrates from commands.lock to ccmd-lock.yaml
func (m *Manager) MigrateLegacyLockFile() error {
	// Load legacy lock file
	lockFile, err := m.LoadLegacyLockFile()
	if err != nil {
		return fmt.Errorf("failed to load legacy lock file: %w", err)
	}

	// No legacy file found
	if lockFile == nil {
		return nil
	}

	// Save as new format
	if err := m.SaveLockFile(lockFile); err != nil {
		return fmt.Errorf("failed to save migrated lock file: %w", err)
	}

	// Remove legacy file
	lockPath := filepath.Join(filepath.Dir(m.rootDir), ".claude", "commands.lock")
	if err := m.fs.Remove(lockPath); err != nil {
		// Not critical if we can't remove
		return nil
	}

	return nil
}

// preserveAndUpdateCommands preserves the existing YAML structure while updating only commands
func (m *Manager) preserveAndUpdateCommands(updateFunc func([]ConfigCommand) ([]ConfigCommand, error)) error {
	// Read the raw YAML data
	rawData, err := m.fs.ReadFile(m.ConfigPath())
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a Node to preserve structure and order
	var doc yaml.Node
	if err := yaml.Unmarshal(rawData, &doc); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Load the config normally to get current commands
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current commands
	commands, err := config.GetCommands()
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	// Apply the update function
	updatedCommands, err := updateFunc(commands)
	if err != nil {
		return err
	}

	// Convert updated commands to string format
	stringCommands := make([]string, len(updatedCommands))
	for i, cmd := range updatedCommands {
		if cmd.Version != "" {
			stringCommands[i] = cmd.Repo + "@" + cmd.Version
		} else {
			stringCommands[i] = cmd.Repo
		}
	}
	commandsValue := stringCommands

	// Update the commands field in the YAML node
	if err := updateYAMLNodeField(&doc, "commands", commandsValue); err != nil {
		return fmt.Errorf("failed to update commands field: %w", err)
	}

	// Marshal back to YAML preserving order
	updatedData, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write the file
	return m.fs.WriteFile(m.ConfigPath(), updatedData, 0o644)
}

// updateYAMLNodeField updates or adds a field in a YAML node preserving order
func updateYAMLNodeField(node *yaml.Node, fieldName string, value interface{}) error {
	// Ensure we have a document node
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return fmt.Errorf("invalid YAML document structure")
	}

	// Get the root mapping node
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("root node is not a mapping")
	}

	// Look for existing field
	for i := 0; i < len(root.Content); i += 2 {
		keyNode := root.Content[i]
		if keyNode.Value == fieldName {
			// Field exists, update its value
			// Create new value node
			newValueNode := &yaml.Node{}
			if err := newValueNode.Encode(value); err != nil {
				return fmt.Errorf("failed to encode value: %w", err)
			}

			// Replace the value node
			root.Content[i+1] = newValueNode
			return nil
		}
	}

	// Field doesn't exist, add it at the end
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: fieldName,
	}

	valueNode := &yaml.Node{}
	if err := valueNode.Encode(value); err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	root.Content = append(root.Content, keyNode, valueNode)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

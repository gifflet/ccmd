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
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/errors"
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
		return errors.AlreadyExists(fmt.Sprintf("config file at %s", m.ConfigPath()))
	}

	config := &Config{
		Commands: []ConfigCommand{},
	}

	return m.SaveConfig(config)
}

// InitializeLock creates a new lock file
func (m *Manager) InitializeLock() error {
	if m.LockExists() {
		return errors.AlreadyExists(fmt.Sprintf("lock file at %s", m.LockPath()))
	}

	lockFile := NewLockFile()
	return m.SaveLockFile(lockFile)
}

// AddCommand adds a new command to the configuration
func (m *Manager) AddCommand(repo, version string) error {
	// Validate repo format first
	if err := validateRepoFormat(repo); err != nil {
		return errors.InvalidInput(fmt.Sprintf("invalid repo format: %v", err))
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
				return nil, errors.AlreadyExists(fmt.Sprintf("command %s in configuration", repo))
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
			return nil, errors.NotFound(fmt.Sprintf("command %s in configuration", repo))
		}

		return newCommands, nil
	})
}

// UpdateCommandInLockFile updates command information in the lock file
func (m *Manager) UpdateCommandInLockFile(cmd *Command) error {
	lockFile, err := m.LoadLockFile()
	if err != nil {
		return err
	}

	if err := lockFile.AddCommand(cmd); err != nil {
		return err
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
			return nil, errors.NotFound(fmt.Sprintf("command %s in configuration", repo))
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
		return err
	}

	var lockFile *LockFile
	if m.LockExists() {
		lockFile, err = m.LoadLockFile()
		if err != nil {
			return err
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

// preserveAndUpdateCommands preserves the existing YAML structure while updating only commands
func (m *Manager) preserveAndUpdateCommands(updateFunc func([]ConfigCommand) ([]ConfigCommand, error)) error {
	// Read the raw YAML data
	rawData, err := m.fs.ReadFile(m.ConfigPath())
	if err != nil {
		return errors.FileError("read config file", m.ConfigPath(), err)
	}

	// Parse YAML into a Node to preserve structure and order
	var doc yaml.Node
	if err := yaml.Unmarshal(rawData, &doc); err != nil {
		return errors.FileError("parse YAML", m.ConfigPath(), err)
	}

	// Load the config normally to get current commands
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	// Get current commands
	commands, err := config.GetCommands()
	if err != nil {
		return err
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
		return err
	}

	// Marshal back to YAML preserving order
	updatedData, err := yaml.Marshal(&doc)
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to marshal updated config: %v", err))
	}

	// Write the file
	return m.fs.WriteFile(m.ConfigPath(), updatedData, 0o644)
}

// updateYAMLNodeField updates or adds a field in a YAML node preserving order
func updateYAMLNodeField(node *yaml.Node, fieldName string, value interface{}) error {
	// Ensure we have a document node
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return errors.InvalidInput("invalid YAML document structure")
	}

	// Get the root mapping node
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return errors.InvalidInput("root node is not a mapping")
	}

	// Look for existing field
	for i := 0; i < len(root.Content); i += 2 {
		keyNode := root.Content[i]
		if keyNode.Value == fieldName {
			// Field exists, update its value
			// Create new value node
			newValueNode := &yaml.Node{}
			if err := newValueNode.Encode(value); err != nil {
				return errors.InvalidInput(fmt.Sprintf("failed to encode value: %v", err))
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
		return errors.InvalidInput(fmt.Sprintf("failed to encode value: %v", err))
	}

	root.Content = append(root.Content, keyNode, valueNode)
	return nil
}

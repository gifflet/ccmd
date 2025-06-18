package project

import (
	"fmt"
	"os"
	"path/filepath"
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
}

// NewManager creates a new project manager for the given directory
func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir: rootDir,
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
	return LoadConfig(m.ConfigPath())
}

// SaveConfig saves the project configuration file
func (m *Manager) SaveConfig(config *Config) error {
	return SaveConfig(config, m.ConfigPath())
}

// LoadLockFile loads the project lock file
func (m *Manager) LoadLockFile() (*LockFile, error) {
	return LoadFromFile(m.LockPath())
}

// SaveLockFile saves the project lock file
func (m *Manager) SaveLockFile(lockFile *LockFile) error {
	return lockFile.SaveToFile(m.LockPath())
}

// ConfigExists checks if the config file exists
func (m *Manager) ConfigExists() bool {
	_, err := os.Stat(m.ConfigPath())
	return err == nil
}

// LockExists checks if the lock file exists
func (m *Manager) LockExists() bool {
	_, err := os.Stat(m.LockPath())
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
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cmd := ConfigCommand{
		Repo:    repo,
		Version: version,
	}

	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Check if command already exists
	for _, existing := range config.Commands {
		if existing.Repo == repo {
			return fmt.Errorf("command %s already exists in configuration", repo)
		}
	}

	config.Commands = append(config.Commands, cmd)
	return m.SaveConfig(config)
}

// RemoveCommand removes a command from the configuration
func (m *Manager) RemoveCommand(repo string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	found := false
	newCommands := make([]ConfigCommand, 0, len(config.Commands))
	for _, cmd := range config.Commands {
		if cmd.Repo != repo {
			newCommands = append(newCommands, cmd)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("command %s not found in configuration", repo)
	}

	config.Commands = newCommands
	return m.SaveConfig(config)
}

// UpdateCommand updates a command's version in the configuration
func (m *Manager) UpdateCommand(repo, newVersion string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	found := false
	for i, cmd := range config.Commands {
		if cmd.Repo == repo {
			config.Commands[i].Version = newVersion
			if err := config.Commands[i].Validate(); err != nil {
				return fmt.Errorf("invalid version: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("command %s not found in configuration", repo)
	}

	return m.SaveConfig(config)
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

	lockFile, err := m.LoadLockFile()
	if err != nil {
		if os.IsNotExist(err) {
			lockFile = NewLockFile()
		} else {
			return fmt.Errorf("failed to load lock file: %w", err)
		}
	}

	// This is a placeholder - actual sync logic would be implemented
	// by the installation/update system
	_ = config
	_ = lockFile

	return nil
}

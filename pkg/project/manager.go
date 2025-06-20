package project

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

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
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current commands
	commands, err := config.GetCommands()
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	// Check if command already exists
	for _, existing := range commands {
		if existing.Repo == repo {
			return fmt.Errorf("command %s already exists in configuration", repo)
		}
	}

	// Create command string or object based on version
	var newCmd string
	if version != "" {
		newCmd = repo + "@" + version
	} else {
		newCmd = repo
	}

	// Add to commands as string array
	if config.Commands == nil {
		config.Commands = []string{newCmd}
	} else {
		switch v := config.Commands.(type) {
		case []string:
			v = append(v, newCmd)
			config.Commands = v
		case []interface{}:
			v = append(v, newCmd)
			config.Commands = v
		default:
			// Convert to string array
			config.Commands = []string{newCmd}
		}
	}

	return m.SaveConfig(config)
}

// RemoveCommand removes a command from the configuration
func (m *Manager) RemoveCommand(repo string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current commands
	commands, err := config.GetCommands()
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	found := false
	newCommands := make([]string, 0, len(commands))
	for _, cmd := range commands {
		if cmd.Repo != repo {
			// Preserve format
			if cmd.Version != "" {
				newCommands = append(newCommands, cmd.Repo+"@"+cmd.Version)
			} else {
				newCommands = append(newCommands, cmd.Repo)
			}
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
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current commands
	commands, err := config.GetCommands()
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	found := false
	newCommands := make([]string, 0, len(commands))
	for _, cmd := range commands {
		if cmd.Repo == repo {
			found = true
			if newVersion != "" {
				newCommands = append(newCommands, repo+"@"+newVersion)
			} else {
				newCommands = append(newCommands, repo)
			}
		} else {
			// Preserve format
			if cmd.Version != "" {
				newCommands = append(newCommands, cmd.Repo+"@"+cmd.Version)
			} else {
				newCommands = append(newCommands, cmd.Repo)
			}
		}
	}

	if !found {
		return fmt.Errorf("command %s not found in configuration", repo)
	}

	config.Commands = newCommands
	return m.SaveConfig(config)
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

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

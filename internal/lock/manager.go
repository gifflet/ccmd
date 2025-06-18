package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

const (
	// LockFileName is the name of the lock file
	LockFileName = "commands.lock"
	// LockFileVersion is the current version of the lock file format
	LockFileVersion = "1.0"
)

// Manager handles operations on the commands.lock file
type Manager struct {
	mu         sync.RWMutex
	filePath   string
	fileSystem fs.FileSystem
	lockFile   *models.LockFile
	loaded     bool
}

// NewManager creates a new lock file manager
func NewManager(dir string) *Manager {
	return &Manager{
		filePath:   filepath.Join(dir, LockFileName),
		fileSystem: fs.OS{},
	}
}

// NewManagerWithFS creates a new lock file manager with a custom file system
func NewManagerWithFS(dir string, fileSystem fs.FileSystem) *Manager {
	return &Manager{
		filePath:   filepath.Join(dir, LockFileName),
		fileSystem: fileSystem,
	}
}

// Load reads the lock file from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := m.fileSystem.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize empty lock file if it doesn't exist
			m.lockFile = &models.LockFile{
				Version:  LockFileVersion,
				Commands: make(map[string]*models.Command),
			}
			m.loaded = true
			return nil
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile models.LockFile
	if err := json.Unmarshal(data, &lockFile); err != nil {
		return fmt.Errorf("failed to parse lock file: %w", err)
	}

	m.lockFile = &lockFile
	m.loaded = true
	return nil
}

// Save writes the lock file to disk atomically
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return ErrNotLoaded
	}

	data, err := json.MarshalIndent(m.lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	// Create backup before writing
	if err := m.createBackup(); err != nil {
		// Log error but continue
		_ = err
	}

	// Write to temporary file first
	tempPath := m.filePath + ".tmp"
	if err := m.fileSystem.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary lock file: %w", err)
	}

	// Atomically rename temporary file to actual lock file
	if err := m.fileSystem.Rename(tempPath, m.filePath); err != nil {
		// Clean up temporary file
		if removeErr := m.fileSystem.Remove(tempPath); removeErr != nil {
			// Log error but don't fail the operation
			_ = removeErr
		}
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	return nil
}

// AddCommand adds a new command to the lock file
func (m *Manager) AddCommand(cmd *models.Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return ErrNotLoaded
	}

	if cmd.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	cmd.InstalledAt = time.Now()
	cmd.UpdatedAt = cmd.InstalledAt
	m.lockFile.Commands[cmd.Name] = cmd

	return nil
}

// UpdateCommand updates an existing command in the lock file
func (m *Manager) UpdateCommand(name string, updateFn func(*models.Command) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return ErrNotLoaded
	}

	cmd, exists := m.lockFile.Commands[name]
	if !exists {
		return fmt.Errorf("command %q not found", name)
	}

	if err := updateFn(cmd); err != nil {
		return fmt.Errorf("failed to update command: %w", err)
	}

	cmd.UpdatedAt = time.Now()
	return nil
}

// RemoveCommand removes a command from the lock file
func (m *Manager) RemoveCommand(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return ErrNotLoaded
	}

	if _, exists := m.lockFile.Commands[name]; !exists {
		return fmt.Errorf("command %q not found", name)
	}

	delete(m.lockFile.Commands, name)
	return nil
}

// GetCommand retrieves a command from the lock file
func (m *Manager) GetCommand(name string) (*models.Command, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, ErrNotLoaded
	}

	cmd, exists := m.lockFile.Commands[name]
	if !exists {
		return nil, fmt.Errorf("command %q not found", name)
	}

	// Return a copy to prevent external modifications
	cmdCopy := *cmd
	return &cmdCopy, nil
}

// ListCommands returns a list of all commands in the lock file
func (m *Manager) ListCommands() ([]*models.Command, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, ErrNotLoaded
	}

	commands := make([]*models.Command, 0, len(m.lockFile.Commands))
	for _, cmd := range m.lockFile.Commands {
		cmdCopy := *cmd
		commands = append(commands, &cmdCopy)
	}

	return commands, nil
}

// HasCommand checks if a command exists in the lock file
func (m *Manager) HasCommand(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return false
	}

	_, exists := m.lockFile.Commands[name]
	return exists
}

// createBackup creates a backup of the current lock file
func (m *Manager) createBackup() error {
	// Check if lock file exists
	if _, err := m.fileSystem.Stat(m.filePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	backupPath := m.filePath + ".bak"
	data, err := m.fileSystem.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to read lock file for backup: %w", err)
	}

	if err := m.fileSystem.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

package project

import (
	"path/filepath"
	"sync"

	"github.com/gifflet/ccmd/internal/fs"
)

// LockManager provides compatibility wrapper for old internal/lock.Manager
type LockManager struct {
	mu       sync.RWMutex
	manager  *Manager
	lockFile *LockFile
	loaded   bool
}

// NewLockManager creates a new lock manager for the given directory
func NewLockManager(dir string) *LockManager {
	return &LockManager{
		manager: NewManager(filepath.Dir(dir)),
	}
}

// NewLockManagerWithFS creates a new lock manager with a file system (for testing)
func NewLockManagerWithFS(dir string, fileSystem fs.FileSystem) *LockManager {
	return &LockManager{
		manager: NewManagerWithFS(filepath.Dir(dir), fileSystem),
	}
}

// Load reads the lock file from disk
func (m *LockManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.manager.MigrateLegacyLockFile(); err != nil {
		_ = err
	}

	if !m.manager.LockExists() {
		m.lockFile = NewLockFile()
		m.loaded = true
		return nil
	}

	lockFile, err := m.manager.LoadLockFile()
	if err != nil {
		return err
	}

	m.lockFile = lockFile
	m.loaded = true
	return nil
}

// Save writes the lock file to disk
func (m *LockManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded || m.lockFile == nil {
		return nil
	}

	return m.manager.SaveLockFile(m.lockFile)
}

// AddCommand adds a command to the lock file
func (m *LockManager) AddCommand(cmd interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return nil
	}

	// Handle different command types
	switch c := cmd.(type) {
	case *CommandLockInfo:
		return m.lockFile.AddCommand(c)
	case *ModelCommand:
		newCmd := c.ToCommandLockInfo()
		return m.lockFile.AddCommand(newCmd)
	default:
		// Try to convert from old model
		if oldCmd, ok := cmd.(interface {
			GetName() string
			GetVersion() string
			GetSource() string
		}); ok {
			newCmd := &CommandLockInfo{
				Name:     oldCmd.GetName(),
				Version:  oldCmd.GetVersion(),
				Source:   oldCmd.GetSource(),
				Resolved: oldCmd.GetSource() + "@" + oldCmd.GetVersion(),
			}
			return m.lockFile.AddCommand(newCmd)
		}
	}

	return nil
}

// RemoveCommand removes a command from the lock file
func (m *LockManager) RemoveCommand(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return nil
	}

	m.lockFile.RemoveCommand(name)
	return nil
}

// GetCommand retrieves a command by name
func (m *LockManager) GetCommand(name string) (*CommandLockInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, nil
	}

	cmd, exists := m.lockFile.GetCommand(name)
	if !exists {
		return nil, nil
	}

	return cmd, nil
}

// ListCommands returns all commands in the lock file
func (m *LockManager) ListCommands() ([]*CommandLockInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded || m.lockFile == nil {
		return []*CommandLockInfo{}, nil
	}

	return m.lockFile.ListCommands()
}

// UpdateCommand updates an existing command
func (m *LockManager) UpdateCommand(cmd interface{}) error {
	return m.AddCommand(cmd)
}

// HasCommand checks if a command exists in the lock file
func (m *LockManager) HasCommand(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded || m.lockFile == nil {
		return false
	}

	_, exists := m.lockFile.GetCommand(name)
	return exists
}

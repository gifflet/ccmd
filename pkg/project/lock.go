// Package project provides functionality for managing ccmd project files.
package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LockFileVersion represents the current version of the lock file format
const LockFileVersion = "1.0"

// LockFile represents the ccmd-lock.yaml file structure
type LockFile struct {
	Version   string              `yaml:"version"`
	UpdatedAt time.Time           `yaml:"updated_at"`
	Commands  map[string]*Command `yaml:"commands"`
}

// Command represents a locked command entry
type Command struct {
	Name         string            `yaml:"name"`
	Repository   string            `yaml:"repository"`
	Version      string            `yaml:"version"`     // Tag, branch, or commit
	CommitHash   string            `yaml:"commit_hash"` // Exact commit SHA
	InstalledAt  time.Time         `yaml:"installed_at"`
	UpdatedAt    time.Time         `yaml:"updated_at"`
	FileSize     int64             `yaml:"file_size"`
	Checksum     string            `yaml:"checksum"` // SHA256 of the binary
	Dependencies []string          `yaml:"dependencies,omitempty"`
	Metadata     map[string]string `yaml:"metadata,omitempty"`
}

// NewLockFile creates a new lock file with the current version
func NewLockFile() *LockFile {
	return &LockFile{
		Version:   LockFileVersion,
		UpdatedAt: time.Now(),
		Commands:  make(map[string]*Command),
	}
}

// AddCommand adds or updates a command in the lock file
func (lf *LockFile) AddCommand(cmd *Command) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	if lf.Commands == nil {
		lf.Commands = make(map[string]*Command)
	}

	lf.Commands[cmd.Name] = cmd
	lf.UpdatedAt = time.Now()
	return nil
}

// RemoveCommand removes a command from the lock file
func (lf *LockFile) RemoveCommand(name string) bool {
	if _, exists := lf.Commands[name]; exists {
		delete(lf.Commands, name)
		lf.UpdatedAt = time.Now()
		return true
	}
	return false
}

// GetCommand retrieves a command by name
func (lf *LockFile) GetCommand(name string) (*Command, bool) {
	cmd, exists := lf.Commands[name]
	return cmd, exists
}

// Validate validates the lock file structure
func (lf *LockFile) Validate() error {
	if lf.Version == "" {
		return fmt.Errorf("version is required")
	}

	if lf.Commands == nil {
		lf.Commands = make(map[string]*Command)
	}

	for name, cmd := range lf.Commands {
		if cmd.Name != name {
			return fmt.Errorf("command name mismatch: key=%s, name=%s", name, cmd.Name)
		}
		if err := cmd.Validate(); err != nil {
			return fmt.Errorf("invalid command %s: %w", name, err)
		}
	}

	return nil
}

// Validate validates a command entry
func (c *Command) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.CommitHash == "" {
		return fmt.Errorf("commit_hash is required")
	}
	if len(c.CommitHash) != 40 {
		return fmt.Errorf("commit_hash must be a 40-character SHA")
	}
	if c.InstalledAt.IsZero() {
		return fmt.Errorf("installed_at is required")
	}
	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	if c.FileSize <= 0 {
		return fmt.Errorf("file_size must be positive")
	}
	if c.Checksum == "" {
		return fmt.Errorf("checksum is required")
	}
	if len(c.Checksum) != 64 {
		return fmt.Errorf("checksum must be a 64-character SHA256 hash")
	}

	return nil
}

// CalculateChecksum calculates the SHA256 checksum of a file
func CalculateChecksum(filepath string) (string, error) {
	// #nosec G304 -- filepath is provided by the application, not user input
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't fail the operation
			_ = closeErr
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// LoadFromFile loads a lock file from disk
func LoadFromFile(filepath string) (*LockFile, error) {
	// #nosec G304 -- filepath is provided by the application, not user input
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile LockFile
	if err := yaml.Unmarshal(data, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	if err := lockFile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lock file: %w", err)
	}

	return &lockFile, nil
}

// SaveToFile saves the lock file to disk
func (lf *LockFile) SaveToFile(filepath string) error {
	if err := lf.Validate(); err != nil {
		return fmt.Errorf("invalid lock file: %w", err)
	}

	// Lock files should be owner-readable only (0600)
	return writeYAMLFile(filepath, lf, 0o600)
}

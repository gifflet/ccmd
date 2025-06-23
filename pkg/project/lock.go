// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
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

	"github.com/gifflet/ccmd/internal/fs"
)

// LockFileVersion represents the current version of the lock file format
const LockFileVersion = "1.0"

// LockFile represents the ccmd-lock.yaml file structure
type LockFile struct {
	Version         string                      `yaml:"version"`
	LockfileVersion int                         `yaml:"lockfileVersion"`
	Commands        map[string]*CommandLockInfo `yaml:"commands"`
}

// CommandLockInfo represents a locked command entry according to PRD
type CommandLockInfo struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Source       string            `yaml:"source"`
	Resolved     string            `yaml:"resolved"`
	Commit       string            `yaml:"commit,omitempty"`
	InstalledAt  time.Time         `yaml:"installed_at"`
	UpdatedAt    time.Time         `yaml:"updated_at"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	Metadata     map[string]string `yaml:"metadata,omitempty"`
}

// Command is kept for compatibility during migration
type Command = CommandLockInfo

// NewLockFile creates a new lock file with the current version
func NewLockFile() *LockFile {
	return &LockFile{
		Version:         LockFileVersion,
		LockfileVersion: 1,
		Commands:        make(map[string]*CommandLockInfo),
	}
}

// AddCommand adds or updates a command in the lock file
func (lf *LockFile) AddCommand(cmd *Command) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	if lf.Commands == nil {
		lf.Commands = make(map[string]*CommandLockInfo)
	}

	lf.Commands[cmd.Name] = cmd
	return nil
}

// RemoveCommand removes a command from the lock file
func (lf *LockFile) RemoveCommand(name string) bool {
	if _, exists := lf.Commands[name]; exists {
		delete(lf.Commands, name)
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
		lf.Commands = make(map[string]*CommandLockInfo)
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
func (c *CommandLockInfo) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Source == "" {
		return fmt.Errorf("source is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Commit == "" {
		return fmt.Errorf("commit is required")
	}
	if len(c.Commit) != 40 {
		return fmt.Errorf("commit must be a 40-character SHA")
	}
	if c.InstalledAt.IsZero() {
		return fmt.Errorf("installed_at is required")
	}
	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
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
func LoadFromFile(filepath string, fileSystem fs.FileSystem) (*LockFile, error) {
	data, err := fileSystem.ReadFile(filepath)
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
func (lf *LockFile) SaveToFile(filepath string, fileSystem fs.FileSystem) error {
	if err := lf.Validate(); err != nil {
		return fmt.Errorf("invalid lock file: %w", err)
	}

	// Lock files should be owner-readable only (0600)
	return writeYAMLFile(filepath, lf, 0o600, fileSystem)
}

// ListCommands returns a list of all commands in the lock file
func (lf *LockFile) ListCommands() ([]*CommandLockInfo, error) {
	if lf.Commands == nil {
		return []*CommandLockInfo{}, nil
	}

	commands := make([]*CommandLockInfo, 0, len(lf.Commands))
	for _, cmd := range lf.Commands {
		commands = append(commands, cmd)
	}
	return commands, nil
}

// SetCommand sets or updates a command in the lock file (compatibility method)
func (lf *LockFile) SetCommand(name string, cmd *CommandLockInfo) {
	if lf.Commands == nil {
		lf.Commands = make(map[string]*CommandLockInfo)
	}
	lf.Commands[name] = cmd
}

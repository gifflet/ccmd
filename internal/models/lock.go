package models

import (
	"fmt"
	"time"
)

// LockFile represents the commands.lock file structure
type LockFile struct {
	Version  string              `json:"version"`
	Commands map[string]*Command `json:"commands"`
}

// Command represents an installed command in the lock file
type Command struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Source       string            `json:"source"`
	InstalledAt  time.Time         `json:"installed_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Validate validates the lock file structure
func (lf *LockFile) Validate() error {
	if lf.Commands == nil {
		return fmt.Errorf("commands map cannot be nil")
	}

	for name, cmd := range lf.Commands {
		if err := cmd.Validate(); err != nil {
			return fmt.Errorf("invalid command %s: %w", name, err)
		}
	}

	return nil
}

// Validate validates a single command entry
func (c *Command) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Source == "" {
		return fmt.Errorf("source is required")
	}
	if c.InstalledAt.IsZero() {
		return fmt.Errorf("installedAt is required")
	}
	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updatedAt is required")
	}
	return nil
}

// GetCommand returns a command entry by name
func (lf *LockFile) GetCommand(name string) (*Command, bool) {
	cmd, exists := lf.Commands[name]
	return cmd, exists
}

// SetCommand sets or updates a command entry
func (lf *LockFile) SetCommand(name string, cmd *Command) {
	if lf.Commands == nil {
		lf.Commands = make(map[string]*Command)
	}
	lf.Commands[name] = cmd
}

// RemoveCommand removes a command from the lock file
func (lf *LockFile) RemoveCommand(name string) {
	delete(lf.Commands, name)
}

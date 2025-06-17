package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// LockFile represents the commands.lock file structure
type LockFile struct {
	Commands map[string]CommandLock `json:"commands"`
}

// CommandLock represents a single command entry in the lock file
type CommandLock struct {
	Version     string    `json:"version"`
	Repository  string    `json:"repository"`
	InstalledAt time.Time `json:"installedAt"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Validate validates the lock file structure
func (lf *LockFile) Validate() error {
	if lf.Commands == nil {
		return fmt.Errorf("commands map cannot be nil")
	}

	for name, lock := range lf.Commands {
		if err := lock.Validate(); err != nil {
			return fmt.Errorf("invalid lock for command %s: %w", name, err)
		}
	}

	return nil
}

// Validate validates a single command lock entry
func (cl *CommandLock) Validate() error {
	if cl.Version == "" {
		return fmt.Errorf("version is required")
	}
	if cl.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	if cl.InstalledAt.IsZero() {
		return fmt.Errorf("installedAt is required")
	}
	if cl.LastUpdated.IsZero() {
		return fmt.Errorf("lastUpdated is required")
	}
	return nil
}

// MarshalJSON marshals LockFile to JSON with proper formatting
func (lf *LockFile) MarshalJSON() ([]byte, error) {
	type Alias LockFile
	return json.MarshalIndent((*Alias)(lf), "", "  ")
}

// UnmarshalJSON unmarshals JSON data into LockFile
func (lf *LockFile) UnmarshalJSON(data []byte) error {
	type Alias LockFile
	return json.Unmarshal(data, (*Alias)(lf))
}

// GetCommand returns a command lock entry by name
func (lf *LockFile) GetCommand(name string) (CommandLock, bool) {
	lock, exists := lf.Commands[name]
	return lock, exists
}

// SetCommand sets or updates a command lock entry
func (lf *LockFile) SetCommand(name string, lock CommandLock) {
	if lf.Commands == nil {
		lf.Commands = make(map[string]CommandLock)
	}
	lf.Commands[name] = lock
}

// RemoveCommand removes a command from the lock file
func (lf *LockFile) RemoveCommand(name string) {
	delete(lf.Commands, name)
}

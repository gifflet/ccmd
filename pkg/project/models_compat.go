// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package project

import (
	"time"
)

// ModelCommand provides compatibility with internal/models.Command
type ModelCommand struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Source       string            `json:"source"`
	Repository   string            `json:"repository"`
	CommitHash   string            `json:"commit_hash,omitempty"`
	InstalledAt  time.Time         `json:"installed_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	FileSize     int64             `json:"file_size,omitempty"`
	Checksum     string            `json:"checksum,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ToCommandLockInfo converts ModelCommand to CommandLockInfo
func (c *ModelCommand) ToCommandLockInfo() *CommandLockInfo {
	cmd := &CommandLockInfo{
		Name:        c.Name,
		Version:     c.Version,
		Source:      c.Source,
		InstalledAt: c.InstalledAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Generate resolved field
	if c.Version != "" && c.Source != "" {
		cmd.Resolved = c.Source + "@" + c.Version
	}

	return cmd
}

// FromCommandLockInfo creates ModelCommand from CommandLockInfo
func FromCommandLockInfo(cmd *CommandLockInfo) *ModelCommand {
	return &ModelCommand{
		Name:        cmd.Name,
		Version:     cmd.Version,
		Source:      cmd.Source,
		InstalledAt: cmd.InstalledAt,
		UpdatedAt:   cmd.UpdatedAt,
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"encoding/json"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
)

// LockFile represents the ccmd-lock.yaml structure
type LockFile struct {
	Version         string                  `yaml:"version"`
	LockfileVersion int                     `yaml:"lockfileVersion"`
	Commands        map[string]*LockCommand `yaml:"commands"`
}

// LockCommand represents a command entry in the lock file
type LockCommand struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	Source      string    `yaml:"source"`
	Resolved    string    `yaml:"resolved"`
	Commit      string    `yaml:"commit"`
	InstalledAt time.Time `yaml:"installed_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
}

// InstalledCommand represents an installed command
type InstalledCommand struct {
	Name        string
	Version     string
	Description string
	Author      string
	Repository  string
	Path        string
}

// ProjectConfig represents the ccmd.yaml configuration file
type ProjectConfig struct {
	// Project metadata (when ccmd.yaml is for a command)
	Name        string   `yaml:"name,omitempty" json:"name,omitempty"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Description string   `yaml:"description" json:"description"`
	Author      string   `yaml:"author" json:"author"`
	Repository  string   `yaml:"repository" json:"repository"`
	Entry       string   `yaml:"entry,omitempty" json:"entry,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`

	// Commands list (when ccmd.yaml is for a project)
	Commands []string `yaml:"commands,omitempty" json:"commands,omitempty"`
}

// ConfigCommand represents a command in the configuration
type ConfigCommand struct {
	Repo    string `yaml:"repo"`
	Version string `yaml:"version,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

// Validate validates the project config when used as command metadata
func (pc *ProjectConfig) Validate() error {
	// Only validate when used as command metadata (has metadata fields)
	if pc.Name == "" && pc.Version == "" && len(pc.Commands) == 0 {
		return errors.InvalidInput("either command metadata or commands list is required")
	}

	// If it has command metadata, validate required fields
	if pc.Name != "" || pc.Version != "" {
		if pc.Name == "" {
			return errors.InvalidInput("name is required")
		}
		if pc.Version == "" {
			return errors.InvalidInput("version is required")
		}
		if pc.Description == "" {
			return errors.InvalidInput("description is required")
		}
		if pc.Author == "" {
			return errors.InvalidInput("author is required")
		}
		if pc.Repository == "" {
			return errors.InvalidInput("repository is required")
		}
		if pc.Entry == "" {
			return errors.InvalidInput("entry is required")
		}
	}

	return nil
}

// MarshalYAML marshals ProjectConfig to YAML
func (pc *ProjectConfig) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(pc)
}

// UnmarshalYAML unmarshals YAML data into ProjectConfig
func (pc *ProjectConfig) UnmarshalYAML(data []byte) error {
	return yaml.Unmarshal(data, pc)
}

// MarshalJSON marshals ProjectConfig to JSON
func (pc *ProjectConfig) MarshalJSON() ([]byte, error) {
	type Alias ProjectConfig
	return json.Marshal((*Alias)(pc))
}

// UnmarshalJSON unmarshals JSON data into ProjectConfig
func (pc *ProjectConfig) UnmarshalJSON(data []byte) error {
	type Alias ProjectConfig
	return json.Unmarshal(data, (*Alias)(pc))
}

// GetConfigCommands returns the commands from the configuration
func (pc *ProjectConfig) GetConfigCommands() []ConfigCommand {
	commands := make([]ConfigCommand, 0, len(pc.Commands))
	for _, cmd := range pc.Commands {
		repo, version := ParseCommandSpec(cmd)
		commands = append(commands, ConfigCommand{
			Repo:    repo,
			Version: version,
		})
	}
	return commands
}

// ParseCommandSpec parses a command specification (e.g., "owner/repo@version")
func ParseCommandSpec(spec string) (repo, version string) {
	parts := strings.Split(spec, "@")
	repo = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return repo, version
}

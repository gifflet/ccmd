/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package models

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
)

// CommandMetadata represents the structure of a ccmd.yaml file
type CommandMetadata struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Author      string   `yaml:"author" json:"author"`
	Repository  string   `yaml:"repository" json:"repository"`
	Entry       string   `yaml:"entry,omitempty" json:"entry,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
}

// Validate validates the command metadata
func (cm *CommandMetadata) Validate() error {
	if cm.Name == "" {
		return errors.InvalidInput("name is required")
	}
	if cm.Version == "" {
		return errors.InvalidInput("version is required")
	}
	if cm.Description == "" {
		return errors.InvalidInput("description is required")
	}
	if cm.Author == "" {
		return errors.InvalidInput("author is required")
	}
	if cm.Repository == "" {
		return errors.InvalidInput("repository is required")
	}
	if cm.Entry == "" {
		return errors.InvalidInput("entry is required")
	}
	return nil
}

// MarshalYAML marshals CommandMetadata to YAML
func (cm *CommandMetadata) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(cm)
}

// UnmarshalYAML unmarshals YAML data into CommandMetadata
func (cm *CommandMetadata) UnmarshalYAML(data []byte) error {
	return yaml.Unmarshal(data, cm)
}

// MarshalJSON marshals CommandMetadata to JSON
func (cm *CommandMetadata) MarshalJSON() ([]byte, error) {
	type Alias CommandMetadata
	return json.Marshal((*Alias)(cm))
}

// UnmarshalJSON unmarshals JSON data into CommandMetadata
func (cm *CommandMetadata) UnmarshalJSON(data []byte) error {
	type Alias CommandMetadata
	return json.Unmarshal(data, (*Alias)(cm))
}

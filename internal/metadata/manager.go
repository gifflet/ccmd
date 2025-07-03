/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package metadata

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/errors"
)

const (
	// MetadataFileName is the standard name for command metadata files
	MetadataFileName = "ccmd.yaml"
)

// Manager provides methods for reading and writing command metadata
type Manager struct{}

// NewManager creates a new metadata manager
func NewManager() *Manager {
	return &Manager{}
}

// ReadCommandMetadata reads and parses a command metadata file from the specified directory
func (m *Manager) ReadCommandMetadata(commandDir string) (*models.CommandMetadata, error) {
	metadataPath := filepath.Join(commandDir, MetadataFileName)

	data, err := os.ReadFile(metadataPath) // #nosec G304 - path is safely constructed with filepath.Join
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NotFound(fmt.Sprintf("metadata file at %s", metadataPath))
		}
		return nil, errors.FileError("read metadata file", metadataPath, err)
	}

	var metadata models.CommandMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, errors.FileError("parse metadata file", metadataPath, err)
	}

	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// WriteCommandMetadata writes command metadata to a file in the specified directory
func (m *Manager) WriteCommandMetadata(commandDir string, metadata *models.CommandMetadata) error {
	if err := metadata.Validate(); err != nil {
		return err
	}

	metadataPath := filepath.Join(commandDir, MetadataFileName)

	data, err := yaml.Marshal(metadata)
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to marshal metadata: %v", err))
	}

	// Ensure the directory exists
	if err := os.MkdirAll(commandDir, 0o750); err != nil {
		return errors.FileError("create command directory", commandDir, err)
	}

	if err := os.WriteFile(metadataPath, data, 0o600); err != nil {
		return errors.FileError("write metadata file", metadataPath, err)
	}

	return nil
}

// UpdateCommandMetadata reads existing metadata, applies updates, and writes it back
func (m *Manager) UpdateCommandMetadata(commandDir string, updates func(*models.CommandMetadata) error) error {
	// Read existing metadata
	metadata, err := m.ReadCommandMetadata(commandDir)
	if err != nil {
		return err
	}

	// Apply updates
	if err := updates(metadata); err != nil {
		return err
	}

	// Write updated metadata
	if err := m.WriteCommandMetadata(commandDir, metadata); err != nil {
		return err
	}

	return nil
}

// ExtractCommandInfo extracts basic command information from metadata
func (m *Manager) ExtractCommandInfo(metadata *models.CommandMetadata) map[string]interface{} {
	info := make(map[string]interface{})

	info["name"] = metadata.Name
	info["version"] = metadata.Version
	info["description"] = metadata.Description
	info["author"] = metadata.Author
	info["repository"] = metadata.Repository

	if metadata.Entry != "" {
		info["entry"] = metadata.Entry
	}

	if len(metadata.Tags) > 0 {
		info["tags"] = metadata.Tags
	}

	if metadata.License != "" {
		info["license"] = metadata.License
	}

	if metadata.Homepage != "" {
		info["homepage"] = metadata.Homepage
	}

	return info
}

// Exists checks if a metadata file exists in the specified directory
func (m *Manager) Exists(commandDir string) bool {
	metadataPath := filepath.Join(commandDir, MetadataFileName)
	_, err := os.Stat(metadataPath)
	return err == nil
}

package installer

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

// metadataManager handles metadata operations using the FileSystem interface
type metadataManager struct {
	fs fs.FileSystem
}

// newMetadataManager creates a new metadata manager with the given filesystem
func newMetadataManager(fs fs.FileSystem) *metadataManager {
	return &metadataManager{fs: fs}
}

// ReadCommandMetadata reads and parses a command metadata file from the specified directory
func (m *metadataManager) ReadCommandMetadata(commandDir string) (*models.CommandMetadata, error) {
	metadataPath := filepath.Join(commandDir, "ccmd.yaml")

	data, err := m.fs.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("metadata file not found at %s", metadataPath)
	}

	var metadata models.CommandMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	return &metadata, nil
}

// WriteCommandMetadata writes command metadata to a file in the specified directory
func (m *metadataManager) WriteCommandMetadata(commandDir string, metadata *models.CommandMetadata) error {
	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	metadataPath := filepath.Join(commandDir, "ccmd.yaml")

	data, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Ensure the directory exists
	if err := m.fs.MkdirAll(commandDir, 0o750); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	if err := m.fs.WriteFile(metadataPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// UpdateCommandMetadata reads existing metadata, applies updates, and writes it back
func (m *metadataManager) UpdateCommandMetadata(commandDir string, updates func(*models.CommandMetadata) error) error {
	// Read existing metadata
	metadata, err := m.ReadCommandMetadata(commandDir)
	if err != nil {
		return fmt.Errorf("failed to read existing metadata: %w", err)
	}

	// Apply updates
	if err := updates(metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// Write updated metadata
	if err := m.WriteCommandMetadata(commandDir, metadata); err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}

	return nil
}

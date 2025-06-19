package metadata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/models"
)

func TestManager_ReadCommandMetadata(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, dir string)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, metadata *models.CommandMetadata)
	}{
		{
			name: "successfully read valid metadata",
			setupFunc: func(t *testing.T, dir string) {
				content := `name: test-command
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/command
entry: main.go
tags:
  - cli
  - testing
license: MIT
homepage: https://example.com`
				err := os.WriteFile(filepath.Join(dir, MetadataFileName), []byte(content), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
			validate: func(t *testing.T, metadata *models.CommandMetadata) {
				assert.Equal(t, "test-command", metadata.Name)
				assert.Equal(t, "1.0.0", metadata.Version)
				assert.Equal(t, "A test command", metadata.Description)
				assert.Equal(t, "Test Author", metadata.Author)
				assert.Equal(t, "https://github.com/test/command", metadata.Repository)
				assert.Equal(t, "main.go", metadata.Entry)
				assert.Equal(t, []string{"cli", "testing"}, metadata.Tags)
				assert.Equal(t, "MIT", metadata.License)
				assert.Equal(t, "https://example.com", metadata.Homepage)
			},
		},
		{
			name:        "metadata file not found",
			setupFunc:   func(t *testing.T, dir string) {},
			wantErr:     true,
			errContains: "metadata file not found",
		},
		{
			name: "invalid yaml format",
			setupFunc: func(t *testing.T, dir string) {
				content := `invalid yaml content
				[this is not valid yaml`
				err := os.WriteFile(filepath.Join(dir, MetadataFileName), []byte(content), 0644)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "failed to parse metadata file",
		},
		{
			name: "missing required fields",
			setupFunc: func(t *testing.T, dir string) {
				content := `name: test-command
version: 1.0.0
description: A test command`
				err := os.WriteFile(filepath.Join(dir, MetadataFileName), []byte(content), 0644)
				require.NoError(t, err)
			},
			wantErr:     true,
			errContains: "invalid metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupFunc(t, tmpDir)

			manager := NewManager()
			metadata, err := manager.ReadCommandMetadata(tmpDir)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metadata)
				if tt.validate != nil {
					tt.validate(t, metadata)
				}
			}
		})
	}
}

func TestManager_WriteCommandMetadata(t *testing.T) {
	tests := []struct {
		name        string
		metadata    *models.CommandMetadata
		wantErr     bool
		errContains string
	}{
		{
			name: "successfully write valid metadata",
			metadata: &models.CommandMetadata{
				Name:        "test-command",
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "https://github.com/test/command",
				Entry:       "main.go",
				Tags:        []string{"cli", "testing"},
				License:     "MIT",
				Homepage:    "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid metadata - missing name",
			metadata: &models.CommandMetadata{
				Version:     "1.0.0",
				Description: "A test command",
				Author:      "Test Author",
				Repository:  "https://github.com/test/command",
			},
			wantErr:     true,
			errContains: "invalid metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewManager()

			err := manager.WriteCommandMetadata(tmpDir, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify the file was written
				metadataPath := filepath.Join(tmpDir, MetadataFileName)
				assert.FileExists(t, metadataPath)

				// Read it back to verify content
				readMetadata, err := manager.ReadCommandMetadata(tmpDir)
				require.NoError(t, err)
				assert.Equal(t, tt.metadata, readMetadata)
			}
		})
	}
}

func TestManager_UpdateCommandMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	// Write initial metadata
	initialMetadata := &models.CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "https://github.com/test/command",
	}
	err := manager.WriteCommandMetadata(tmpDir, initialMetadata)
	require.NoError(t, err)

	// Update metadata
	err = manager.UpdateCommandMetadata(tmpDir, func(metadata *models.CommandMetadata) error {
		metadata.Version = "2.0.0"
		metadata.Tags = []string{"updated", "cli"}
		return nil
	})
	assert.NoError(t, err)

	// Verify update
	updatedMetadata, err := manager.ReadCommandMetadata(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", updatedMetadata.Version)
	assert.Equal(t, []string{"updated", "cli"}, updatedMetadata.Tags)
}

func TestManager_ExtractCommandInfo(t *testing.T) {
	manager := NewManager()

	metadata := &models.CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "https://github.com/test/command",
		Entry:       "main.go",
		Tags:        []string{"cli", "testing"},
		License:     "MIT",
		Homepage:    "https://example.com",
	}

	info := manager.ExtractCommandInfo(metadata)

	assert.Equal(t, "test-command", info["name"])
	assert.Equal(t, "1.0.0", info["version"])
	assert.Equal(t, "A test command", info["description"])
	assert.Equal(t, "Test Author", info["author"])
	assert.Equal(t, "https://github.com/test/command", info["repository"])
	assert.Equal(t, "main.go", info["entry"])
	assert.Equal(t, []string{"cli", "testing"}, info["tags"])
	assert.Equal(t, "MIT", info["license"])
	assert.Equal(t, "https://example.com", info["homepage"])
}

func TestManager_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	// Test when metadata doesn't exist
	assert.False(t, manager.Exists(tmpDir))

	// Create metadata file
	metadata := &models.CommandMetadata{
		Name:        "test-command",
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "https://github.com/test/command",
	}
	err := manager.WriteCommandMetadata(tmpDir, metadata)
	require.NoError(t, err)

	// Test when metadata exists
	assert.True(t, manager.Exists(tmpDir))
}

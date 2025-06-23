// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package metadata

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/models"
)

func TestHelperFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	// Test Exists when file doesn't exist
	assert.False(t, Exists(tmpDir))

	// Test WriteCommandMetadata
	metadata := &models.CommandMetadata{
		Name:        "helper-test",
		Version:     "1.0.0",
		Description: "Testing helper functions",
		Author:      "Test Author",
		Repository:  "https://github.com/test/helper",
		Tags:        []string{"test", "helper"},
		Entry:       "helper-test",
	}

	err := WriteCommandMetadata(tmpDir, metadata)
	require.NoError(t, err)

	// Test Exists when file exists
	assert.True(t, Exists(tmpDir))
	assert.FileExists(t, filepath.Join(tmpDir, MetadataFileName))

	// Test ReadCommandMetadata
	readMetadata, err := ReadCommandMetadata(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, metadata, readMetadata)

	// Test ExtractCommandInfo
	info := ExtractCommandInfo(readMetadata)
	assert.Equal(t, "helper-test", info["name"])
	assert.Equal(t, "1.0.0", info["version"])
	assert.Equal(t, []string{"test", "helper"}, info["tags"])

	// Test UpdateCommandMetadata
	err = UpdateCommandMetadata(tmpDir, func(m *models.CommandMetadata) error {
		m.Version = "2.0.0"
		m.License = "MIT"
		return nil
	})
	require.NoError(t, err)

	// Verify update
	updatedMetadata, err := ReadCommandMetadata(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", updatedMetadata.Version)
	assert.Equal(t, "MIT", updatedMetadata.License)
}

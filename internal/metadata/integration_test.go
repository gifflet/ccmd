// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/metadata"
	"github.com/gifflet/ccmd/internal/models"
)

func TestIntegration_MetadataLifecycle(t *testing.T) {
	// Create a temporary directory to simulate command installation
	tmpDir := t.TempDir()
	commandDir := filepath.Join(tmpDir, "commands", "my-awesome-command")

	// Step 1: Check if metadata exists (should not exist)
	assert.False(t, metadata.Exists(commandDir))

	// Step 2: Create and write new command metadata
	cmdMetadata := &models.CommandMetadata{
		Name:        "my-awesome-command",
		Version:     "1.0.0",
		Description: "An awesome command that does amazing things",
		Author:      "John Doe <john@example.com>",
		Repository:  "https://github.com/johndoe/my-awesome-command",
		Entry:       "cmd/main.go",
		Tags:        []string{"cli", "tool", "awesome"},
		License:     "MIT",
		Homepage:    "https://myawesomecommand.com",
	}

	err := metadata.WriteCommandMetadata(commandDir, cmdMetadata)
	require.NoError(t, err)

	// Step 3: Verify metadata file was created
	assert.True(t, metadata.Exists(commandDir))
	assert.FileExists(t, filepath.Join(commandDir, "ccmd.yaml"))

	// Step 4: Read the metadata back
	readMetadata, err := metadata.ReadCommandMetadata(commandDir)
	require.NoError(t, err)
	assert.Equal(t, cmdMetadata, readMetadata)

	// Step 5: Extract command info
	info := metadata.ExtractCommandInfo(readMetadata)
	assert.Equal(t, "my-awesome-command", info["name"])
	assert.Equal(t, "1.0.0", info["version"])
	assert.Equal(t, []string{"cli", "tool", "awesome"}, info["tags"])

	// Step 6: Update the metadata (simulate version update)
	err = metadata.UpdateCommandMetadata(commandDir, func(m *models.CommandMetadata) error {
		m.Version = "1.1.0"
		m.Tags = append(m.Tags, "updated")
		m.Description = "An awesome command that does even more amazing things"
		return nil
	})
	require.NoError(t, err)

	// Step 7: Verify the update
	updatedMetadata, err := metadata.ReadCommandMetadata(commandDir)
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", updatedMetadata.Version)
	assert.Contains(t, updatedMetadata.Tags, "updated")
	assert.Equal(t, "An awesome command that does even more amazing things", updatedMetadata.Description)

	// Step 8: Test reading invalid metadata file
	invalidDir := filepath.Join(tmpDir, "invalid-command")
	err = os.MkdirAll(invalidDir, 0o755)
	require.NoError(t, err)

	// Write invalid YAML
	invalidMetadataPath := filepath.Join(invalidDir, "ccmd.yaml")
	err = os.WriteFile(invalidMetadataPath, []byte("invalid: yaml: content: ["), 0o644)
	require.NoError(t, err)

	_, err = metadata.ReadCommandMetadata(invalidDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse metadata file")
}

func TestIntegration_MultipleCommands(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, "commands")

	// Create metadata for multiple commands
	commands := []models.CommandMetadata{
		{
			Name:        "command-one",
			Version:     "2.0.0",
			Description: "First command",
			Author:      "Author One",
			Repository:  "https://github.com/author/command-one",
			Entry:       "command-one",
		},
		{
			Name:        "command-two",
			Version:     "1.5.0",
			Description: "Second command",
			Author:      "Author Two",
			Repository:  "https://github.com/author/command-two",
			Tags:        []string{"utility", "helper"},
			Entry:       "command-two",
		},
		{
			Name:        "command-three",
			Version:     "3.1.0",
			Description: "Third command",
			Author:      "Author Three",
			Repository:  "https://github.com/author/command-three",
			License:     "Apache-2.0",
			Entry:       "command-three",
		},
	}

	// Write metadata for each command
	for _, cmd := range commands {
		cmdDir := filepath.Join(commandsDir, cmd.Name)
		err := metadata.WriteCommandMetadata(cmdDir, &cmd)
		require.NoError(t, err)
	}

	// Verify all commands have metadata
	entries, err := os.ReadDir(commandsDir)
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	// Read and verify each command's metadata
	for _, cmd := range commands {
		cmdDir := filepath.Join(commandsDir, cmd.Name)

		// Check existence
		assert.True(t, metadata.Exists(cmdDir))

		// Read metadata
		readCmd, err := metadata.ReadCommandMetadata(cmdDir)
		require.NoError(t, err)

		// Verify basic fields
		assert.Equal(t, cmd.Name, readCmd.Name)
		assert.Equal(t, cmd.Version, readCmd.Version)
		assert.Equal(t, cmd.Description, readCmd.Description)
	}
}

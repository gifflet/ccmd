/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "install [repository]", cmd.Use)
	assert.Equal(t, "Install a command from a Git repository or from ccmd.yaml", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	versionFlag := cmd.Flags().Lookup("version")
	assert.NotNil(t, versionFlag)
	assert.Equal(t, "v", versionFlag.Shorthand)

	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	nameFlag := cmd.Flags().Lookup("name")
	assert.NotNil(t, nameFlag)
	assert.Equal(t, "n", nameFlag.Shorthand)
}

func TestCommandArgs(t *testing.T) {
	cmd := NewCommand()

	// Test with valid args
	cmd.SetArgs([]string{"github.com/user/repo"})
	err := cmd.Args(cmd, []string{"github.com/user/repo"})
	assert.NoError(t, err)

	// Test with too many args
	cmd.SetArgs([]string{"arg1", "arg2"})
	err = cmd.Args(cmd, []string{"arg1", "arg2"})
	assert.Error(t, err)
}

// Note: Full integration tests for install command would require
// mocking Git operations and file system access. The core functionality
// is tested through the integration tests when running the actual commands.

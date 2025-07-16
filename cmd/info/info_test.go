/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "info <command-name>", cmd.Use)
	assert.Equal(t, "Display detailed information about an installed command", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	jsonFlag := cmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "false", jsonFlag.DefValue)
}

func TestCommandIntegration(t *testing.T) {
	// Integration test using the actual command
	cmd := NewCommand()

	// Test with missing argument
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")

	// Test with too many arguments
	cmd.SetArgs([]string{"cmd1", "cmd2"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 2")
}

// Note: Full integration tests for info command would require setting up
// actual commands in a test environment. The core functionality is tested
// through the integration tests when running the actual commands.

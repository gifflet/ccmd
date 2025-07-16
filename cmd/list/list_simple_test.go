/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandStructure(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List all commands managed by ccmd", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	longFlag := cmd.Flags().Lookup("long")
	assert.NotNil(t, longFlag)
	assert.Equal(t, "l", longFlag.Shorthand)
	assert.Equal(t, "Show detailed output including metadata", longFlag.Usage)
	assert.Equal(t, "false", longFlag.DefValue)
}

// Note: Formatting functions are not exported, so they can't be tested directly.
// They are tested indirectly through integration tests.

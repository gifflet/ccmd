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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommitHashUpdate(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"valid short hash", "a76c963", true},
		{"valid full hash", "a76c96359914b84ed1bcdbc11df03e6313e09ecf", true},
		{"tag version", "v1.0.0", false},
		{"branch name", "main", false},
		{"branch with slash", "feature/test", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommitHash(tt.version)
			assert.Equal(t, tt.expected, result, "isCommitHash(%q) should return %v", tt.version, tt.expected)
		})
	}
}

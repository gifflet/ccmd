// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package metadata

import (
	"github.com/gifflet/ccmd/internal/models"
)

// defaultManager is a package-level instance for convenience functions
var defaultManager = NewManager()

// ReadCommandMetadata reads and parses a command metadata file from the specified directory
func ReadCommandMetadata(commandDir string) (*models.CommandMetadata, error) {
	return defaultManager.ReadCommandMetadata(commandDir)
}

// WriteCommandMetadata writes command metadata to a file in the specified directory
func WriteCommandMetadata(commandDir string, metadata *models.CommandMetadata) error {
	return defaultManager.WriteCommandMetadata(commandDir, metadata)
}

// UpdateCommandMetadata reads existing metadata, applies updates, and writes it back
func UpdateCommandMetadata(commandDir string, updates func(*models.CommandMetadata) error) error {
	return defaultManager.UpdateCommandMetadata(commandDir, updates)
}

// ExtractCommandInfo extracts basic command information from metadata
func ExtractCommandInfo(metadata *models.CommandMetadata) map[string]interface{} {
	return defaultManager.ExtractCommandInfo(metadata)
}

// Exists checks if a metadata file exists in the specified directory
func Exists(commandDir string) bool {
	return defaultManager.Exists(commandDir)
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package project

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/errors"
)

// writeYAMLFile writes data to a file atomically with the specified permissions
func writeYAMLFile(filepath string, data interface{}, perm os.FileMode, fileSystem fs.FileSystem) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to marshal data: %v", err))
	}

	// Write atomically using a temporary file
	tempFile := filepath + ".tmp"
	if err := fileSystem.WriteFile(tempFile, yamlData, perm); err != nil {
		return errors.FileError("write temporary file", tempFile, err)
	}

	if err := fileSystem.Rename(tempFile, filepath); err != nil {
		// Clean up temp file on failure
		_ = fileSystem.Remove(tempFile) //nolint:errcheck // Best effort cleanup
		return errors.FileError("save file", filepath, err)
	}

	return nil
}

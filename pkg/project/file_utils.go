package project

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
)

// writeYAMLFile writes data to a file atomically with the specified permissions
func writeYAMLFile(filepath string, data interface{}, perm os.FileMode, fileSystem fs.FileSystem) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write atomically using a temporary file
	tempFile := filepath + ".tmp"
	if err := fileSystem.WriteFile(tempFile, yamlData, perm); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := fileSystem.Rename(tempFile, filepath); err != nil {
		// Clean up temp file on failure
		_ = fileSystem.Remove(tempFile) //nolint:errcheck // Best effort cleanup
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

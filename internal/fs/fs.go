// Package fs provides file system operations for ccmd
package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// validateFilePath validates that a file path is safe to read
func validateFilePath(path string) error {
	// Convert to absolute path for validation
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check for suspicious patterns that could indicate path traversal
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path contains suspicious traversal patterns: %s", path)
	}

	// Additional safety check: ensure path doesn't contain null bytes
	if strings.ContainsRune(path, 0) {
		return fmt.Errorf("path contains null byte: %s", path)
	}

	return nil
}

// safeReadFile safely reads a file after validating the path
func safeReadFile(path string) ([]byte, error) {
	if err := validateFilePath(path); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	return os.ReadFile(path) //nolint:gosec // Path is validated above
}

// GetClaudeCommandsDir returns the path to .claude/commands/ directory
// Creates the directory if it doesn't exist
func GetClaudeCommandsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "commands")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(claudeDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create .claude/commands directory: %w", err)
	}

	return claudeDir, nil
}

// ReadYAMLFile reads and unmarshals a YAML file
func ReadYAMLFile(path string, v interface{}) error {
	data, err := safeReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", path, err)
	}

	return nil
}

// WriteYAMLFile marshals and writes data to a YAML file
func WriteYAMLFile(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// ReadJSONFile reads and unmarshals a JSON file
func ReadJSONFile(path string, v interface{}) error {
	data, err := safeReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

// WriteJSONFile marshals and writes data to a JSON file with indentation
func WriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDir creates a directory with all necessary parents
func CreateDir(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// RemoveFile removes a file
func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", path, err)
	}
	return nil
}

// RemoveDir removes a directory and all its contents
func RemoveDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}

// JoinPath joins path elements safely across platforms
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// GetWorkingDir returns the current working directory
func GetWorkingDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

// NewOSFileSystem returns a new OS file system implementation
func NewOSFileSystem() FileSystem {
	return OS{}
}

// NewMemoryFileSystem returns a new in-memory file system implementation
func NewMemoryFileSystem() FileSystem {
	return NewMemFS()
}

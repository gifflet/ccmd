/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package fs provides file system operations for ccmd
package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
)

// validateFilePath validates that a file path is safe to read
func validateFilePath(path string) error {
	// Clean the path to remove any ".." or other traversal attempts
	cleaned := filepath.Clean(path)
	if cleaned != path {
		return errors.InvalidInput(fmt.Sprintf("path contains invalid characters or traversal patterns: %s", path))
	}
	return nil
}

// safeReadFile safely reads a file after validating the path
func safeReadFile(path string) ([]byte, error) {
	if err := validateFilePath(path); err != nil {
		return nil, err
	}
	return os.ReadFile(path) //nolint:gosec // Path is validated above
}

// GetClaudeCommandsDir returns the path to .claude/commands/ directory
// Creates the directory if it doesn't exist
// Deprecated: This function uses the home directory instead of project directory.
// Use project-relative paths (.claude) instead.
func GetClaudeCommandsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.FileError("get home directory", "", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", "commands")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(claudeDir, 0o750); err != nil {
		return "", errors.FileError("create .claude/commands directory", claudeDir, err)
	}

	return claudeDir, nil
}

// ReadYAMLFile reads and unmarshals a YAML file
func ReadYAMLFile(path string, v interface{}) error {
	data, err := safeReadFile(path)
	if err != nil {
		return errors.FileError("read file", path, err)
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return errors.FileError("unmarshal YAML", path, err)
	}

	return nil
}

// WriteYAMLFile marshals and writes data to a YAML file
func WriteYAMLFile(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to marshal YAML: %v", err))
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return errors.FileError("write file", path, err)
	}

	return nil
}

// ReadJSONFile reads and unmarshals a JSON file
func ReadJSONFile(path string, v interface{}) error {
	data, err := safeReadFile(path)
	if err != nil {
		return errors.FileError("read file", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return errors.FileError("unmarshal JSON", path, err)
	}

	return nil
}

// WriteJSONFile marshals and writes data to a JSON file with indentation
func WriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to marshal JSON: %v", err))
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return errors.FileError("write file", path, err)
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
		return errors.FileError("create directory", path, err)
	}
	return nil
}

// RemoveFile removes a file
func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil {
		return errors.FileError("remove file", path, err)
	}
	return nil
}

// RemoveDir removes a directory and all its contents
func RemoveDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.FileError("remove directory", path, err)
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
		return "", errors.FileError("get working directory", "", err)
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

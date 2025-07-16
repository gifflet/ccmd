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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupTestDir configures a temporary directory for tests and returns a cleanup function
func setupTestDir(t *testing.T) func() {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	return func() {
		os.Chdir(oldDir)
	}
}

// writeLockFile creates and writes a lock file
func writeLockFile(t *testing.T, lockFile *LockFile) {
	data, err := yaml.Marshal(lockFile)
	require.NoError(t, err)
	err = os.WriteFile("ccmd-lock.yaml", data, 0644)
	require.NoError(t, err)
}

// readLockFile reads and unmarshals a lock file
func readLockFile(t *testing.T) *LockFile {
	data, err := os.ReadFile("ccmd-lock.yaml")
	require.NoError(t, err)

	var lockFile LockFile
	err = yaml.Unmarshal(data, &lockFile)
	require.NoError(t, err)

	return &lockFile
}

// createCommandStructure creates directory and .md file for a command
func createCommandStructure(t *testing.T, commandName string) {
	err := os.MkdirAll(filepath.Join(".claude", "commands", commandName), 0755)
	require.NoError(t, err)

	mdContent := "# " + commandName
	err = os.WriteFile(filepath.Join(".claude", "commands", commandName+".md"), []byte(mdContent), 0644)
	require.NoError(t, err)
}

// writeConfig creates and writes a ccmd.yaml configuration file
func writeConfig(t *testing.T, commands []string) {
	config := map[string]interface{}{
		"commands": commands,
	}

	data, err := yaml.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile("ccmd.yaml", data, 0644)
	require.NoError(t, err)
}

// writeConfigMap writes a generic config map to ccmd.yaml
func writeConfigMap(t *testing.T, config map[string]interface{}) {
	data, err := yaml.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile("ccmd.yaml", data, 0644)
	require.NoError(t, err)
}

// createTestLockCommand creates a LockCommand with test defaults
func createTestLockCommand(name, version, source string) *LockCommand {
	return &LockCommand{
		Name:        name,
		Version:     version,
		Source:      source,
		Resolved:    source + "@" + version,
		Commit:      "abc123",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createBasicLockFile creates a basic lock file structure
func createBasicLockFile() *LockFile {
	return &LockFile{
		Version:         "1.0",
		LockfileVersion: 1,
		Commands:        make(map[string]*LockCommand),
	}
}

// writeLockFileToPath writes a lock file to a specific path
func writeLockFileToPath(t *testing.T, path string, lockFile *LockFile) {
	data, err := yaml.Marshal(lockFile)
	require.NoError(t, err)
	err = os.WriteFile(path, data, 0644)
	require.NoError(t, err)
}

// readLockFileFromPath reads a lock file from a specific path
func readLockFileFromPath(t *testing.T, path string) *LockFile {
	lockFile, err := ReadLockFile(path)
	require.NoError(t, err)
	return lockFile
}

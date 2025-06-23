/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package commands

import (
	"path/filepath"
	"testing"

	"github.com/gifflet/ccmd/internal/fs"
)

func TestListWithMetadata(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemFS()
	baseDir := "/test"

	// Create lock file with commands
	lockContent := `version: "1.0"
lockfileVersion: 1
commands:
  test-cmd:
    name: test-cmd
    version: 1.0.0
    source: github.com/user/test-cmd
    resolved: github.com/user/test-cmd@1.0.0
    commit: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z
  no-metadata-cmd:
    name: no-metadata-cmd
    version: 0.5.0
    source: github.com/user/no-metadata
    resolved: github.com/user/no-metadata@0.5.0
    commit: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z`
	// Write lock file directly with YAML content
	if err := mockFS.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}
	lockPath := filepath.Join(baseDir, "ccmd-lock.yaml")
	if err := mockFS.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create command directories and files
	commandsDir := filepath.Join(baseDir, ".claude", "commands")
	if err := mockFS.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create test-cmd with metadata
	testCmdDir := filepath.Join(commandsDir, "test-cmd")
	if err := mockFS.MkdirAll(testCmdDir, 0o755); err != nil {
		t.Fatalf("Failed to create test-cmd directory: %v", err)
	}

	// Create markdown file
	mdPath := filepath.Join(commandsDir, "test-cmd.md")
	if err := mockFS.WriteFile(mdPath, []byte("# test-cmd\n"), 0o644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Create ccmd.yaml with metadata
	metadataContent := `name: test-cmd
version: 1.2.0
description: A test command with metadata
author: Test Author
repository: github.com/user/test-cmd
entry: cmd/main.go
tags:
  - cli
  - testing
license: MIT
homepage: https://test-cmd.example.com
`
	metadataPath := filepath.Join(testCmdDir, "ccmd.yaml")
	if err := mockFS.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	// Create no-metadata-cmd with directory but no metadata
	noMetadataCmdDir := filepath.Join(commandsDir, "no-metadata-cmd")
	if err := mockFS.MkdirAll(noMetadataCmdDir, 0o755); err != nil {
		t.Fatalf("Failed to create no-metadata-cmd directory: %v", err)
	}
	mdPath2 := filepath.Join(commandsDir, "no-metadata-cmd.md")
	if err := mockFS.WriteFile(mdPath2, []byte("# no-metadata-cmd\n"), 0o644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Run List command
	opts := ListOptions{
		BaseDir:    baseDir,
		FileSystem: mockFS,
	}

	details, err := List(opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify results
	if len(details) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(details))
	}

	// Find test-cmd in results
	var testCmd *CommandDetail
	var noMetadataCmd *CommandDetail
	for _, d := range details {
		if d.Name == "test-cmd" {
			testCmd = d
		} else if d.Name == "no-metadata-cmd" {
			noMetadataCmd = d
		}
	}

	if testCmd == nil {
		t.Fatal("test-cmd not found in results")
	}
	if noMetadataCmd == nil {
		t.Fatal("no-metadata-cmd not found in results")
	}

	// Verify test-cmd has metadata
	if testCmd.CommandMetadata == nil {
		t.Error("test-cmd should have metadata")
	} else {
		// Verify metadata fields
		if testCmd.CommandMetadata.Version != "1.2.0" {
			t.Errorf("Expected metadata version 1.2.0, got %s", testCmd.CommandMetadata.Version)
		}
		if testCmd.CommandMetadata.Description != "A test command with metadata" {
			t.Errorf("Expected description 'A test command with metadata', got %s", testCmd.CommandMetadata.Description)
		}
		if testCmd.CommandMetadata.Author != "Test Author" {
			t.Errorf("Expected author 'Test Author', got %s", testCmd.CommandMetadata.Author)
		}
		if testCmd.CommandMetadata.Entry != "cmd/main.go" {
			t.Errorf("Expected entry 'cmd/main.go', got %s", testCmd.CommandMetadata.Entry)
		}
		if len(testCmd.CommandMetadata.Tags) != 2 || testCmd.CommandMetadata.Tags[0] != "cli" {
			t.Errorf("Expected tags [cli, testing], got %v", testCmd.CommandMetadata.Tags)
		}
		if testCmd.CommandMetadata.License != "MIT" {
			t.Errorf("Expected license 'MIT', got %s", testCmd.CommandMetadata.License)
		}
		if testCmd.CommandMetadata.Homepage != "https://test-cmd.example.com" {
			t.Errorf("Expected homepage 'https://test-cmd.example.com', got %s", testCmd.CommandMetadata.Homepage)
		}
	}

	// Verify structure validation
	if !testCmd.StructureValid {
		t.Error("test-cmd should have valid structure")
	}
	if !testCmd.HasDirectory {
		t.Error("test-cmd should have directory")
	}
	if !testCmd.HasMarkdownFile {
		t.Error("test-cmd should have markdown file")
	}

	// Verify no-metadata-cmd has no metadata
	if noMetadataCmd.CommandMetadata != nil {
		t.Error("no-metadata-cmd should not have metadata")
	}
	if !noMetadataCmd.StructureValid {
		t.Error("no-metadata-cmd should have valid structure")
	}
}

func TestListWithInvalidMetadata(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemFS()
	baseDir := "/test"

	// Create lock file
	lockContent := `version: "1.0"
lockfileVersion: 1
commands:
  invalid-metadata-cmd:
    name: invalid-metadata-cmd
    version: 1.0.0
    source: github.com/user/invalid
    resolved: github.com/user/invalid@1.0.0
    commit: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z`
	// Write lock file directly with YAML content
	if err := mockFS.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}
	lockPath := filepath.Join(baseDir, "ccmd-lock.yaml")
	if err := mockFS.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create command directory
	commandsDir := filepath.Join(baseDir, ".claude", "commands")
	cmdDir := filepath.Join(commandsDir, "invalid-metadata-cmd")
	if err := mockFS.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("Failed to create command directory: %v", err)
	}

	// Create markdown file
	mdPath := filepath.Join(commandsDir, "invalid-metadata-cmd.md")
	if err := mockFS.WriteFile(mdPath, []byte("# invalid-metadata-cmd\n"), 0o644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Create invalid ccmd.yaml (missing required fields)
	metadataContent := `name: invalid-metadata-cmd
description: Missing required fields
`
	metadataPath := filepath.Join(cmdDir, "ccmd.yaml")
	if err := mockFS.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	// Run List command
	opts := ListOptions{
		BaseDir:    baseDir,
		FileSystem: mockFS,
	}

	details, err := List(opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(details) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(details))
	}

	// Command should be listed but without metadata due to validation failure
	cmd := details[0]
	if cmd.CommandMetadata != nil {
		t.Error("Command with invalid metadata should not have CommandMetadata populated")
	}
	if !cmd.StructureValid {
		t.Error("Command should still have valid structure even with invalid metadata")
	}
}

func TestListFiltersByLockFile(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemFS()
	baseDir := "/test"

	// Create lock file with only one command
	lockContent := `version: "1.0"
lockfileVersion: 1
commands:
  tracked-cmd:
    name: tracked-cmd
    version: 1.0.0
    source: github.com/user/tracked
    resolved: github.com/user/tracked@1.0.0
    commit: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z`
	// Write lock file directly with YAML content
	if err := mockFS.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}
	lockPath := filepath.Join(baseDir, "ccmd-lock.yaml")
	if err := mockFS.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create command directories
	commandsDir := filepath.Join(baseDir, "commands")

	// Create tracked command
	trackedDir := filepath.Join(commandsDir, "tracked-cmd")
	if err := mockFS.MkdirAll(trackedDir, 0o755); err != nil {
		t.Fatalf("Failed to create tracked directory: %v", err)
	}
	mdPath := filepath.Join(commandsDir, "tracked-cmd.md")
	if err := mockFS.WriteFile(mdPath, []byte("# tracked-cmd\n"), 0o644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Create untracked command (not in lock file)
	untrackedDir := filepath.Join(commandsDir, "untracked-cmd")
	if err := mockFS.MkdirAll(untrackedDir, 0o755); err != nil {
		t.Fatalf("Failed to create untracked directory: %v", err)
	}
	mdPath2 := filepath.Join(commandsDir, "untracked-cmd.md")
	if err := mockFS.WriteFile(mdPath2, []byte("# untracked-cmd\n"), 0o644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Run List command
	opts := ListOptions{
		BaseDir:    baseDir,
		FileSystem: mockFS,
	}

	details, err := List(opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should only return tracked command
	if len(details) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(details))
	}

	if details[0].Name != "tracked-cmd" {
		t.Errorf("Expected tracked-cmd, got %s", details[0].Name)
	}
}

func TestListHandlesMissingFiles(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemFS()
	baseDir := "/test"

	// Create lock file with a command that has missing files
	lockContent := `version: "1.0"
lockfileVersion: 1
commands:
  missing-files-cmd:
    name: missing-files-cmd
    version: 1.0.0
    source: github.com/user/missing
    resolved: github.com/user/missing@1.0.0
    commit: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z`
	// Write lock file directly with YAML content
	if err := mockFS.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}
	lockPath := filepath.Join(baseDir, "ccmd-lock.yaml")
	if err := mockFS.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Don't create command directory or markdown file

	// Run List command
	opts := ListOptions{
		BaseDir:    baseDir,
		FileSystem: mockFS,
	}

	details, err := List(opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(details) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(details))
	}

	cmd := details[0]
	if cmd.StructureValid {
		t.Error("Command with missing files should not have valid structure")
	}
	if cmd.HasDirectory {
		t.Error("Command should not have directory")
	}
	if cmd.HasMarkdownFile {
		t.Error("Command should not have markdown file")
	}
	if cmd.CommandMetadata != nil {
		t.Error("Command should not have metadata when directory is missing")
	}
}

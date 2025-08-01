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
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/output"
)

func TestListCommandIntegration(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemFS()
	baseDir := ".claude"

	// Create lock file with test data
	lockContent := `version: "1.0"
lockfileVersion: 1
commands:
    hello-world:
        name: hello-world
        version: 1.0.0
        source: git@github.com:gifflet/hello-world.git
        resolved: git@github.com:gifflet/hello-world.git@v1.0.0
        commit: 05d746d17f6e2235ad9a93acc307b68caa18a281
        installed_at: 2025-06-22T01:07:51.524358-03:00
        updated_at: 2025-06-22T01:07:51.524358-03:00
    dev-tools:
        name: dev-tools
        version: 2.5.0
        source: github.com/user/dev-tools
        resolved: github.com/user/dev-tools@2.5.0
		commit: 151213c3e25a85f8a5ca12ce8cc79d3b80f5850e
        installed_at: 2024-01-02T10:00:00Z
        updated_at: 2024-01-03T10:00:00Z
`

	// Setup filesystem
	lockPath := "ccmd-lock.yaml"
	if err := mockFS.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatalf("Failed to write lock file: %v", err)
	}

	// Create command directories and files
	commandsDir := filepath.Join(baseDir, "commands")

	// Create hello-world command with metadata
	helloDir := filepath.Join(commandsDir, "hello-world")
	if err := mockFS.MkdirAll(helloDir, 0o755); err != nil {
		t.Fatalf("Failed to create hello-world directory: %v", err)
	}
	helloMd := filepath.Join(commandsDir, "hello-world.md")
	if err := mockFS.WriteFile(helloMd, []byte("# Hello World\n"), 0o644); err != nil {
		t.Fatalf("Failed to create hello-world.md: %v", err)
	}

	// Create hello-world metadata
	helloMetadata := `name: hello-world
version: 1.0.1
description: A simple hello world command
author: Test Developer
repository: github.com/gifflet/hello-world
entry: main.go
tags:
  - example
  - beginner
license: MIT
`
	helloMetadataPath := filepath.Join(helloDir, "ccmd.yaml")
	if err := mockFS.WriteFile(helloMetadataPath, []byte(helloMetadata), 0o644); err != nil {
		t.Fatalf("Failed to write hello-world metadata: %v", err)
	}

	// Create dev-tools command (without metadata)
	devDir := filepath.Join(commandsDir, "dev-tools")
	if err := mockFS.MkdirAll(devDir, 0o755); err != nil {
		t.Fatalf("Failed to create dev-tools directory: %v", err)
	}
	devMd := filepath.Join(commandsDir, "dev-tools.md")
	if err := mockFS.WriteFile(devMd, []byte("# Dev Tools\n"), 0o644); err != nil {
		t.Fatalf("Failed to create dev-tools.md: %v", err)
	}

	// Test cases
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "simple list",
			args: []string{},
		},
		{
			name: "long format",
			args: []string{"--long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test passes if the command runs without error
			cmd := NewCommand()
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				long, _ := cmd.Flags().GetBool("long")
				// Mock the filesystem for testing
				return runListWithFS(long, baseDir, mockFS)
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// runListWithFS is a test helper that allows injecting a custom filesystem
func runListWithFS(long bool, baseDir string, filesystem fs.FileSystem) error {
	// Get detailed command information
	opts := core.ListOptions{
		ProjectPath: ".",
	}
	details, err := core.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(details) == 0 {
		output.PrintInfof("No commands installed yet.")
		output.PrintInfof("Use 'ccmd install' to install commands.")
		return nil
	}

	// Sort by name
	sort.Slice(details, func(i, j int) bool {
		return details[i].Name < details[j].Name
	})

	// Check for structure issues
	hasStructureIssues := false
	for _, detail := range details {
		if detail.BrokenStructure {
			hasStructureIssues = true
			break
		}
	}

	// Print table
	if long {
		printLongList(details)
	} else {
		printSimpleList(details)
	}

	// Show warning if there are structure issues
	if hasStructureIssues {
		output.PrintWarningf("\nSome commands have broken dual structure (missing directory or .md file).")
		output.PrintWarningf("Run with --long flag to see details.")
	}

	return nil
}

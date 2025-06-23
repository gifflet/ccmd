// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package project

import (
	"strings"
	"testing"

	"github.com/gifflet/ccmd/internal/fs"
)

func TestManager_PreserveFieldOrder(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemoryFileSystem()

	// Initial YAML content with specific field order
	initialContent := `name: test-project
version: 1.0.0
description: "Test project"
author: "Test Author"
repository: "test/repo"
entry: index.md
`

	// Write initial content
	if err := mockFS.WriteFile("ccmd.yaml", []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial content: %v", err)
	}

	// Create manager with mock filesystem
	m := &Manager{
		rootDir: ".",
		fs:      mockFS,
	}

	// Add a command
	if err := m.AddCommand("owner/repo", "v1.0.0"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	// Read the updated content
	updatedData, err := mockFS.ReadFile("ccmd.yaml")
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	updatedContent := string(updatedData)

	// Verify field order is preserved
	expectedOrder := []string{"name:", "version:", "description:", "author:", "repository:", "entry:", "commands:"}

	lastIndex := -1
	for _, field := range expectedOrder {
		index := strings.Index(updatedContent, field)
		if index == -1 {
			t.Errorf("field %s not found in updated content", field)
			continue
		}
		if index <= lastIndex {
			t.Errorf("field %s is out of order: found at index %d, previous was at %d", field, index, lastIndex)
		}
		lastIndex = index
	}

	// Verify commands field is at the end
	if !strings.Contains(updatedContent, "commands:") {
		t.Error("commands field not found")
	}

	// Check that commands appears after entry
	entryIndex := strings.Index(updatedContent, "entry:")
	commandsIndex := strings.Index(updatedContent, "commands:")
	if commandsIndex <= entryIndex {
		t.Errorf("commands field should be after entry field, but commands is at %d and entry is at %d", commandsIndex, entryIndex)
	}

	t.Logf("Updated content:\n%s", updatedContent)
}

func TestManager_CreateMinimalConfig(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMemoryFileSystem()

	// Create manager with mock filesystem
	m := &Manager{
		rootDir: ".",
		fs:      mockFS,
	}

	// Add a command to non-existent config
	if err := m.AddCommand("owner/repo", "v1.0.0"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	// Read the created content
	createdData, err := mockFS.ReadFile("ccmd.yaml")
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	createdContent := string(createdData)

	// Verify only commands field is present
	if strings.Contains(createdContent, "name:") {
		t.Error("name field should not be present in minimal config")
	}
	if strings.Contains(createdContent, "version:") {
		t.Error("version field should not be present in minimal config")
	}
	if strings.Contains(createdContent, "description:") {
		t.Error("description field should not be present in minimal config")
	}
	if strings.Contains(createdContent, "author:") {
		t.Error("author field should not be present in minimal config")
	}
	if strings.Contains(createdContent, "repository:") {
		t.Error("repository field should not be present in minimal config")
	}
	if strings.Contains(createdContent, "entry:") {
		t.Error("entry field should not be present in minimal config")
	}

	// Verify commands field is present
	if !strings.Contains(createdContent, "commands:") {
		t.Error("commands field not found in minimal config")
	}

	t.Logf("Created minimal config:\n%s", createdContent)
}

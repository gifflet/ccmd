// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetClaudeCommandsDir(t *testing.T) {
	dir, err := GetClaudeCommandsDir()
	if err != nil {
		t.Fatalf("GetClaudeCommandsDir() error = %v", err)
	}

	// Check if directory was created
	if !DirExists(dir) {
		t.Errorf("GetClaudeCommandsDir() directory not created: %s", dir)
	}

	// Check if path contains .claude/commands
	if !filepath.IsAbs(dir) {
		t.Errorf("GetClaudeCommandsDir() should return absolute path, got: %s", dir)
	}
}

func TestFileOperations(t *testing.T) {
	// Create temp dir for tests
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Test FileExists with non-existent file
	if FileExists(testFile) {
		t.Error("FileExists() should return false for non-existent file")
	}

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test FileExists with existing file
	if !FileExists(testFile) {
		t.Error("FileExists() should return true for existing file")
	}

	// Test RemoveFile
	if err := RemoveFile(testFile); err != nil {
		t.Errorf("RemoveFile() error = %v", err)
	}

	if FileExists(testFile) {
		t.Error("RemoveFile() file still exists after removal")
	}
}

func TestDirOperations(t *testing.T) {
	// Create temp dir for tests
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "subdir", "nested")

	// Test DirExists with non-existent dir
	if DirExists(testDir) {
		t.Error("DirExists() should return false for non-existent directory")
	}

	// Test CreateDir
	if err := CreateDir(testDir); err != nil {
		t.Errorf("CreateDir() error = %v", err)
	}

	// Test DirExists with existing dir
	if !DirExists(testDir) {
		t.Error("DirExists() should return true for existing directory")
	}

	// Test RemoveDir
	parentDir := filepath.Join(tempDir, "subdir")
	if err := RemoveDir(parentDir); err != nil {
		t.Errorf("RemoveDir() error = %v", err)
	}

	if DirExists(parentDir) {
		t.Error("RemoveDir() directory still exists after removal")
	}
}

func TestYAMLOperations(t *testing.T) {
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test.yaml")

	type TestStruct struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	// Test WriteYAMLFile
	original := TestStruct{Name: "test", Value: 42}
	if err := WriteYAMLFile(yamlFile, original); err != nil {
		t.Errorf("WriteYAMLFile() error = %v", err)
	}

	// Test ReadYAMLFile
	var loaded TestStruct
	if err := ReadYAMLFile(yamlFile, &loaded); err != nil {
		t.Errorf("ReadYAMLFile() error = %v", err)
	}

	if loaded.Name != original.Name || loaded.Value != original.Value {
		t.Errorf("YAML round-trip failed: got %+v, want %+v", loaded, original)
	}
}

func TestJSONOperations(t *testing.T) {
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test.json")

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test WriteJSONFile
	original := TestStruct{Name: "test", Value: 42}
	if err := WriteJSONFile(jsonFile, original); err != nil {
		t.Errorf("WriteJSONFile() error = %v", err)
	}

	// Test ReadJSONFile
	var loaded TestStruct
	if err := ReadJSONFile(jsonFile, &loaded); err != nil {
		t.Errorf("ReadJSONFile() error = %v", err)
	}

	if loaded.Name != original.Name || loaded.Value != original.Value {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", loaded, original)
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		want     string
	}{
		{
			name:     "simple join",
			elements: []string{"a", "b", "c"},
			want:     filepath.Join("a", "b", "c"),
		},
		{
			name:     "with dots",
			elements: []string{".", "subdir", "file.txt"},
			want:     filepath.Join(".", "subdir", "file.txt"),
		},
		{
			name:     "empty elements",
			elements: []string{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinPath(tt.elements...)
			if got != tt.want {
				t.Errorf("JoinPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWorkingDir(t *testing.T) {
	dir, err := GetWorkingDir()
	if err != nil {
		t.Errorf("GetWorkingDir() error = %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("GetWorkingDir() should return absolute path, got: %s", dir)
	}
}

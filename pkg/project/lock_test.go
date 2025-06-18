package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNewLockFile(t *testing.T) {
	lf := NewLockFile()

	if lf.Version != LockFileVersion {
		t.Errorf("expected version %s, got %s", LockFileVersion, lf.Version)
	}

	if lf.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	if lf.Commands == nil {
		t.Error("expected Commands map to be initialized")
	}
}

func TestCommand_Validate(t *testing.T) {
	validTime := time.Now()
	validCmd := &Command{
		Name:         "test-cmd",
		Repository:   "github.com/user/repo",
		Version:      "v1.0.0",
		CommitHash:   strings.Repeat("a", 40),
		InstalledAt:  validTime,
		UpdatedAt:    validTime,
		FileSize:     1024,
		Checksum:     strings.Repeat("b", 64),
		Dependencies: []string{"dep1", "dep2"},
		Metadata:     map[string]string{"key": "value"},
	}

	tests := []struct {
		name    string
		modify  func(*Command)
		wantErr string
	}{
		{
			name:    "valid command",
			modify:  func(c *Command) {},
			wantErr: "",
		},
		{
			name:    "missing name",
			modify:  func(c *Command) { c.Name = "" },
			wantErr: "name is required",
		},
		{
			name:    "missing repository",
			modify:  func(c *Command) { c.Repository = "" },
			wantErr: "repository is required",
		},
		{
			name:    "missing version",
			modify:  func(c *Command) { c.Version = "" },
			wantErr: "version is required",
		},
		{
			name:    "missing commit hash",
			modify:  func(c *Command) { c.CommitHash = "" },
			wantErr: "commit_hash is required",
		},
		{
			name:    "invalid commit hash length",
			modify:  func(c *Command) { c.CommitHash = "abc123" },
			wantErr: "commit_hash must be a 40-character SHA",
		},
		{
			name:    "missing installed_at",
			modify:  func(c *Command) { c.InstalledAt = time.Time{} },
			wantErr: "installed_at is required",
		},
		{
			name:    "missing updated_at",
			modify:  func(c *Command) { c.UpdatedAt = time.Time{} },
			wantErr: "updated_at is required",
		},
		{
			name:    "invalid file size",
			modify:  func(c *Command) { c.FileSize = 0 },
			wantErr: "file_size must be positive",
		},
		{
			name:    "missing checksum",
			modify:  func(c *Command) { c.Checksum = "" },
			wantErr: "checksum is required",
		},
		{
			name:    "invalid checksum length",
			modify:  func(c *Command) { c.Checksum = "abc123" },
			wantErr: "checksum must be a 64-character SHA256 hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := *validCmd // Create a copy
			tt.modify(&cmd)

			err := cmd.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestLockFile_AddCommand(t *testing.T) {
	lf := NewLockFile()
	cmd := &Command{
		Name:        "test-cmd",
		Repository:  "github.com/user/repo",
		Version:     "v1.0.0",
		CommitHash:  strings.Repeat("a", 40),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		FileSize:    1024,
		Checksum:    strings.Repeat("b", 64),
	}

	err := lf.AddCommand(cmd)
	if err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	if len(lf.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(lf.Commands))
	}

	stored, exists := lf.GetCommand("test-cmd")
	if !exists {
		t.Error("command not found")
	}

	if stored.Name != cmd.Name {
		t.Errorf("expected name %s, got %s", cmd.Name, stored.Name)
	}
}

func TestLockFile_RemoveCommand(t *testing.T) {
	lf := NewLockFile()
	cmd := &Command{
		Name:        "test-cmd",
		Repository:  "github.com/user/repo",
		Version:     "v1.0.0",
		CommitHash:  strings.Repeat("a", 40),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		FileSize:    1024,
		Checksum:    strings.Repeat("b", 64),
	}

	lf.AddCommand(cmd)

	if !lf.RemoveCommand("test-cmd") {
		t.Error("expected RemoveCommand to return true")
	}

	if len(lf.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(lf.Commands))
	}

	if lf.RemoveCommand("non-existent") {
		t.Error("expected RemoveCommand to return false for non-existent command")
	}
}

func TestLockFile_Validate(t *testing.T) {
	tests := []struct {
		name     string
		lockFile *LockFile
		wantErr  string
	}{
		{
			name: "valid lock file",
			lockFile: &LockFile{
				Version:   LockFileVersion,
				UpdatedAt: time.Now(),
				Commands:  make(map[string]*Command),
			},
			wantErr: "",
		},
		{
			name: "missing version",
			lockFile: &LockFile{
				UpdatedAt: time.Now(),
				Commands:  make(map[string]*Command),
			},
			wantErr: "version is required",
		},
		{
			name: "command name mismatch",
			lockFile: &LockFile{
				Version:   LockFileVersion,
				UpdatedAt: time.Now(),
				Commands: map[string]*Command{
					"cmd1": {
						Name:        "cmd2",
						Repository:  "github.com/user/repo",
						Version:     "v1.0.0",
						CommitHash:  strings.Repeat("a", 40),
						InstalledAt: time.Now(),
						UpdatedAt:   time.Now(),
						FileSize:    1024,
						Checksum:    strings.Repeat("b", 64),
					},
				},
			},
			wantErr: "command name mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lockFile.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestLockFile_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	lockFilePath := filepath.Join(tmpDir, "ccmd-lock.yaml")

	// Create a lock file with some commands
	lf := NewLockFile()
	cmd1 := &Command{
		Name:         "cmd1",
		Repository:   "github.com/user/repo1",
		Version:      "v1.0.0",
		CommitHash:   strings.Repeat("a", 40),
		InstalledAt:  time.Now().Truncate(time.Second), // Truncate for comparison
		UpdatedAt:    time.Now().Truncate(time.Second),
		FileSize:     1024,
		Checksum:     strings.Repeat("b", 64),
		Dependencies: []string{"dep1", "dep2"},
		Metadata:     map[string]string{"key": "value"},
	}
	cmd2 := &Command{
		Name:        "cmd2",
		Repository:  "github.com/user/repo2",
		Version:     "v2.0.0",
		CommitHash:  strings.Repeat("c", 40),
		InstalledAt: time.Now().Truncate(time.Second),
		UpdatedAt:   time.Now().Truncate(time.Second),
		FileSize:    2048,
		Checksum:    strings.Repeat("d", 64),
	}

	lf.AddCommand(cmd1)
	lf.AddCommand(cmd2)

	// Save to file
	if err := lf.SaveToFile(lockFilePath); err != nil {
		t.Fatalf("failed to save lock file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(lockFilePath); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}

	// Load from file
	loaded, err := LoadFromFile(lockFilePath)
	if err != nil {
		t.Fatalf("failed to load lock file: %v", err)
	}

	// Verify loaded content
	if loaded.Version != lf.Version {
		t.Errorf("expected version %s, got %s", lf.Version, loaded.Version)
	}

	if len(loaded.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(loaded.Commands))
	}

	// Verify cmd1
	loadedCmd1, exists := loaded.GetCommand("cmd1")
	if !exists {
		t.Error("cmd1 not found in loaded file")
	} else {
		if loadedCmd1.Repository != cmd1.Repository {
			t.Errorf("cmd1: expected repository %s, got %s", cmd1.Repository, loadedCmd1.Repository)
		}
		if len(loadedCmd1.Dependencies) != 2 {
			t.Errorf("cmd1: expected 2 dependencies, got %d", len(loadedCmd1.Dependencies))
		}
		if loadedCmd1.Metadata["key"] != "value" {
			t.Errorf("cmd1: expected metadata key=value, got %v", loadedCmd1.Metadata)
		}
	}

	// Verify cmd2
	loadedCmd2, exists := loaded.GetCommand("cmd2")
	if !exists {
		t.Error("cmd2 not found in loaded file")
	} else {
		if loadedCmd2.FileSize != cmd2.FileSize {
			t.Errorf("cmd2: expected file size %d, got %d", cmd2.FileSize, loadedCmd2.FileSize)
		}
	}
}

func TestLockFile_YAMLFormat(t *testing.T) {
	lf := NewLockFile()
	lf.UpdatedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	cmd := &Command{
		Name:         "test-cmd",
		Repository:   "github.com/user/repo",
		Version:      "v1.0.0",
		CommitHash:   strings.Repeat("a", 40),
		InstalledAt:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		FileSize:     1024,
		Checksum:     strings.Repeat("b", 64),
		Dependencies: []string{"dep1", "dep2"},
		Metadata: map[string]string{
			"arch": "amd64",
			"os":   "linux",
		},
	}
	lf.AddCommand(cmd)

	// Marshal to YAML
	data, err := yaml.Marshal(lf)
	if err != nil {
		t.Fatalf("failed to marshal to YAML: %v", err)
	}

	yamlStr := string(data)

	// Verify YAML structure contains expected fields
	expectedFields := []string{
		"version:",
		"updated_at:",
		"commands:",
		"test-cmd:",
		"repository:",
		"commit_hash:",
		"installed_at:",
		"file_size:",
		"checksum:",
		"dependencies:",
		"metadata:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(yamlStr, field) {
			t.Errorf("expected YAML to contain %q", field)
		}
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create a temporary file with known content
	tmpFile, err := os.CreateTemp("", "checksum-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("Hello, World!")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Calculate checksum
	checksum, err := CalculateChecksum(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	// Expected SHA256 of "Hello, World!"
	expected := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if checksum != expected {
		t.Errorf("expected checksum %s, got %s", expected, checksum)
	}

	// Test with non-existent file
	_, err = CalculateChecksum("/non/existent/file")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLockFile_AddCommand_InvalidCommand(t *testing.T) {
	lf := NewLockFile()

	// Test with invalid command
	invalidCmd := &Command{
		Name: "invalid", // Missing required fields
	}

	err := lf.AddCommand(invalidCmd)
	if err == nil {
		t.Error("expected error for invalid command")
	}
	if !strings.Contains(err.Error(), "invalid command") {
		t.Errorf("expected error to contain 'invalid command', got: %v", err)
	}
}

func TestLockFile_AddCommand_NilCommandsMap(t *testing.T) {
	lf := &LockFile{
		Version:   LockFileVersion,
		UpdatedAt: time.Now(),
		Commands:  nil, // Start with nil map
	}

	cmd := &Command{
		Name:        "test-cmd",
		Repository:  "github.com/user/repo",
		Version:     "v1.0.0",
		CommitHash:  strings.Repeat("a", 40),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		FileSize:    1024,
		Checksum:    strings.Repeat("b", 64),
	}

	err := lf.AddCommand(cmd)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify Commands map was initialized
	if lf.Commands == nil {
		t.Error("expected Commands map to be initialized")
	}

	if _, exists := lf.Commands["test-cmd"]; !exists {
		t.Error("expected command to be added")
	}
}

func TestLockFile_Validate_NilCommands(t *testing.T) {
	lf := &LockFile{
		Version:   LockFileVersion,
		UpdatedAt: time.Now(),
		Commands:  nil, // Test nil Commands map
	}

	err := lf.Validate()
	if err != nil {
		t.Errorf("expected no error for nil Commands map, got: %v", err)
	}

	// After validation, Commands should be initialized
	if lf.Commands == nil {
		t.Error("expected Commands map to be initialized after Validate")
	}
}

func TestLockFile_SaveToFile_InvalidLockFile(t *testing.T) {
	lf := &LockFile{
		// Missing required version field
		UpdatedAt: time.Now(),
		Commands:  make(map[string]*Command),
	}

	tmpDir := t.TempDir()
	lockFilePath := filepath.Join(tmpDir, "invalid-lock.yaml")

	err := lf.SaveToFile(lockFilePath)
	if err == nil {
		t.Error("expected error for invalid lock file")
	}
	if !strings.Contains(err.Error(), "invalid lock file") {
		t.Errorf("expected error to contain 'invalid lock file', got: %v", err)
	}
}

func TestLockFile_SaveToFile_AtomicWriteFailure(t *testing.T) {
	lf := NewLockFile()

	// Use a directory path instead of file to trigger write error
	err := lf.SaveToFile("/")
	if err == nil {
		t.Error("expected error when saving to invalid path")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	invalidYAMLPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML content
	invalidContent := []byte("invalid: yaml: content: [")
	if err := os.WriteFile(invalidYAMLPath, invalidContent, 0o600); err != nil {
		t.Fatalf("failed to write invalid YAML file: %v", err)
	}

	_, err := LoadFromFile(invalidYAMLPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse lock file") {
		t.Errorf("expected error to contain 'failed to parse lock file', got: %v", err)
	}
}

func TestLoadFromFile_InvalidLockFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	invalidLockPath := filepath.Join(tmpDir, "invalid-lock.yaml")

	// Write valid YAML but invalid lock file structure
	invalidContent := []byte(`version: ""
commands:
  test:
    name: test
    repository: ""`)

	if err := os.WriteFile(invalidLockPath, invalidContent, 0o600); err != nil {
		t.Fatalf("failed to write invalid lock file: %v", err)
	}

	_, err := LoadFromFile(invalidLockPath)
	if err == nil {
		t.Error("expected error for invalid lock file content")
	}
	if !strings.Contains(err.Error(), "invalid lock file") {
		t.Errorf("expected error to contain 'invalid lock file', got: %v", err)
	}
}

func TestLockFile_Validate_InvalidCommand(t *testing.T) {
	lf := &LockFile{
		Version:   LockFileVersion,
		UpdatedAt: time.Now(),
		Commands: map[string]*Command{
			"invalid": {
				Name: "invalid",
				// Missing required fields
			},
		},
	}

	err := lf.Validate()
	if err == nil {
		t.Error("expected error for invalid command")
	}
	if !strings.Contains(err.Error(), "invalid command") {
		t.Errorf("expected error to contain 'invalid command', got: %v", err)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	_, err := LoadFromFile("/non/existent/file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to read lock file") {
		t.Errorf("expected error to contain 'failed to read lock file', got: %v", err)
	}
}

func TestCalculateChecksum_FileCloseError(t *testing.T) {
	// This test is to ensure the deferred close error handling is covered
	// The error is intentionally ignored, so we just ensure the function runs
	tmpFile, err := os.CreateTemp("", "checksum-close-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Write some content
	if err := os.WriteFile(tmpPath, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Calculate checksum - this will open and close the file
	_, err = CalculateChecksum(tmpPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

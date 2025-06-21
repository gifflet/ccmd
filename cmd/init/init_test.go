package init

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestPromptUser(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		defaultValue string
		input        string
		expected     string
	}{
		{
			name:         "uses default when empty",
			prompt:       "name",
			defaultValue: "default-name",
			input:        "",
			expected:     "default-name",
		},
		{
			name:         "uses input when provided",
			prompt:       "name",
			defaultValue: "default-name",
			input:        "custom-name",
			expected:     "custom-name",
		},
		{
			name:         "handles no default",
			prompt:       "description",
			defaultValue: "",
			input:        "my description",
			expected:     "my description",
		},
		{
			name:         "trims whitespace",
			prompt:       "author",
			defaultValue: "",
			input:        "  John Doe  ",
			expected:     "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input + "\n")
			scanner := bufio.NewScanner(reader)

			result := promptUser(scanner, tt.prompt, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("promptUser() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsConfirmation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"y", true},
		{"Y", true},
		{"", true},
		{"no", false},
		{"n", false},
		{"maybe", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isConfirmation(tt.input)
			if result != tt.expected {
				t.Errorf("isConfirmation(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRunInit(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Simulate user input
	input := `test-command
1.0.0
Test command description
Test Author
https://github.com/test/repo
index.md
automation, testing
yes
`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()
	defer func() { os.Stdin = oldStdin }()

	// Run the init command
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	// Check if .claude/commands directory was created
	claudeDir := filepath.Join(tempDir, ".claude", "commands")
	if !fs.DirExists(claudeDir) {
		t.Error(".claude/commands directory was not created")
	}

	// Check if ccmd.yaml was created
	ccmdPath := filepath.Join(tempDir, "ccmd.yaml")
	if !fs.FileExists(ccmdPath) {
		t.Error("ccmd.yaml was not created")
	}

	// Read and verify the content
	var metadata models.CommandMetadata
	if err := fs.ReadYAMLFile(ccmdPath, &metadata); err != nil {
		t.Fatalf("Failed to read ccmd.yaml: %v", err)
	}

	// Verify fields
	if metadata.Name != "test-command" {
		t.Errorf("Name = %v, want %v", metadata.Name, "test-command")
	}
	if metadata.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", metadata.Version, "1.0.0")
	}
	if metadata.Description != "Test command description" {
		t.Errorf("Description = %v, want %v", metadata.Description, "Test command description")
	}
	if metadata.Author != "Test Author" {
		t.Errorf("Author = %v, want %v", metadata.Author, "Test Author")
	}
	if metadata.Repository != "https://github.com/test/repo" {
		t.Errorf("Repository = %v, want %v", metadata.Repository, "https://github.com/test/repo")
	}
	if metadata.Entry != "index.md" {
		t.Errorf("Entry = %v, want %v", metadata.Entry, "index.md")
	}
	if len(metadata.Tags) != 2 || metadata.Tags[0] != "automation" || metadata.Tags[1] != "testing" {
		t.Errorf("Tags = %v, want %v", metadata.Tags, []string{"automation", "testing"})
	}
}

func TestRunInitCancelled(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Simulate user input with "no" confirmation
	input := `test-command
1.0.0



index.md

no
`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()
	defer func() { os.Stdin = oldStdin }()

	// Run the init command
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	// Check that ccmd.yaml was NOT created
	ccmdPath := filepath.Join(tempDir, "ccmd.yaml")
	if fs.FileExists(ccmdPath) {
		t.Error("ccmd.yaml should not have been created when cancelled")
	}
}

func TestRunInitDefaults(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Get directory name for default
	dirName := filepath.Base(tempDir)

	// Simulate user input using all defaults (empty lines for defaults)
	input := "\n\n\n\n\n\n\nyes\n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()
	defer func() { os.Stdin = oldStdin }()

	// Run the init command
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	// Read and verify the content
	var metadata models.CommandMetadata
	ccmdPath := filepath.Join(tempDir, "ccmd.yaml")
	if err := fs.ReadYAMLFile(ccmdPath, &metadata); err != nil {
		t.Fatalf("Failed to read ccmd.yaml: %v", err)
	}

	// Verify defaults
	if metadata.Name != dirName {
		t.Errorf("Name = %v, want %v", metadata.Name, dirName)
	}
	if metadata.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", metadata.Version, "1.0.0")
	}
	if metadata.Description != "" {
		t.Errorf("Description = %v, want empty", metadata.Description)
	}
	if metadata.Author != "" {
		t.Errorf("Author = %v, want empty", metadata.Author)
	}
	if metadata.Repository != "" {
		t.Errorf("Repository = %v, want empty", metadata.Repository)
	}
	if metadata.Entry != "index.md" {
		t.Errorf("Entry = %v, want %v", metadata.Entry, "index.md")
	}
	if len(metadata.Tags) != 0 {
		t.Errorf("Tags = %v, want empty", metadata.Tags)
	}
}

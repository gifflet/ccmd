package info

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "info <command-name>", cmd.Use)
	assert.Equal(t, "Display detailed information about an installed command", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	jsonFlag := cmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "false", jsonFlag.DefValue)
}

func TestRunInfo(t *testing.T) {
	// Setup test filesystem
	memfs := fs.NewMemFS()

	// Create base directory structure (project-local)
	baseDir := ".claude"
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands"), 0o755))

	// Create a test command
	commandName := "test-command"
	commandDir := filepath.Join(baseDir, "commands", commandName)
	require.NoError(t, memfs.MkdirAll(commandDir, 0o755))

	// Create ccmd.yaml
	metadata := models.CommandMetadata{
		Name:        commandName,
		Version:     "1.0.0",
		Description: "A test command",
		Author:      "Test Author",
		Repository:  "https://github.com/test/test-command",
		Tags:        []string{"test", "example"},
		License:     "MIT",
		Homepage:    "https://example.com",
		Entry:       "index.md",
	}

	yamlData, err := metadata.MarshalYAML()
	require.NoError(t, err)
	require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), yamlData, 0o644))

	// Create index.md
	indexContent := `# Test Command

This is a test command for demonstration purposes.

## Usage

` + "```bash" + `
test-command --help
` + "```" + `

## Features

- Feature 1
- Feature 2
- Feature 3
`
	require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "index.md"), []byte(indexContent), 0o644))

	// Create standalone .md file
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands", commandName+".md"), []byte("Test content"), 0o644))

	// Create lock file
	lockManager := lock.NewManagerWithFS(baseDir, memfs)
	require.NoError(t, lockManager.Load())
	cmd := &models.Command{
		Name:        commandName,
		Version:     "1.0.0",
		Source:      "https://github.com/test/test-command",
		InstalledAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		Metadata: map[string]string{
			"description": "A test command",
		},
	}
	require.NoError(t, lockManager.AddCommand(cmd))
	require.NoError(t, lockManager.Save())

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err = runInfoWithFS(commandName, false, memfs)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify no error
	assert.NoError(t, err)

	// Verify output contains expected information
	assert.Contains(t, output, "Command Information")
	assert.Contains(t, output, commandName)
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "Test Author")
	assert.Contains(t, output, "A test command")
	assert.Contains(t, output, "https://github.com/test/test-command")
	assert.Contains(t, output, "Installation Details")
	assert.Contains(t, output, "Structure Verification")
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, "Content Preview")
	assert.Contains(t, output, "# Test Command")
}

func TestRunInfoJSON(t *testing.T) {
	// Setup test filesystem
	memfs := fs.NewMemFS()

	// Create base directory structure (project-local)
	baseDir := ".claude"
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands"), 0o755))

	// Create a test command
	commandName := "json-test"
	commandDir := filepath.Join(baseDir, "commands", commandName)
	require.NoError(t, memfs.MkdirAll(commandDir, 0o755))

	// Create ccmd.yaml
	metadata := models.CommandMetadata{
		Name:        commandName,
		Version:     "2.0.0",
		Description: "JSON test command",
		Author:      "JSON Author",
		Repository:  "https://github.com/test/json-test",
	}

	yamlData, err := metadata.MarshalYAML()
	require.NoError(t, err)
	require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), yamlData, 0o644))

	// Create standalone .md file
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands", commandName+".md"), []byte("JSON test"), 0o644))

	// Create lock file
	lockManager := lock.NewManagerWithFS(baseDir, memfs)
	require.NoError(t, lockManager.Load())
	cmd := &models.Command{
		Name:        commandName,
		Version:     "2.0.0",
		Source:      "https://github.com/test/json-test",
		InstalledAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}
	require.NoError(t, lockManager.AddCommand(cmd))
	require.NoError(t, lockManager.Save())

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command with JSON output
	err = runInfoWithFS(commandName, true, memfs)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify no error
	assert.NoError(t, err)

	// Parse JSON output
	var info Output
	err = json.Unmarshal([]byte(output), &info)
	assert.NoError(t, err)

	// Verify JSON structure
	assert.Equal(t, commandName, info.Name)
	assert.Equal(t, "2.0.0", info.Version)
	assert.Equal(t, "JSON Author", info.Author)
	assert.Equal(t, "JSON test command", info.Description)
	assert.Equal(t, "https://github.com/test/json-test", info.Repository)
	assert.True(t, info.Structure.DirectoryExists)
	assert.True(t, info.Structure.MarkdownExists)
	assert.True(t, info.Structure.HasCcmdYaml)
	assert.False(t, info.Structure.HasIndexMd)
}

func TestRunInfoCommandNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Create empty lock file
	baseDir := filepath.Join(tempDir, ".claude")
	memfs := fs.NewMemFS()
	require.NoError(t, memfs.MkdirAll(baseDir, 0o755))

	lockManager := lock.NewManagerWithFS(baseDir, memfs)
	require.NoError(t, lockManager.Load())
	require.NoError(t, lockManager.Save())

	// Run the command for non-existent command
	err := runInfoWithFS("non-existent", false, memfs)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not found")
}

func TestCheckCommandStructure(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(fs.FileSystem, string)
		expectedInfo StructureInfo
		hasMetadata  bool
	}{
		{
			name: "Complete structure",
			setupFunc: func(memfs fs.FileSystem, baseDir string) {
				commandDir := filepath.Join(baseDir, "commands", "test")
				_ = memfs.MkdirAll(commandDir, 0o755)
				_ = memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), []byte("name: test\nversion: 1.0.0\ndescription: Test\nauthor: Test\nrepository: test"), 0o644)
				_ = memfs.WriteFile(filepath.Join(commandDir, "index.md"), []byte("# Test"), 0o644)
				_ = memfs.WriteFile(filepath.Join(baseDir, "commands", "test.md"), []byte("Test"), 0o644)
			},
			expectedInfo: StructureInfo{
				DirectoryExists: true,
				MarkdownExists:  true,
				HasCcmdYaml:     true,
				HasIndexMd:      true,
				IsValid:         true,
				Issues:          []string{},
			},
			hasMetadata: true,
		},
		{
			name: "Missing directory",
			setupFunc: func(memfs fs.FileSystem, baseDir string) {
				_ = memfs.MkdirAll(filepath.Join(baseDir, "commands"), 0o755)
				_ = memfs.WriteFile(filepath.Join(baseDir, "commands", "test.md"), []byte("Test"), 0o644)
			},
			expectedInfo: StructureInfo{
				DirectoryExists: false,
				MarkdownExists:  true,
				HasCcmdYaml:     false,
				HasIndexMd:      false,
				IsValid:         false,
				Issues:          []string{"Command directory is missing"},
			},
			hasMetadata: false,
		},
		{
			name: "Missing markdown file",
			setupFunc: func(memfs fs.FileSystem, baseDir string) {
				commandDir := filepath.Join(baseDir, "commands", "test")
				_ = memfs.MkdirAll(commandDir, 0o755)
				_ = memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), []byte("name: test\nversion: 1.0.0\ndescription: Test\nauthor: Test\nrepository: test"), 0o644)
			},
			expectedInfo: StructureInfo{
				DirectoryExists: true,
				MarkdownExists:  false,
				HasCcmdYaml:     true,
				HasIndexMd:      false,
				IsValid:         false,
				Issues:          []string{"Standalone markdown file is missing"},
			},
			hasMetadata: true,
		},
		{
			name: "Missing ccmd.yaml",
			setupFunc: func(memfs fs.FileSystem, baseDir string) {
				commandDir := filepath.Join(baseDir, "commands", "test")
				_ = memfs.MkdirAll(commandDir, 0o755)
				_ = memfs.WriteFile(filepath.Join(baseDir, "commands", "test.md"), []byte("Test"), 0o644)
			},
			expectedInfo: StructureInfo{
				DirectoryExists: true,
				MarkdownExists:  true,
				HasCcmdYaml:     false,
				HasIndexMd:      false,
				IsValid:         false,
				Issues:          []string{"ccmd.yaml is missing"},
			},
			hasMetadata: false,
		},
		{
			name: "Malformed ccmd.yaml",
			setupFunc: func(memfs fs.FileSystem, baseDir string) {
				commandDir := filepath.Join(baseDir, "commands", "test")
				_ = memfs.MkdirAll(commandDir, 0o755)
				_ = memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), []byte("invalid: yaml: content:"), 0o644)
				_ = memfs.WriteFile(filepath.Join(baseDir, "commands", "test.md"), []byte("Test"), 0o644)
			},
			expectedInfo: StructureInfo{
				DirectoryExists: true,
				MarkdownExists:  true,
				HasCcmdYaml:     true,
				HasIndexMd:      false,
				IsValid:         true,
				Issues:          []string{"ccmd.yaml is malformed"},
			},
			hasMetadata: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			memfs := fs.NewMemFS()

			tt.setupFunc(memfs, tempDir)

			info, metadata := checkCommandStructure("test", tempDir, memfs)

			assert.Equal(t, tt.expectedInfo.DirectoryExists, info.DirectoryExists)
			assert.Equal(t, tt.expectedInfo.MarkdownExists, info.MarkdownExists)
			assert.Equal(t, tt.expectedInfo.HasCcmdYaml, info.HasCcmdYaml)
			assert.Equal(t, tt.expectedInfo.HasIndexMd, info.HasIndexMd)
			assert.Equal(t, tt.expectedInfo.IsValid, info.IsValid)
			assert.Equal(t, tt.expectedInfo.Issues, info.Issues)

			if tt.hasMetadata {
				assert.NotNil(t, metadata)
			} else {
				assert.Nil(t, metadata)
			}
		})
	}
}

func TestPrintStatus(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test successful status
	printStatus("Test item", true)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains checkmark
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, "Test item")

	// Test failed status
	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	printStatus("Failed item", false)

	w2.Close()
	os.Stdout = oldStdout

	buf.Reset()
	_, _ = buf.ReadFrom(r2)
	output = buf.String()

	// Verify output contains X mark
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, "Failed item")
}

func TestDisplayCommandInfo(t *testing.T) {
	info := Output{
		Name:        "test-display",
		Version:     "1.2.3",
		Author:      "Display Author",
		Description: "Display test description",
		Repository:  "https://github.com/test/display",
		Homepage:    "https://display.example.com",
		License:     "Apache-2.0",
		Tags:        []string{"cli", "tool"},
		Entry:       "main.md",
		Source:      "https://github.com/test/display",
		InstalledAt: time.Now().Add(-72 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
		Structure: StructureInfo{
			DirectoryExists: true,
			MarkdownExists:  true,
			HasCcmdYaml:     true,
			HasIndexMd:      true,
			IsValid:         true,
			Issues:          []string{},
		},
	}

	// Create a test filesystem with index.md
	memfs := fs.NewMemFS()
	commandDir := filepath.Join(".claude", "commands", "test-display")
	_ = memfs.MkdirAll(commandDir, 0o755)

	indexContent := `# Test Display Command

This is for testing the display function.`
	_ = memfs.WriteFile(filepath.Join(commandDir, "index.md"), []byte(indexContent), 0o644)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Display the info
	displayCommandInfo(info, ".claude", "test-display", memfs)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify all sections are present
	assert.Contains(t, output, "Command Information")
	assert.Contains(t, output, "test-display")
	assert.Contains(t, output, "1.2.3")
	assert.Contains(t, output, "Display Author")
	assert.Contains(t, output, "Display test description")
	assert.Contains(t, output, "https://github.com/test/display")
	assert.Contains(t, output, "https://display.example.com")
	assert.Contains(t, output, "Apache-2.0")
	assert.Contains(t, output, "cli, tool")
	assert.Contains(t, output, "main.md")

	assert.Contains(t, output, "Installation Details")
	assert.Contains(t, output, "Structure Verification")
	assert.Contains(t, output, "✓")

	// Note: Content preview won't show because we're using a different filesystem
}

func TestInfoWithIncompleteStructure(t *testing.T) {
	// Setup test filesystem
	memfs := fs.NewMemFS()

	// Create base directory structure (project-local)
	baseDir := ".claude"
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands"), 0o755))

	// Create a test command with incomplete structure
	commandName := "incomplete"

	// Only create the markdown file, not the directory
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands", commandName+".md"), []byte("Incomplete"), 0o644))

	// Create lock file
	lockManager := lock.NewManagerWithFS(baseDir, memfs)
	require.NoError(t, lockManager.Load())
	cmd := &models.Command{
		Name:        commandName,
		Version:     "0.1.0",
		Source:      "manual",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, lockManager.AddCommand(cmd))
	require.NoError(t, lockManager.Save())

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runInfoWithFS(commandName, false, memfs)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify no error (command exists)
	assert.NoError(t, err)

	// Verify output shows structure issues
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, "Command directory is missing")
	assert.Contains(t, output, "Issues found:")
}

func TestRunInfoWithJSONError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Create empty lock file
	baseDir := filepath.Join(tempDir, ".claude")
	memfs := fs.NewMemFS()
	require.NoError(t, memfs.MkdirAll(baseDir, 0o755))

	lockManager := lock.NewManagerWithFS(baseDir, memfs)
	require.NoError(t, lockManager.Load())
	require.NoError(t, lockManager.Save())

	// Run the command for non-existent command with JSON output
	err := runInfoWithFS("non-existent", true, memfs)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command 'non-existent' is not installed")
}

func TestOutputMarshaling(t *testing.T) {
	// Test that Output properly marshals to JSON
	info := Output{
		Name:        "marshal-test",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Test marshaling",
		Tags:        []string{"test"},
		Source:      "test",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Structure: StructureInfo{
			DirectoryExists: true,
			MarkdownExists:  true,
			HasCcmdYaml:     true,
			HasIndexMd:      false,
			IsValid:         true,
		},
	}

	data, err := json.Marshal(info)
	assert.NoError(t, err)

	var unmarshaled Output
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, info.Name, unmarshaled.Name)
	assert.Equal(t, info.Version, unmarshaled.Version)
	assert.Equal(t, info.Structure.IsValid, unmarshaled.Structure.IsValid)
}

func TestCommandIntegration(t *testing.T) {
	// Integration test using the actual command
	cmd := NewCommand()

	// Test with missing argument
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")

	// Test with too many arguments
	cmd.SetArgs([]string{"cmd1", "cmd2"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 2")
}

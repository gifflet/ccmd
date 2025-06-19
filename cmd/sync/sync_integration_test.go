package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/project"
)

// mockFileSystem implements fs.FileSystem for testing
type mockFileSystem struct {
	files       map[string][]byte
	directories map[string]bool
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files:       make(map[string][]byte),
		directories: make(map[string]bool),
	}
}

func (m *mockFileSystem) ReadFile(name string) ([]byte, error) {
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(name)
	if dir != "." && dir != "/" {
		if err := m.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	m.files[name] = data
	return nil
}

func (m *mockFileSystem) Remove(name string) error {
	if _, ok := m.files[name]; ok {
		delete(m.files, name)
		return nil
	}
	if _, ok := m.directories[name]; ok {
		delete(m.directories, name)
		return nil
	}
	return os.ErrNotExist
}

func (m *mockFileSystem) RemoveAll(path string) error {
	// Remove all files and directories under path
	toRemove := []string{}
	for p := range m.files {
		if strings.HasPrefix(p, path) {
			toRemove = append(toRemove, p)
		}
	}
	for _, p := range toRemove {
		delete(m.files, p)
	}

	toRemove = []string{}
	for p := range m.directories {
		if strings.HasPrefix(p, path) {
			toRemove = append(toRemove, p)
		}
	}
	for _, p := range toRemove {
		delete(m.directories, p)
	}
	return nil
}

func (m *mockFileSystem) Rename(oldpath, newpath string) error {
	if data, ok := m.files[oldpath]; ok {
		m.files[newpath] = data
		delete(m.files, oldpath)
		return nil
	}
	return os.ErrNotExist
}

func (m *mockFileSystem) Stat(name string) (fs.FileInfo, error) {
	if _, ok := m.files[name]; ok {
		return &mockFileInfo{name: filepath.Base(name), size: 0, mode: 0644, isDir: false}, nil
	}
	if _, ok := m.directories[name]; ok {
		return &mockFileInfo{name: filepath.Base(name), size: 0, mode: 0755, isDir: true}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.directories[path] = true
	// Create parent directories
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" && parent != path {
		return m.MkdirAll(parent, perm)
	}
	return nil
}

func (m *mockFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	if _, ok := m.directories[name]; !ok {
		return nil, os.ErrNotExist
	}

	entries := []fs.DirEntry{}
	seen := make(map[string]bool)

	// Find all direct children
	for path := range m.files {
		if strings.HasPrefix(path, name+"/") {
			rel, _ := filepath.Rel(name, path)
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 0 && !seen[parts[0]] {
				seen[parts[0]] = true
				isDir := false
				if len(parts) > 1 {
					isDir = true
				}
				entries = append(entries, &mockDirEntry{
					name:  parts[0],
					isDir: isDir,
				})
			}
		}
	}

	for path := range m.directories {
		if strings.HasPrefix(path, name+"/") && path != name {
			rel, _ := filepath.Rel(name, path)
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 0 && !seen[parts[0]] {
				seen[parts[0]] = true
				entries = append(entries, &mockDirEntry{
					name:  parts[0],
					isDir: true,
				})
			}
		}
	}

	return entries, nil
}

func (m *mockFileSystem) Exists(path string) (bool, error) {
	_, fileExists := m.files[path]
	_, dirExists := m.directories[path]
	return fileExists || dirExists, nil
}

type mockFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string      { return m.name }
func (m *mockDirEntry) IsDir() bool       { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) {
	return &mockFileInfo{name: m.name, isDir: m.isDir}, nil
}

// Integration tests using mock filesystem
func TestSyncIntegration(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func(fs *mockFileSystem)
		configCommands    []project.ConfigCommand
		installedCommands []*models.Command
		dryRun            bool
		force             bool
		wantInstalled     []string
		wantRemoved       []string
		wantError         bool
	}{
		{
			name: "sync with empty config removes all commands",
			setupFunc: func(fs *mockFileSystem) {
				// Setup lock file with installed commands
				fs.MkdirAll(".claude", 0755)
			},
			configCommands: []project.ConfigCommand{},
			installedCommands: []*models.Command{
				{Name: "tool1", Source: "https://github.com/owner/tool1.git", Version: "v1.0.0", InstalledAt: time.Now(), UpdatedAt: time.Now()},
				{Name: "tool2", Source: "https://github.com/owner/tool2.git", Version: "v2.0.0", InstalledAt: time.Now(), UpdatedAt: time.Now()},
			},
			force:         true,
			wantInstalled: []string{},
			wantRemoved:   []string{"tool1", "tool2"},
		},
		{
			name: "sync installs missing commands",
			setupFunc: func(fs *mockFileSystem) {
				fs.MkdirAll(".claude", 0755)
			},
			configCommands: []project.ConfigCommand{
				{Repo: "owner/tool1", Version: "v1.0.0"},
				{Repo: "owner/tool2", Version: "v2.0.0"},
			},
			installedCommands: []*models.Command{},
			wantInstalled:     []string{"tool1", "tool2"},
			wantRemoved:       []string{},
		},
		{
			name: "sync with dry run makes no changes",
			setupFunc: func(fs *mockFileSystem) {
				fs.MkdirAll(".claude", 0755)
			},
			configCommands: []project.ConfigCommand{
				{Repo: "owner/tool1", Version: "v1.0.0"},
			},
			installedCommands: []*models.Command{
				{Name: "tool2", Source: "https://github.com/owner/tool2.git", Version: "v2.0.0", InstalledAt: time.Now(), UpdatedAt: time.Now()},
			},
			dryRun:        true,
			wantInstalled: []string{},
			wantRemoved:   []string{},
		},
		{
			name: "sync with both install and remove",
			setupFunc: func(fs *mockFileSystem) {
				fs.MkdirAll(".claude", 0755)
			},
			configCommands: []project.ConfigCommand{
				{Repo: "owner/tool1", Version: "v1.0.0"},
				{Repo: "owner/tool3", Version: "v3.0.0"},
			},
			installedCommands: []*models.Command{
				{Name: "tool1", Source: "https://github.com/owner/tool1.git", Version: "v1.0.0", InstalledAt: time.Now(), UpdatedAt: time.Now()},
				{Name: "tool2", Source: "https://github.com/owner/tool2.git", Version: "v2.0.0", InstalledAt: time.Now(), UpdatedAt: time.Now()},
			},
			force:         true,
			wantInstalled: []string{"tool3"},
			wantRemoved:   []string{"tool2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip integration tests that would require actual command operations
			t.Skip("Integration tests require command operation mocking")
		})
	}
}

// TestPromptConfirmationResponses tests confirmation input parsing
func TestPromptConfirmationResponses(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"Yes\n", true},
		{" y \n", true},
		{" yes \n", true},
		{"n\n", false},
		{"N\n", false},
		{"no\n", false},
		{"NO\n", false},
		{"\n", false},
		{"maybe\n", false},
		{"yep\n", false},
		{"yeah\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isConfirmation(strings.TrimSpace(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountErrorsFunction(t *testing.T) {
	errors := []error{
		fmt.Errorf("failed to install tool1: connection error"),
		fmt.Errorf("failed to remove tool2: permission denied"),
		fmt.Errorf("failed to install tool3: timeout"),
		fmt.Errorf("some other error"),
	}

	assert.Equal(t, 2, countErrors(errors, "install"))
	assert.Equal(t, 1, countErrors(errors, "remove"))
	assert.Equal(t, 0, countErrors(errors, "update"))
	assert.Equal(t, 1, countErrors(errors, "other"))
}

func TestUpdateLockFileIntegration(t *testing.T) {
	// This test requires mocking the project manager and commands.List
	t.Skip("Requires comprehensive mocking of project manager")
}

// Test command structure
func TestSyncCommandStructure(t *testing.T) {
	cmd := NewCommand()

	assert.Equal(t, "sync", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
	assert.Equal(t, "false", dryRunFlag.DefValue)

	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)
	assert.Equal(t, "f", forceFlag.Shorthand)
}

// Test sync result structure
func TestSyncResultOperations(t *testing.T) {
	result := Result{
		ToInstall: []string{"tool1", "tool2"},
		ToRemove:  []string{"tool3"},
		Errors:    []error{},
	}

	assert.Len(t, result.ToInstall, 2)
	assert.Contains(t, result.ToInstall, "tool1")
	assert.Contains(t, result.ToInstall, "tool2")

	assert.Len(t, result.ToRemove, 1)
	assert.Contains(t, result.ToRemove, "tool3")

	assert.Empty(t, result.Errors)

	// Add error
	result.Errors = append(result.Errors, fmt.Errorf("test error"))
	assert.Len(t, result.Errors, 1)
}

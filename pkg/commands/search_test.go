package commands

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(fs fs.FileSystem, baseDir string)
		opts        SearchOptions
		wantResults []SearchResult
		wantErr     bool
	}{
		{
			name: "search by keyword in name",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"test-command": {
							Name:        "test-command",
							Version:     "1.0.0",
							Source:      "github.com/example/test",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"description": "A test command",
								"author":      "Test Author",
								"tags":        "test, cli",
							},
						},
						"another-tool": {
							Name:        "another-tool",
							Version:     "2.0.0",
							Source:      "github.com/example/another",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"description": "Another tool for testing",
								"author":      "Another Author",
								"tags":        "tool, utility",
							},
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				Keyword: "test",
			},
			wantResults: []SearchResult{
				{
					Name:        "test-command",
					Version:     "1.0.0",
					Description: "A test command",
					Author:      "Test Author",
					Tags:        []string{"test", "cli"},
					Source:      "github.com/example/test",
				},
				{
					Name:        "another-tool",
					Version:     "2.0.0",
					Description: "Another tool for testing",
					Author:      "Another Author",
					Tags:        []string{"tool", "utility"},
					Source:      "github.com/example/another",
				},
			},
		},
		{
			name: "search by author",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							Source:      "source1",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"author": "John Doe",
							},
						},
						"cmd2": {
							Name:        "cmd2",
							Version:     "1.0.0",
							Source:      "source2",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"author": "Jane Doe",
							},
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				Author: "John Doe",
			},
			wantResults: []SearchResult{
				{
					Name:    "cmd1",
					Version: "1.0.0",
					Author:  "John Doe",
					Source:  "source1",
				},
			},
		},
		{
			name: "search by tags",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							Source:      "source1",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"tags": "cli, tool, dev",
							},
						},
						"cmd2": {
							Name:        "cmd2",
							Version:     "1.0.0",
							Source:      "source2",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"tags": "web, api",
							},
						},
						"cmd3": {
							Name:        "cmd3",
							Version:     "1.0.0",
							Source:      "source3",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"tags": "cli, api",
							},
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				Tags: []string{"cli"},
			},
			wantResults: []SearchResult{
				{
					Name:    "cmd1",
					Version: "1.0.0",
					Tags:    []string{"cli", "tool", "dev"},
					Source:  "source1",
				},
				{
					Name:    "cmd3",
					Version: "1.0.0",
					Tags:    []string{"cli", "api"},
					Source:  "source3",
				},
			},
		},
		{
			name: "search with multiple filters",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"test-cli": {
							Name:        "test-cli",
							Version:     "1.0.0",
							Source:      "source1",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"description": "Test CLI tool",
								"author":      "Test Author",
								"tags":        "cli, test",
							},
						},
						"another-cli": {
							Name:        "another-cli",
							Version:     "1.0.0",
							Source:      "source2",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"description": "Another CLI tool",
								"author":      "Test Author",
								"tags":        "cli, utility",
							},
						},
						"test-web": {
							Name:        "test-web",
							Version:     "1.0.0",
							Source:      "source3",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Metadata: map[string]string{
								"description": "Test web tool",
								"author":      "Test Author",
								"tags":        "web, test",
							},
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				Keyword: "test",
				Tags:    []string{"cli"},
			},
			wantResults: []SearchResult{
				{
					Name:        "test-cli",
					Version:     "1.0.0",
					Description: "Test CLI tool",
					Author:      "Test Author",
					Tags:        []string{"cli", "test"},
					Source:      "source1",
				},
			},
		},
		{
			name: "show all commands",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							Source:      "source1",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
						"cmd2": {
							Name:        "cmd2",
							Version:     "2.0.0",
							Source:      "source2",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				ShowAll: true,
			},
			wantResults: []SearchResult{
				{
					Name:    "cmd1",
					Version: "1.0.0",
					Source:  "source1",
				},
				{
					Name:    "cmd2",
					Version: "2.0.0",
					Source:  "source2",
				},
			},
		},
		{
			name: "no results found",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							Source:      "source1",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
					},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				Keyword: "nonexistent",
			},
			wantResults: []SearchResult{},
		},
		{
			name: "empty lock file",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version:  "1.0",
					Commands: map[string]*models.Command{},
				}
				writeTestLockFile(t, mockFS, baseDir, lockFile)
			},
			opts: SearchOptions{
				ShowAll: true,
			},
			wantResults: []SearchResult{},
		},
		{
			name: "no lock file",
			setupFunc: func(mockFS fs.FileSystem, baseDir string) {
				// Don't create any lock file
			},
			opts: SearchOptions{
				ShowAll: true,
			},
			wantResults: []SearchResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := fs.NewMemFS()
			baseDir := "/test"

			if tt.setupFunc != nil {
				tt.setupFunc(mockFS, baseDir)
			}

			tt.opts.BaseDir = baseDir
			tt.opts.FileSystem = mockFS

			results, err := Search(tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.wantResults), len(results))

			// Compare results (order may vary)
			for _, want := range tt.wantResults {
				found := false
				for _, got := range results {
					if got.Name == want.Name {
						found = true
						assert.Equal(t, want.Version, got.Version)
						assert.Equal(t, want.Description, got.Description)
						assert.Equal(t, want.Author, got.Author)
						assert.Equal(t, want.Source, got.Source)
						assert.ElementsMatch(t, want.Tags, got.Tags)
						break
					}
				}
				assert.True(t, found, "Expected result with name %s not found", want.Name)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single tag",
			input:    "cli",
			expected: []string{"cli"},
		},
		{
			name:     "multiple tags",
			input:    "cli, tool, dev",
			expected: []string{"cli", "tool", "dev"},
		},
		{
			name:     "tags with extra spaces",
			input:    "  cli  ,  tool  ,  dev  ",
			expected: []string{"cli", "tool", "dev"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "tags with empty values",
			input:    "cli, , tool",
			expected: []string{"cli", "tool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *models.Command
		opts     SearchOptions
		expected bool
	}{
		{
			name: "match by name substring",
			cmd: &models.Command{
				Name: "test-command",
			},
			opts: SearchOptions{
				Keyword: "test",
			},
			expected: true,
		},
		{
			name: "match by description",
			cmd: &models.Command{
				Name: "cmd",
				Metadata: map[string]string{
					"description": "This is a test tool",
				},
			},
			opts: SearchOptions{
				Keyword: "test",
			},
			expected: true,
		},
		{
			name: "case insensitive match",
			cmd: &models.Command{
				Name: "TEST-COMMAND",
			},
			opts: SearchOptions{
				Keyword: "test",
			},
			expected: true,
		},
		{
			name: "no match when no criteria",
			cmd: &models.Command{
				Name: "command",
			},
			opts:     SearchOptions{},
			expected: false,
		},
		{
			name: "match with show all",
			cmd: &models.Command{
				Name: "command",
			},
			opts: SearchOptions{
				ShowAll: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matches(tt.cmd, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func writeTestLockFile(t *testing.T, mockFS fs.FileSystem, baseDir string, lockFile *models.LockFile) {
	data, err := lockFile.ToJSON()
	require.NoError(t, err)

	lockPath := filepath.Join(baseDir, "commands.lock")
	err = mockFS.WriteFile(lockPath, data, 0o644)
	require.NoError(t, err)
}

package commands

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
)

func TestList(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, memFS *fs.MemFS, baseDir string)
		expectedCount int
		checkFunc     func(t *testing.T, details []*CommandDetail)
		wantErr       bool
		errContains   string
	}{
		{
			name: "no lock file returns empty list",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// No setup - lock file doesn't exist
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "empty lock file",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "single command with valid structure",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "test-cmd",
					Version:     "1.0.0",
					Source:      "github.com/example/test-cmd",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Create valid structure
				commandDir := filepath.Join(baseDir, "commands", "test-cmd")
				err = memFS.MkdirAll(commandDir, 0o755)
				require.NoError(t, err)

				markdownFile := filepath.Join(baseDir, "commands", "test-cmd.md")
				err = memFS.WriteFile(markdownFile, []byte("# Test Command"), 0o644)
				require.NoError(t, err)
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, details []*CommandDetail) {
				assert.Equal(t, "test-cmd", details[0].Name)
				assert.True(t, details[0].HasDirectory)
				assert.True(t, details[0].HasMarkdownFile)
				assert.True(t, details[0].StructureValid)
				assert.Empty(t, details[0].StructureMessage)
			},
		},
		{
			name: "command missing directory",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-dir",
					Version:     "1.0.0",
					Source:      "github.com/example/no-dir",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Only create markdown file
				markdownFile := filepath.Join(baseDir, "commands", "no-dir.md")
				err = memFS.WriteFile(markdownFile, []byte("# No Dir Command"), 0o644)
				require.NoError(t, err)
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, details []*CommandDetail) {
				assert.Equal(t, "no-dir", details[0].Name)
				assert.False(t, details[0].HasDirectory)
				assert.True(t, details[0].HasMarkdownFile)
				assert.False(t, details[0].StructureValid)
				assert.Contains(t, details[0].StructureMessage, "missing directory")
			},
		},
		{
			name: "command missing markdown",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-md",
					Version:     "1.0.0",
					Source:      "github.com/example/no-md",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Only create directory
				commandDir := filepath.Join(baseDir, "commands", "no-md")
				err = memFS.MkdirAll(commandDir, 0o755)
				require.NoError(t, err)
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, details []*CommandDetail) {
				assert.Equal(t, "no-md", details[0].Name)
				assert.True(t, details[0].HasDirectory)
				assert.False(t, details[0].HasMarkdownFile)
				assert.False(t, details[0].StructureValid)
				assert.Contains(t, details[0].StructureMessage, "missing .md file")
			},
		},
		{
			name: "command missing both files",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-files",
					Version:     "1.0.0",
					Source:      "github.com/example/no-files",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Don't create any files
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, details []*CommandDetail) {
				assert.Equal(t, "no-files", details[0].Name)
				assert.False(t, details[0].HasDirectory)
				assert.False(t, details[0].HasMarkdownFile)
				assert.False(t, details[0].StructureValid)
				assert.Contains(t, details[0].StructureMessage, "missing directory")
				assert.Contains(t, details[0].StructureMessage, "missing .md file")
			},
		},
		{
			name: "multiple commands with mixed structure",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				// Add multiple commands
				commands := []struct {
					cmd       *models.Command
					createDir bool
					createMD  bool
				}{
					{
						cmd: &models.Command{
							Name:        "valid-cmd",
							Version:     "1.0.0",
							Source:      "github.com/example/valid-cmd",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
						createDir: true,
						createMD:  true,
					},
					{
						cmd: &models.Command{
							Name:        "broken-cmd",
							Version:     "2.0.0",
							Source:      "github.com/example/broken-cmd",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
						createDir: true,
						createMD:  false,
					},
					{
						cmd: &models.Command{
							Name:        "another-cmd",
							Version:     "3.0.0",
							Source:      "github.com/example/another-cmd",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
						},
						createDir: false,
						createMD:  false,
					},
				}

				for _, c := range commands {
					err = lockManager.AddCommand(c.cmd)
					require.NoError(t, err)

					if c.createDir {
						commandDir := filepath.Join(baseDir, "commands", c.cmd.Name)
						err = memFS.MkdirAll(commandDir, 0o755)
						require.NoError(t, err)
					}

					if c.createMD {
						markdownFile := filepath.Join(baseDir, "commands", c.cmd.Name+".md")
						err = memFS.WriteFile(markdownFile, []byte("# "+c.cmd.Name), 0o644)
						require.NoError(t, err)
					}
				}

				err = lockManager.Save()
				require.NoError(t, err)
			},
			expectedCount: 3,
			checkFunc: func(t *testing.T, details []*CommandDetail) {
				// Find each command and check its structure
				validCount := 0
				brokenCount := 0
				for _, d := range details {
					if d.StructureValid {
						validCount++
					} else {
						brokenCount++
					}
				}
				assert.Equal(t, 1, validCount)
				assert.Equal(t, 2, brokenCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := fs.NewMemFS()
			baseDir := "/test/.claude"

			// Create base directory
			err := memFS.MkdirAll(baseDir, 0o755)
			require.NoError(t, err)

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc(t, memFS, baseDir)
			}

			// Execute List
			opts := ListOptions{
				BaseDir:    baseDir,
				FileSystem: memFS,
			}
			details, err := List(opts)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, details, tt.expectedCount)

			// Run additional checks
			if tt.checkFunc != nil && len(details) > 0 {
				tt.checkFunc(t, details)
			}
		})
	}
}

func TestVerifyCommandStructure(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		setupFunc   func(t *testing.T, memFS *fs.MemFS, baseDir string)
		wantValid   bool
		wantMessage string
		wantErr     bool
		errContains string
	}{
		{
			name:        "command not in lock file",
			commandName: "nonexistent",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create empty lock file
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)
			},
			wantValid:   false,
			wantMessage: "command not found in lock file",
			wantErr:     false,
		},
		{
			name:        "valid structure",
			commandName: "valid-cmd",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file with command
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "valid-cmd",
					Version:     "1.0.0",
					Source:      "github.com/example/valid-cmd",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Create both directory and markdown file
				commandDir := filepath.Join(baseDir, "commands", "valid-cmd")
				err = memFS.MkdirAll(commandDir, 0o755)
				require.NoError(t, err)

				markdownFile := filepath.Join(baseDir, "commands", "valid-cmd.md")
				err = memFS.WriteFile(markdownFile, []byte("# Valid Command"), 0o644)
				require.NoError(t, err)
			},
			wantValid:   true,
			wantMessage: "",
			wantErr:     false,
		},
		{
			name:        "missing directory",
			commandName: "no-dir",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file with command
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-dir",
					Version:     "1.0.0",
					Source:      "github.com/example/no-dir",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Only create markdown file
				markdownFile := filepath.Join(baseDir, "commands", "no-dir.md")
				err = memFS.WriteFile(markdownFile, []byte("# No Dir"), 0o644)
				require.NoError(t, err)
			},
			wantValid:   false,
			wantMessage: "broken structure: [missing directory]",
			wantErr:     false,
		},
		{
			name:        "missing markdown",
			commandName: "no-md",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file with command
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-md",
					Version:     "1.0.0",
					Source:      "github.com/example/no-md",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Only create directory
				commandDir := filepath.Join(baseDir, "commands", "no-md")
				err = memFS.MkdirAll(commandDir, 0o755)
				require.NoError(t, err)
			},
			wantValid:   false,
			wantMessage: "broken structure: [missing .md file]",
			wantErr:     false,
		},
		{
			name:        "missing both",
			commandName: "no-files",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// Create lock file with command
				lockManager := lock.NewManagerWithFS(baseDir, memFS)
				err := lockManager.Load()
				require.NoError(t, err)

				cmd := &models.Command{
					Name:        "no-files",
					Version:     "1.0.0",
					Source:      "github.com/example/no-files",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				err = lockManager.AddCommand(cmd)
				require.NoError(t, err)
				err = lockManager.Save()
				require.NoError(t, err)

				// Don't create any files
			},
			wantValid:   false,
			wantMessage: "broken structure: [missing directory missing .md file]",
			wantErr:     false,
		},
		{
			name:        "no lock file",
			commandName: "any-cmd",
			setupFunc: func(t *testing.T, memFS *fs.MemFS, baseDir string) {
				// No setup - lock file doesn't exist
			},
			wantValid:   false,
			wantMessage: "command not found in lock file",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := fs.NewMemFS()
			baseDir := "/test/.claude"

			// Create base directory
			err := memFS.MkdirAll(baseDir, 0o755)
			require.NoError(t, err)

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc(t, memFS, baseDir)
			}

			// Verify structure
			valid, message, err := VerifyCommandStructure(tt.commandName, baseDir, memFS)

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValid, valid)
				assert.Equal(t, tt.wantMessage, message)
			}
		})
	}
}

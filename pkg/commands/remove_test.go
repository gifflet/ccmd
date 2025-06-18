package commands

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestRemove(t *testing.T) {
	tests := []struct {
		name          string
		opts          RemoveOptions
		setupFunc     func(fs.FileSystem, string)
		expectedError string
	}{
		{
			name: "successful removal",
			opts: RemoveOptions{
				Name:    "test-cmd",
				BaseDir: "/test/.claude",
			},
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Create lock file with command
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"test-cmd": {
							Name:        "test-cmd",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
							Metadata: map[string]string{
								"author":      "Test Author",
								"description": "Test command",
							},
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)

				// Create command directory
				_ = fs.MkdirAll(filepath.Join(baseDir, "commands", "test-cmd"), 0755)
				// Create command markdown file
				_ = fs.WriteFile(filepath.Join(baseDir, "commands", "test-cmd.md"), []byte("# Test Command"), 0644)
			},
		},
		{
			name: "successful removal with only markdown file",
			opts: RemoveOptions{
				Name:    "md-only-cmd",
				BaseDir: "/test/.claude",
			},
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Create lock file with command
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"md-only-cmd": {
							Name:        "md-only-cmd",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)

				// Create only command markdown file (no directory)
				_ = fs.MkdirAll(filepath.Join(baseDir, "commands"), 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands", "md-only-cmd.md"), []byte("# MD Only Command"), 0644)
			},
		},
		{
			name: "successful removal with only directory",
			opts: RemoveOptions{
				Name:    "dir-only-cmd",
				BaseDir: "/test/.claude",
			},
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Create lock file with command
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"dir-only-cmd": {
							Name:        "dir-only-cmd",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)

				// Create only command directory (no markdown file)
				_ = fs.MkdirAll(filepath.Join(baseDir, "commands", "dir-only-cmd"), 0755)
			},
		},
		{
			name: "command not found",
			opts: RemoveOptions{
				Name:    "non-existent",
				BaseDir: "/test/.claude",
			},
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Create empty lock file
				lockFile := &models.LockFile{
					Version:  "1.0",
					Commands: map[string]*models.Command{},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedError: "command 'non-existent' not found",
		},
		{
			name: "missing command name",
			opts: RemoveOptions{
				Name:    "",
				BaseDir: "/test/.claude",
			},
			expectedError: "command name is required",
		},
		{
			name: "lock file not found",
			opts: RemoveOptions{
				Name:    "test-cmd",
				BaseDir: "/test/.claude",
			},
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Don't create anything
			},
			expectedError: "command 'test-cmd' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			memFS := fs.NewMemFS()
			tt.opts.FileSystem = memFS

			if tt.setupFunc != nil {
				tt.setupFunc(memFS, tt.opts.BaseDir)
			}

			// Execute
			err := Remove(tt.opts)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				// Verify command was removed from lock file
				lockData, err := memFS.ReadFile(filepath.Join(tt.opts.BaseDir, "commands.lock"))
				require.NoError(t, err)

				lockFile := &models.LockFile{}
				err = json.Unmarshal(lockData, lockFile)
				require.NoError(t, err)

				_, exists := lockFile.Commands[tt.opts.Name]
				assert.False(t, exists)

				// Verify command directory was removed
				_, err = memFS.Stat(filepath.Join(tt.opts.BaseDir, "commands", tt.opts.Name))
				assert.Error(t, err) // Should return error because directory doesn't exist

				// Verify command markdown file was removed
				_, err = memFS.Stat(filepath.Join(tt.opts.BaseDir, "commands", tt.opts.Name+".md"))
				assert.Error(t, err) // Should return error because file doesn't exist
			}
		})
	}
}

func TestListCommands(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(fs.FileSystem, string)
		expectedCmds  []string
		expectedError string
	}{
		{
			name: "list multiple commands",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
						},
						"cmd2": {
							Name:        "cmd2",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
						},
						"cmd3": {
							Name:        "cmd3",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedCmds: []string{"cmd1", "cmd2", "cmd3"},
		},
		{
			name: "empty command list",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version:  "1.0",
					Commands: map[string]*models.Command{},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedCmds: []string{},
		},
		{
			name: "no lock file",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Don't create lock file
			},
			expectedCmds: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			memFS := fs.NewMemFS()
			baseDir := "/test/.claude"

			if tt.setupFunc != nil {
				tt.setupFunc(memFS, baseDir)
			}

			// Execute
			commands, err := ListCommands(baseDir, memFS)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedCmds, commands)
			}
		})
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name           string
		cmdName        string
		setupFunc      func(fs.FileSystem, string)
		expectedExists bool
		expectedError  string
	}{
		{
			name:    "command exists",
			cmdName: "existing-cmd",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"existing-cmd": {
							Name:        "existing-cmd",
							Version:     "1.0.0",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedExists: true,
		},
		{
			name:    "command does not exist",
			cmdName: "non-existent",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version:  "1.0",
					Commands: map[string]*models.Command{},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedExists: false,
		},
		{
			name:    "no lock file",
			cmdName: "any-cmd",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				// Don't create lock file
			},
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			memFS := fs.NewMemFS()
			baseDir := "/test/.claude"

			if tt.setupFunc != nil {
				tt.setupFunc(memFS, baseDir)
			}

			// Execute
			exists, err := CommandExists(tt.cmdName, baseDir, memFS)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedExists, exists)
			}
		})
	}
}

func TestGetCommandInfo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		cmdName       string
		setupFunc     func(fs.FileSystem, string)
		expectedInfo  *models.Command
		expectedError string
	}{
		{
			name:    "get existing command info",
			cmdName: "test-cmd",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version: "1.0",
					Commands: map[string]*models.Command{
						"test-cmd": {
							Name:        "test-cmd",
							Version:     "1.0.0",
							InstalledAt: now,
							UpdatedAt:   now,
							Source:      "https://github.com/user/repo",
							Metadata: map[string]string{
								"author":      "Test Author",
								"description": "Test command",
							},
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedInfo: &models.Command{
				Name:        "test-cmd",
				Version:     "1.0.0",
				InstalledAt: now,
				UpdatedAt:   now,
				Source:      "https://github.com/user/repo",
				Metadata: map[string]string{
					"author":      "Test Author",
					"description": "Test command",
				},
			},
		},
		{
			name:    "command not found",
			cmdName: "non-existent",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &models.LockFile{
					Version:  "1.0",
					Commands: map[string]*models.Command{},
				}
				data, _ := json.Marshal(lockFile)
				_ = fs.MkdirAll(baseDir, 0755)
				_ = fs.WriteFile(filepath.Join(baseDir, "commands.lock"), data, 0644)
			},
			expectedError: "command \"non-existent\" not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			memFS := fs.NewMemFS()
			baseDir := "/test/.claude"

			if tt.setupFunc != nil {
				tt.setupFunc(memFS, baseDir)
			}

			// Execute
			info, err := GetCommandInfo(tt.cmdName, baseDir, memFS)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedInfo.Name, info.Name)
				assert.Equal(t, tt.expectedInfo.Version, info.Version)
				assert.Equal(t, tt.expectedInfo.Source, info.Source)
				assert.Equal(t, tt.expectedInfo.Metadata["author"], info.Metadata["author"])
				assert.Equal(t, tt.expectedInfo.Metadata["description"], info.Metadata["description"])
			}
		})
	}
}

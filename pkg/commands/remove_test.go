/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package commands

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/project"
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
				lockFile := &project.LockFile{
					Version:         "1.0",
					LockfileVersion: 1,
					Commands: map[string]*project.CommandLockInfo{
						"test-cmd": {
							Name:        "test-cmd",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
							Resolved:    "https://github.com/user/repo@1.0.0",
							Metadata: map[string]string{
								"author":      "Test Author",
								"description": "Test command",
							},
						},
					},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)

				// Create command directory
				_ = fs.MkdirAll(filepath.Join(baseDir, ".claude", "commands", "test-cmd"), 0o755)
				// Create command markdown file
				_ = fs.WriteFile(filepath.Join(baseDir, ".claude", "commands", "test-cmd.md"), []byte("# Test Command"), 0o644)
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
				lockFile := &project.LockFile{
					Version:         "1.0",
					LockfileVersion: 1,
					Commands: map[string]*project.CommandLockInfo{
						"md-only-cmd": {
							Name:        "md-only-cmd",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
							Resolved:    "https://github.com/user/repo@1.0.0",
						},
					},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)

				// Create only command markdown file (no directory)
				_ = fs.MkdirAll(filepath.Join(baseDir, ".claude", "commands"), 0o755)
				_ = fs.WriteFile(filepath.Join(baseDir, ".claude", "commands", "md-only-cmd.md"), []byte("# MD Only Command"), 0o644)
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
				lockFile := &project.LockFile{
					Version:         "1.0",
					LockfileVersion: 1,
					Commands: map[string]*project.CommandLockInfo{
						"dir-only-cmd": {
							Name:        "dir-only-cmd",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "https://github.com/user/repo",
							Resolved:    "https://github.com/user/repo@1.0.0",
						},
					},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)

				// Create only command directory (no markdown file)
				_ = fs.MkdirAll(filepath.Join(baseDir, ".claude", "commands", "dir-only-cmd"), 0o755)
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
				lockFile := &project.LockFile{
					Version:  "1.0",
					Commands: map[string]*project.CommandLockInfo{},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)
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
				lockData, err := memFS.ReadFile("ccmd-lock.yaml")
				require.NoError(t, err)

				lockFile := &project.LockFile{}
				err = yaml.Unmarshal(lockData, lockFile)
				require.NoError(t, err)

				_, exists := lockFile.Commands[tt.opts.Name]
				assert.False(t, exists)

				// Verify command directory was removed
				_, err = memFS.Stat(filepath.Join(tt.opts.BaseDir, ".claude", "commands", tt.opts.Name))
				assert.Error(t, err) // Should return error because directory doesn't exist

				// Verify command markdown file was removed
				_, err = memFS.Stat(filepath.Join(tt.opts.BaseDir, ".claude", "commands", tt.opts.Name+".md"))
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
				lockFile := &project.LockFile{
					Version:         "1.0",
					LockfileVersion: 1,
					Commands: map[string]*project.CommandLockInfo{
						"cmd1": {
							Name:        "cmd1",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
							Resolved:    "test@1.0.0",
						},
						"cmd2": {
							Name:        "cmd2",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
							Resolved:    "test@1.0.0",
						},
						"cmd3": {
							Name:        "cmd3",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: time.Now(),
							UpdatedAt:   time.Now(),
							Source:      "test",
							Resolved:    "test@1.0.0",
						},
					},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)
			},
			expectedCmds: []string{"cmd1", "cmd2", "cmd3"},
		},
		{
			name: "empty command list",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &project.LockFile{
					Version:  "1.0",
					Commands: map[string]*project.CommandLockInfo{},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)
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

func TestGetCommandInfo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		cmdName       string
		setupFunc     func(fs.FileSystem, string)
		expectedInfo  *project.CommandLockInfo
		expectedError string
	}{
		{
			name:    "get existing command info",
			cmdName: "test-cmd",
			setupFunc: func(fs fs.FileSystem, baseDir string) {
				lockFile := &project.LockFile{
					Version:         "1.0",
					LockfileVersion: 1,
					Commands: map[string]*project.CommandLockInfo{
						"test-cmd": {
							Name:        "test-cmd",
							Version:     "1.0.0",
							Commit:      "1234567890abcdef1234567890abcdef12345678",
							InstalledAt: now,
							UpdatedAt:   now,
							Source:      "https://github.com/user/repo",
							Resolved:    "https://github.com/user/repo@1.0.0",
							Metadata: map[string]string{
								"author":      "Test Author",
								"description": "Test command",
							},
						},
					},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)
			},
			expectedInfo: &project.CommandLockInfo{
				Name:        "test-cmd",
				Version:     "1.0.0",
				Commit:      "1234567890abcdef1234567890abcdef12345678",
				InstalledAt: now,
				UpdatedAt:   now,
				Source:      "https://github.com/user/repo",
				Resolved:    "https://github.com/user/repo@1.0.0",
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
				lockFile := &project.LockFile{
					Version:  "1.0",
					Commands: map[string]*project.CommandLockInfo{},
				}
				data, _ := yaml.Marshal(lockFile)
				_ = fs.WriteFile("ccmd-lock.yaml", data, 0o644)
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
				require.NotNil(t, info, "Command info should not be nil")
				assert.Equal(t, tt.expectedInfo.Name, info.Name)
				assert.Equal(t, tt.expectedInfo.Version, info.Version)
				assert.Equal(t, tt.expectedInfo.Source, info.Source)
				if info.Metadata != nil {
					assert.Equal(t, tt.expectedInfo.Metadata["author"], info.Metadata["author"])
					assert.Equal(t, tt.expectedInfo.Metadata["description"], info.Metadata["description"])
				}
			}
		})
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package remove

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/core"
)

func TestRunRemove(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		force       bool
		save        bool
		setupFunc   func(t *testing.T, tmpDir string)
		wantErr     bool
		checkFunc   func(t *testing.T, tmpDir string)
	}{
		{
			name:        "remove existing command without save",
			commandName: "test-cmd",
			force:       true,
			save:        false,
			setupFunc: func(t *testing.T, tmpDir string) {
				// Create command directory structure
				require.NoError(t, os.MkdirAll(filepath.Join(".claude", "commands", "test-cmd"), 0755))

				// Setup lock file with command
				lockPath := "ccmd-lock.yaml"
				lockFile := &struct {
					Version  string                 `yaml:"version"`
					Commands map[string]interface{} `yaml:"commands"`
				}{
					Version: "1.0",
					Commands: map[string]interface{}{
						"test-cmd": map[string]interface{}{
							"version":      "v1.0.0",
							"repository":   "git@github.com:test/test-cmd.git",
							"commit":       "1234567890abcdef1234567890abcdef12345678",
							"installed_at": time.Now().Format(time.RFC3339),
							"updated_at":   time.Now().Format(time.RFC3339),
						},
					},
				}
				data, err := yaml.Marshal(lockFile)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(lockPath, data, 0644))
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tmpDir string) {
				// Command should be removed from lock file
				lockPath := "ccmd-lock.yaml"
				data, err := os.ReadFile(lockPath)
				require.NoError(t, err)

				var lock struct {
					Commands map[string]interface{} `yaml:"commands"`
				}
				err = yaml.Unmarshal(data, &lock)
				require.NoError(t, err)
				_, exists := lock.Commands["test-cmd"]
				assert.False(t, exists)
			},
		},
		{
			name:        "remove existing command with save flag",
			commandName: "test-cmd",
			force:       true,
			save:        true,
			setupFunc: func(t *testing.T, tmpDir string) {
				// Create command directory structure
				require.NoError(t, os.MkdirAll(filepath.Join(".claude", "commands", "test-cmd"), 0755))

				// Create ccmd.yaml
				config := &core.ProjectConfig{
					Commands: []string{
						"test/test-cmd@v1.0.0",
					},
				}
				require.NoError(t, core.SaveProjectConfig(".", config))

				// Create ccmd-lock.yaml directly in the format expected by core
				lockData := map[string]interface{}{
					"commands": map[string]interface{}{
						"test-cmd": map[string]interface{}{
							"repository":   "test/test-cmd",
							"version":      "v1.0.0",
							"commit":       "abcdef1234567890abcdef1234567890abcdef12",
							"installed_at": time.Now().Format(time.RFC3339),
							"updated_at":   time.Now().Format(time.RFC3339),
						},
					},
				}

				lockYAML, err := yaml.Marshal(lockData)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile("ccmd-lock.yaml", lockYAML, 0644))
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tmpDir string) {
				// Check ccmd.yaml - command should be removed
				config, err := core.LoadProjectConfig(".")
				require.NoError(t, err)
				assert.Empty(t, config.Commands)

				// Check ccmd-lock.yaml - command should be removed
				lockData, err := os.ReadFile("ccmd-lock.yaml")
				require.NoError(t, err)

				var lockYAML map[string]interface{}
				require.NoError(t, yaml.Unmarshal(lockData, &lockYAML))

				commands, ok := lockYAML["commands"].(map[string]interface{})
				if ok {
					_, exists := commands["test-cmd"]
					assert.False(t, exists)
				}
			},
		},
		{
			name:        "remove non-existent command",
			commandName: "non-existent",
			force:       true,
			save:        false,
			setupFunc:   func(t *testing.T, tmpDir string) {},
			wantErr:     true,
			checkFunc:   func(t *testing.T, tmpDir string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "ccmd-remove-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			defer os.Chdir(oldWd)

			// Create .claude directory
			require.NoError(t, os.MkdirAll(".claude", 0755))
			require.NoError(t, os.MkdirAll(filepath.Join(".claude", "commands"), 0755))

			// Create initial lock file
			lockContent := `version: "1.0"
lockfileVersion: 1
commands: {}`
			require.NoError(t, os.WriteFile("ccmd-lock.yaml", []byte(lockContent), 0644))

			// Run setup
			tt.setupFunc(t, tmpDir)

			// Execute
			err = runRemove(tt.commandName, tt.force, tt.save)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Run checks
			tt.checkFunc(t, tmpDir)
		})
	}
}

/*
func TestUpdateProjectFiles(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		cmdInfo     *project.CommandLockInfo
		setupFunc   func(t *testing.T, tmpDir string)
		wantErr     bool
		checkFunc   func(t *testing.T, tmpDir string)
	}{
		{
			name:        "update both ccmd.yaml and ccmd-lock.yaml",
			commandName: "test-cmd",
			cmdInfo: &project.CommandLockInfo{
				Name:        "test-cmd",
				Version:     "v1.0.0",
				Source:      "git@github.com:test/test-cmd.git",
				Resolved:    "git@github.com:test/test-cmd.git@v1.0.0",
				Commit:      "1234567890abcdef1234567890abcdef12345678",
				InstalledAt: time.Now(),
				UpdatedAt:   time.Now(),
			},
			setupFunc: func(t *testing.T, tmpDir string) {
				manager := project.NewManager(".")

				// Create ccmd.yaml with the command
				config := &project.Config{
					Commands: []string{
						"test/test-cmd@v1.0.0",
						"test/other-cmd@v2.0.0",
					},
				}
				require.NoError(t, manager.SaveConfig(config))

				// Create ccmd-lock.yaml with the command
				lockFile := project.NewLockFile()
				cmd1 := &project.Command{
					Name:        "test-cmd",
					Source:      "https://github.com/test/test-cmd.git",
					Resolved:    "https://github.com/test/test-cmd.git@v1.0.0",
					Version:     "v1.0.0",
					Commit:      "abcdef1234567890abcdef1234567890abcdef12",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				cmd2 := &project.Command{
					Name:        "other-cmd",
					Source:      "https://github.com/test/other-cmd.git",
					Resolved:    "https://github.com/test/other-cmd.git@v2.0.0",
					Version:     "v2.0.0",
					Commit:      "def4567890abcdef1234567890abcdef12345678",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
				}
				require.NoError(t, lockFile.AddCommand(cmd1))
				require.NoError(t, lockFile.AddCommand(cmd2))
				require.NoError(t, manager.SaveLockFile(lockFile))
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tmpDir string) {
				manager := project.NewManager(".")

				// Check ccmd.yaml - only other-cmd should remain
				config, err := manager.LoadConfig()
				require.NoError(t, err)
				commands, err := config.GetCommands()
				require.NoError(t, err)
				assert.Len(t, commands, 1)
				assert.Equal(t, "test/other-cmd", commands[0].Repo)

				// Check ccmd-lock.yaml - only other-cmd should remain
				lockFile, err := manager.LoadLockFile()
				require.NoError(t, err)
				_, exists := lockFile.GetCommand("test-cmd")
				assert.False(t, exists)
				cmd, exists := lockFile.GetCommand("other-cmd")
				assert.True(t, exists)
				assert.Equal(t, "other-cmd", cmd.Name)
			},
		},
		{
			name:        "command info without repository metadata",
			commandName: "test-cmd",
			cmdInfo: &project.CommandLockInfo{
				Name:    "test-cmd",
				Version: "v1.0.0",
			},
			setupFunc: func(t *testing.T, tmpDir string) {},
			wantErr:   true,
			checkFunc: func(t *testing.T, tmpDir string) {},
		},
		{
			name:        "no project files exist",
			commandName: "test-cmd",
			cmdInfo: &project.CommandLockInfo{
				Name:        "test-cmd",
				Version:     "v1.0.0",
				Source:      "git@github.com:test/test-cmd.git",
				Resolved:    "test/test-cmd@v1.0.0",
				Commit:      "fedcba0987654321fedcba0987654321fedcba09",
				InstalledAt: time.Now(),
				UpdatedAt:   time.Now(),
				Metadata:    map[string]string{"repository": "test/test-cmd"},
			},
			setupFunc: func(t *testing.T, tmpDir string) {},
			wantErr:   false,
			checkFunc: func(t *testing.T, tmpDir string) {
				manager := project.NewManager(".")
				assert.False(t, manager.ConfigExists())
				assert.False(t, manager.LockExists())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "ccmd-update-files-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			defer os.Chdir(oldWd)

			// Run setup
			tt.setupFunc(t, tmpDir)

			// Execute
			err = updateProjectFiles(tt.commandName, tt.cmdInfo)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Run checks
			tt.checkFunc(t, tmpDir)
		})
	}
}
*/

/*
func TestExtractRepoFromSource(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "git SSH URL",
			source:  "git@github.com:gifflet/hello-world.git",
			want:    "gifflet/hello-world",
			wantErr: false,
		},
		{
			name:    "git SSH URL without .git suffix",
			source:  "git@github.com:owner/repo",
			want:    "owner/repo",
			wantErr: false,
		},
		{
			name:    "HTTPS URL",
			source:  "https://github.com/gifflet/hello-world.git",
			want:    "gifflet/hello-world",
			wantErr: false,
		},
		{
			name:    "HTTPS URL without .git suffix",
			source:  "https://github.com/owner/repo",
			want:    "owner/repo",
			wantErr: false,
		},
		{
			name:    "HTTP URL",
			source:  "http://github.com/owner/repo.git",
			want:    "owner/repo",
			wantErr: false,
		},
		{
			name:    "empty source",
			source:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid git URL format",
			source:  "git@github.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid HTTPS URL format",
			source:  "https://github.com/",
			want:    "",
			wantErr: true,
		},
		{
			name:    "unsupported URL format",
			source:  "ftp://example.com/repo.git",
			want:    "",
			wantErr: true,
		},
		{
			name:    "git URL with spaces (trimmed)",
			source:  "  git@github.com:owner/repo.git  ",
			want:    "owner/repo",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractRepoFromSource(tt.source)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
*/

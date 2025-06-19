package remove

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/project"
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
				// Setup lock file with command
				lockManager := lock.NewManagerWithFS(".claude", fs.OS{})
				require.NoError(t, lockManager.Load())
				cmd := &models.Command{
					Name:     "test-cmd",
					Version:  "v1.0.0",
					Metadata: map[string]string{"repository": "test/test-cmd"},
				}
				require.NoError(t, lockManager.AddCommand(cmd))
				require.NoError(t, lockManager.Save())
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tmpDir string) {
				// Command should be removed from lock file
				exists, err := commands.CommandExists("test-cmd", "", nil)
				require.NoError(t, err)
				assert.False(t, exists)
			},
		},
		{
			name:        "remove existing command with save flag",
			commandName: "test-cmd",
			force:       true,
			save:        true,
			setupFunc: func(t *testing.T, tmpDir string) {
				// Create ccmd.yaml
				manager := project.NewManager(".")
				config := &project.Config{
					Commands: []project.ConfigCommand{
						{
							Repo:    "test/test-cmd",
							Version: "v1.0.0",
						},
					},
				}
				require.NoError(t, manager.SaveConfig(config))

				// Create ccmd-lock.yaml with command
				lockFile := project.NewLockFile()
				cmd := &project.Command{
					Name:        "test-cmd",
					Repository:  "https://github.com/test/test-cmd.git",
					Version:     "v1.0.0",
					CommitHash:  "abcdef1234567890abcdef1234567890abcdef12",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
					FileSize:    1024,
					Checksum:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
					Metadata:    map[string]string{"repository": "test/test-cmd"},
				}
				require.NoError(t, lockFile.AddCommand(cmd))
				require.NoError(t, manager.SaveLockFile(lockFile))

				// Also add to .claude lock file for the command to exist
				lockManager := lock.NewManagerWithFS(".claude", fs.OS{})
				require.NoError(t, lockManager.Load())
				internalCmd := &models.Command{
					Name:     "test-cmd",
					Version:  "v1.0.0",
					Metadata: map[string]string{"repository": "test/test-cmd"},
				}
				require.NoError(t, lockManager.AddCommand(internalCmd))
				require.NoError(t, lockManager.Save())
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tmpDir string) {
				manager := project.NewManager(".")

				// Check ccmd.yaml - command should be removed
				config, err := manager.LoadConfig()
				require.NoError(t, err)
				assert.Empty(t, config.Commands)

				// Check ccmd-lock.yaml - command should be removed
				lockFile, err := manager.LoadLockFile()
				require.NoError(t, err)
				_, exists := lockFile.GetCommand("test-cmd")
				assert.False(t, exists)
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
			lockContent := `commands: {}`
			require.NoError(t, os.WriteFile(filepath.Join(".claude", "ccmd-lock.yaml"), []byte(lockContent), 0644))

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

func TestUpdateProjectFiles(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		cmdInfo     *models.Command
		setupFunc   func(t *testing.T, tmpDir string)
		wantErr     bool
		checkFunc   func(t *testing.T, tmpDir string)
	}{
		{
			name:        "update both ccmd.yaml and ccmd-lock.yaml",
			commandName: "test-cmd",
			cmdInfo: &models.Command{
				Name:     "test-cmd",
				Version:  "v1.0.0",
				Metadata: map[string]string{"repository": "test/test-cmd"},
			},
			setupFunc: func(t *testing.T, tmpDir string) {
				manager := project.NewManager(".")

				// Create ccmd.yaml with the command
				config := &project.Config{
					Commands: []project.ConfigCommand{
						{
							Repo:    "test/test-cmd",
							Version: "v1.0.0",
						},
						{
							Repo:    "test/other-cmd",
							Version: "v2.0.0",
						},
					},
				}
				require.NoError(t, manager.SaveConfig(config))

				// Create ccmd-lock.yaml with the command
				lockFile := project.NewLockFile()
				cmd1 := &project.Command{
					Name:        "test-cmd",
					Repository:  "https://github.com/test/test-cmd.git",
					Version:     "v1.0.0",
					CommitHash:  "abcdef1234567890abcdef1234567890abcdef12",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
					FileSize:    1024,
					Checksum:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				}
				cmd2 := &project.Command{
					Name:        "other-cmd",
					Repository:  "https://github.com/test/other-cmd.git",
					Version:     "v2.0.0",
					CommitHash:  "def4567890abcdef1234567890abcdef12345678",
					InstalledAt: time.Now(),
					UpdatedAt:   time.Now(),
					FileSize:    2048,
					Checksum:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
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
				assert.Len(t, config.Commands, 1)
				assert.Equal(t, "test/other-cmd", config.Commands[0].Repo)

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
			cmdInfo: &models.Command{
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
			cmdInfo: &models.Command{
				Name:     "test-cmd",
				Version:  "v1.0.0",
				Metadata: map[string]string{"repository": "test/test-cmd"},
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

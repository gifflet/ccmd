package update

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestVersionNeedsUpdate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{
			name:    "same version",
			current: "v1.0.0",
			latest:  "v1.0.0",
			want:    false,
			wantErr: false,
		},
		{
			name:    "newer version available",
			current: "v1.0.0",
			latest:  "v1.1.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "older version available",
			current: "v2.0.0",
			latest:  "v1.0.0",
			want:    false,
			wantErr: false,
		},
		{
			name:    "patch update available",
			current: "v1.0.0",
			latest:  "v1.0.1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "major update available",
			current: "v1.0.0",
			latest:  "v2.0.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "current is commit, latest is semver",
			current: "abc1234",
			latest:  "v1.0.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "current is semver, latest is commit",
			current: "v1.0.0",
			latest:  "abc1234",
			want:    false,
			wantErr: false,
		},
		{
			name:    "both are commits",
			current: "abc1234",
			latest:  "def5678",
			want:    false,
			wantErr: true,
		},
		{
			name:    "same commits",
			current: "abc1234",
			latest:  "abc1234",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := versionNeedsUpdate(tt.current, tt.latest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUpdateCommand(t *testing.T) {
	memfs := fs.NewMemFS()
	baseDir := "/home/user/.claude"

	// Create directory structure
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands", "test-cmd"), 0o755))

	// Create lock file with a command
	lockContent := models.LockFile{
		Version: "1.0",
		Commands: map[string]*models.Command{
			"test-cmd": {
				Name:        "test-cmd",
				Version:     "v1.0.0",
				Source:      "https://github.com/user/test-cmd",
				InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	lockData, _ := json.Marshal(lockContent)
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands.lock"), lockData, 0o644))

	// Create command structure
	commandDir := filepath.Join(baseDir, "commands", "test-cmd")
	require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), []byte(`
name: test-cmd
description: Test command
author: Test Author
repository: https://github.com/user/test-cmd
entry: test.sh
`), 0o644))

	require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "index.md"), []byte("# Test Command"), 0o644))
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands", "test-cmd.md"), []byte("# Test Command"), 0o644))

	t.Run("command not installed", func(t *testing.T) {
		result := updateCommand("nonexistent", baseDir, memfs, "", false)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "not installed")
	})

	t.Run("command exists", func(t *testing.T) {
		// This test would need a mock git client to fully test
		// For now, we test the initial checks
		result := updateCommand("test-cmd", baseDir, memfs, "", false)
		assert.Equal(t, "test-cmd", result.Name)
		assert.Equal(t, "v1.0.0", result.CurrentVersion)
		// The actual update would fail due to git operations
		assert.Error(t, result.Error)
	})
}

func TestUpdateAllCommands(t *testing.T) {
	// Set test environment to disable spinners
	t.Setenv("GO_TEST", "1")

	memfs := fs.NewMemFS()
	baseDir := "/home/user/.claude"

	// Create directory structure
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands"), 0o755))

	t.Run("no commands installed", func(t *testing.T) {
		// Create empty lock file
		lockContent := models.LockFile{
			Version:  "1.0",
			Commands: map[string]*models.Command{},
		}
		lockData, _ := json.Marshal(lockContent)
		require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands.lock"), lockData, 0o644))

		err := updateAllCommands(baseDir, memfs, false)
		assert.NoError(t, err)
	})

	t.Run("multiple commands", func(t *testing.T) {
		// Create lock file with multiple commands
		lockContent := models.LockFile{
			Version: "1.0",
			Commands: map[string]*models.Command{
				"cmd1": {
					Name:        "cmd1",
					Version:     "v1.0.0",
					Source:      "https://github.com/user/cmd1",
					InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				"cmd2": {
					Name:        "cmd2",
					Version:     "v2.0.0",
					Source:      "https://github.com/user/cmd2",
					InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		}
		lockData, _ := json.Marshal(lockContent)
		require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands.lock"), lockData, 0o644))

		// Create command structures
		for _, cmdName := range []string{"cmd1", "cmd2"} {
			commandDir := filepath.Join(baseDir, "commands", cmdName)
			require.NoError(t, memfs.MkdirAll(commandDir, 0o755))
			require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "ccmd.yaml"), []byte(fmt.Sprintf(`
name: %s
description: Test command
`, cmdName)), 0o644))
			require.NoError(t, memfs.WriteFile(filepath.Join(commandDir, "index.md"), []byte("# Test"), 0o644))
			require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands", cmdName+".md"), []byte("# Test"), 0o644))
		}

		// This would fail due to git operations, but we test the flow
		err := updateAllCommands(baseDir, memfs, false)
		assert.Error(t, err) // Expected due to missing git operations
	})
}

func TestRunUpdateWithFS(t *testing.T) {
	memfs := fs.NewMemFS()

	t.Run("no arguments without --all", func(t *testing.T) {
		err := runUpdateWithFS([]string{}, false, "", false, memfs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command name required")
	})

	t.Run("argument with --all", func(t *testing.T) {
		err := runUpdateWithFS([]string{"cmd"}, true, "", false, memfs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot specify command name with --all")
	})
}

func TestDisplayResult(t *testing.T) {
	tests := []struct {
		name   string
		result Result
	}{
		{
			name: "successful update",
			result: Result{
				Name:           "test-cmd",
				CurrentVersion: "v1.0.0",
				NewVersion:     "v1.1.0",
				Updated:        true,
			},
		},
		{
			name: "already up to date",
			result: Result{
				Name:           "test-cmd",
				CurrentVersion: "v1.0.0",
				NewVersion:     "v1.0.0",
				Updated:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			displayResult(tt.result)
		})
	}
}

func TestPerformUpdate(t *testing.T) {
	memfs := fs.NewMemFS()
	baseDir := "/home/user/.claude"

	// Create directory structure
	require.NoError(t, memfs.MkdirAll(filepath.Join(baseDir, "commands", "test-cmd"), 0o755))

	// Create lock file
	lockContent := models.LockFile{
		Version: "1.0",
		Commands: map[string]*models.Command{
			"test-cmd": {
				Name:        "test-cmd",
				Version:     "v1.0.0",
				Source:      "https://github.com/user/test-cmd",
				InstalledAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	lockData, _ := json.Marshal(lockContent)
	require.NoError(t, memfs.WriteFile(filepath.Join(baseDir, "commands.lock"), lockData, 0o644))

	cmdInfo := &models.Command{
		Name:        "test-cmd",
		Version:     "v1.0.0",
		Source:      "https://github.com/user/test-cmd",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// This will fail due to missing git operations, but we test the flow
	err := performUpdate(cmdInfo, "v1.1.0", baseDir, memfs)
	assert.Error(t, err) // Expected due to missing git operations
}

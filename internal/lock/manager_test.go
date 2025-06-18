package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestNewManager(t *testing.T) {
	dir := "/test/dir"
	manager := NewManager(dir)

	expectedPath := filepath.Join(dir, LockFileName)
	if manager.filePath != expectedPath {
		t.Errorf("expected file path %s, got %s", expectedPath, manager.filePath)
	}
}

func TestManager_Load(t *testing.T) {
	tests := []struct {
		name         string
		setupFS      func(*fs.MemFS) error
		wantErr      bool
		wantCommands int
	}{
		{
			name: "load existing lock file",
			setupFS: func(memFS *fs.MemFS) error {
				lockFile := models.LockFile{
					Version: LockFileVersion,
					Commands: map[string]*models.Command{
						"test-cmd": {
							Name:    "test-cmd",
							Version: "1.0.0",
							Source:  "github.com/example/test-cmd",
						},
					},
				}
				data, _ := json.Marshal(lockFile)
				return memFS.WriteFile("commands.lock", data, 0o644)
			},
			wantErr:      false,
			wantCommands: 1,
		},
		{
			name:         "create new lock file when not exists",
			setupFS:      func(*fs.MemFS) error { return nil },
			wantErr:      false,
			wantCommands: 0,
		},
		{
			name: "error on invalid JSON",
			setupFS: func(memFS *fs.MemFS) error {
				return memFS.WriteFile("commands.lock", []byte("invalid json"), 0o644)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := fs.NewMemFS()
			if err := tt.setupFS(memFS); err != nil {
				t.Fatalf("failed to setup filesystem: %v", err)
			}

			manager := NewManagerWithFS(".", memFS)
			err := manager.Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(manager.lockFile.Commands) != tt.wantCommands {
					t.Errorf("expected %d commands, got %d", tt.wantCommands, len(manager.lockFile.Commands))
				}
			}
		})
	}
}

func TestManager_Save(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Manager) error
		wantErr bool
		verify  func(fs.FileSystem, string) error
	}{
		{
			name: "save lock file successfully",
			setup: func(m *Manager) error {
				m.lockFile = &models.LockFile{
					Version: LockFileVersion,
					Commands: map[string]*models.Command{
						"test-cmd": {
							Name:    "test-cmd",
							Version: "1.0.0",
						},
					},
				}
				m.loaded = true
				return nil
			},
			wantErr: false,
			verify: func(fileSystem fs.FileSystem, path string) error {
				data, err := fileSystem.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read lock file: %v", err)
				}

				var lockFile models.LockFile
				if err := json.Unmarshal(data, &lockFile); err != nil {
					return fmt.Errorf("failed to unmarshal lock file: %v", err)
				}

				if len(lockFile.Commands) != 1 {
					return fmt.Errorf("expected 1 command, got %d", len(lockFile.Commands))
				}
				return nil
			},
		},
		{
			name:    "error when not loaded",
			setup:   func(m *Manager) error { return nil },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := fs.NewMemFS()
			manager := NewManagerWithFS(".", memFS)

			if err := tt.setup(manager); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := manager.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				if err := tt.verify(memFS, manager.filePath); err != nil {
					t.Errorf("verification failed: %v", err)
				}
			}
		})
	}
}

func TestManager_AddCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *models.Command
		wantErr bool
	}{
		{
			name: "add valid command",
			cmd: &models.Command{
				Name:    "test-cmd",
				Version: "1.0.0",
				Source:  "github.com/example/test-cmd",
			},
			wantErr: false,
		},
		{
			name: "error on empty name",
			cmd: &models.Command{
				Version: "1.0.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFS := fs.NewMemFS()
			manager := NewManagerWithFS(".", memFS)
			_ = manager.Load()

			err := manager.AddCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !manager.HasCommand(tt.cmd.Name) {
					t.Errorf("command %s not found after adding", tt.cmd.Name)
				}

				cmd, _ := manager.GetCommand(tt.cmd.Name)
				if cmd.InstalledAt.IsZero() {
					t.Error("InstalledAt not set")
				}
				if cmd.UpdatedAt.IsZero() {
					t.Error("UpdatedAt not set")
				}
			}
		})
	}
}

func TestManager_UpdateCommand(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add a command first
	cmd := &models.Command{
		Name:    "test-cmd",
		Version: "1.0.0",
	}
	_ = manager.AddCommand(cmd)

	originalUpdatedAt := manager.lockFile.Commands["test-cmd"].UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Update the command
	err := manager.UpdateCommand("test-cmd", func(c *models.Command) error {
		c.Version = "2.0.0"
		return nil
	})

	if err != nil {
		t.Errorf("UpdateCommand() error = %v", err)
		return
	}

	updated, _ := manager.GetCommand("test-cmd")
	if updated.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", updated.Version)
	}

	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt not updated")
	}
}

func TestManager_RemoveCommand(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add a command first
	cmd := &models.Command{
		Name:    "test-cmd",
		Version: "1.0.0",
	}
	_ = manager.AddCommand(cmd)

	// Remove the command
	err := manager.RemoveCommand("test-cmd")
	if err != nil {
		t.Errorf("RemoveCommand() error = %v", err)
		return
	}

	if manager.HasCommand("test-cmd") {
		t.Error("command still exists after removal")
	}

	// Try to remove non-existent command
	err = manager.RemoveCommand("non-existent")
	if err == nil {
		t.Error("expected error when removing non-existent command")
	}
}

func TestManager_ListCommands(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add multiple commands
	commands := []*models.Command{
		{Name: "cmd1", Version: "1.0.0"},
		{Name: "cmd2", Version: "2.0.0"},
		{Name: "cmd3", Version: "3.0.0"},
	}

	for _, cmd := range commands {
		_ = manager.AddCommand(cmd)
	}

	list, err := manager.ListCommands()
	if err != nil {
		t.Errorf("ListCommands() error = %v", err)
		return
	}

	if len(list) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(list))
	}

	// Verify all commands are present
	nameMap := make(map[string]bool)
	for _, cmd := range list {
		nameMap[cmd.Name] = true
	}

	for _, cmd := range commands {
		if !nameMap[cmd.Name] {
			t.Errorf("command %s not found in list", cmd.Name)
		}
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Run concurrent operations
	done := make(chan bool)
	errors := make(chan error, 10)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			cmd := &models.Command{
				Name:    fmt.Sprintf("cmd%d", i),
				Version: "1.0.0",
			}
			if err := manager.AddCommand(cmd); err != nil {
				errors <- err
			}
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			_, _ = manager.ListCommands()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	close(errors)
	for err := range errors {
		t.Errorf("concurrent operation error: %v", err)
	}
}

func TestManager_BackupCreation(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add a command and save
	cmd := &models.Command{
		Name:    "test-cmd",
		Version: "1.0.0",
	}
	_ = manager.AddCommand(cmd)
	_ = manager.Save()

	// Modify and save again
	_ = manager.UpdateCommand("test-cmd", func(c *models.Command) error {
		c.Version = "2.0.0"
		return nil
	})
	_ = manager.Save()

	// Check if backup exists
	backupPath := manager.filePath + ".bak"
	if _, err := memFS.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file not created")
	}
}

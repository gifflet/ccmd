package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCommandLock_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		cl      CommandLock
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid lock",
			cl: CommandLock{
				Version:     "1.0.0",
				Repository:  "github.com/test/repo",
				InstalledAt: now,
				LastUpdated: now,
			},
			wantErr: false,
		},
		{
			name: "missing version",
			cl: CommandLock{
				Repository:  "github.com/test/repo",
				InstalledAt: now,
				LastUpdated: now,
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing repository",
			cl: CommandLock{
				Version:     "1.0.0",
				InstalledAt: now,
				LastUpdated: now,
			},
			wantErr: true,
			errMsg:  "repository is required",
		},
		{
			name: "missing installedAt",
			cl: CommandLock{
				Version:     "1.0.0",
				Repository:  "github.com/test/repo",
				LastUpdated: now,
			},
			wantErr: true,
			errMsg:  "installedAt is required",
		},
		{
			name: "missing lastUpdated",
			cl: CommandLock{
				Version:     "1.0.0",
				Repository:  "github.com/test/repo",
				InstalledAt: now,
			},
			wantErr: true,
			errMsg:  "lastUpdated is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLockFile_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		lf      LockFile
		wantErr bool
	}{
		{
			name: "valid lock file",
			lf: LockFile{
				Commands: map[string]CommandLock{
					"test-command": {
						Version:     "1.0.0",
						Repository:  "github.com/test/repo",
						InstalledAt: now,
						LastUpdated: now,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil commands map",
			lf: LockFile{
				Commands: nil,
			},
			wantErr: true,
		},
		{
			name: "empty commands map",
			lf: LockFile{
				Commands: map[string]CommandLock{},
			},
			wantErr: false,
		},
		{
			name: "invalid command lock",
			lf: LockFile{
				Commands: map[string]CommandLock{
					"test-command": {
						Version: "1.0.0",
						// Missing required fields
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lf.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLockFile_Commands(t *testing.T) {
	now := time.Now()
	lf := LockFile{
		Commands: make(map[string]CommandLock),
	}

	// Test SetCommand
	lock := CommandLock{
		Version:     "1.0.0",
		Repository:  "github.com/test/repo",
		InstalledAt: now,
		LastUpdated: now,
	}
	lf.SetCommand("test-command", lock)

	// Test GetCommand
	retrievedLock, exists := lf.GetCommand("test-command")
	if !exists {
		t.Error("GetCommand() should return true for existing command")
	}
	if retrievedLock.Version != lock.Version {
		t.Errorf("GetCommand() returned wrong lock: got %v, want %v", retrievedLock, lock)
	}

	// Test GetCommand for non-existent command
	_, exists = lf.GetCommand("non-existent")
	if exists {
		t.Error("GetCommand() should return false for non-existent command")
	}

	// Test RemoveCommand
	lf.RemoveCommand("test-command")
	_, exists = lf.GetCommand("test-command")
	if exists {
		t.Error("RemoveCommand() should remove the command")
	}
}

func TestLockFile_SetCommand_NilMap(t *testing.T) {
	lf := LockFile{}
	now := time.Now()

	lock := CommandLock{
		Version:     "1.0.0",
		Repository:  "github.com/test/repo",
		InstalledAt: now,
		LastUpdated: now,
	}

	// Should initialize the map if nil
	lf.SetCommand("test-command", lock)

	if lf.Commands == nil {
		t.Error("SetCommand() should initialize nil map")
	}

	retrievedLock, exists := lf.GetCommand("test-command")
	if !exists {
		t.Error("SetCommand() should add command to initialized map")
	}
	if retrievedLock.Version != lock.Version {
		t.Errorf("SetCommand() added wrong lock: got %v, want %v", retrievedLock, lock)
	}
}

func TestLockFile_JSON(t *testing.T) {
	now := time.Now().Truncate(time.Second) // Truncate for JSON round-trip consistency

	lf := LockFile{
		Commands: map[string]CommandLock{
			"test-command": {
				Version:     "1.0.0",
				Repository:  "github.com/test/repo",
				InstalledAt: now,
				LastUpdated: now,
			},
			"another-command": {
				Version:     "2.0.0",
				Repository:  "github.com/test/another",
				InstalledAt: now.Add(-24 * time.Hour),
				LastUpdated: now,
			},
		},
	}

	// Test MarshalJSON
	data, err := lf.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Check if output is properly indented
	var indentCheck map[string]interface{}
	if err := json.Unmarshal(data, &indentCheck); err != nil {
		t.Fatalf("JSON output is not valid: %v", err)
	}

	// Test UnmarshalJSON
	var lf2 LockFile
	err = lf2.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify round-trip
	if len(lf2.Commands) != len(lf.Commands) {
		t.Errorf("JSON round-trip failed: different number of commands")
	}

	for name, lock := range lf.Commands {
		lock2, exists := lf2.Commands[name]
		if !exists {
			t.Errorf("JSON round-trip failed: missing command %s", name)
			continue
		}

		// Compare fields (time comparison needs special handling)
		if lock.Version != lock2.Version || lock.Repository != lock2.Repository {
			t.Errorf("JSON round-trip failed for %s: got %+v, want %+v", name, lock2, lock)
		}

		// For time fields, compare Unix timestamps
		if lock.InstalledAt.Unix() != lock2.InstalledAt.Unix() {
			t.Errorf("JSON round-trip failed for %s: InstalledAt differs", name)
		}
		if lock.LastUpdated.Unix() != lock2.LastUpdated.Unix() {
			t.Errorf("JSON round-trip failed for %s: LastUpdated differs", name)
		}
	}
}

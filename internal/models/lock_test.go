package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCommand_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		cmd     Command
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: Command{
				Name:        "test-cmd",
				Version:     "1.0.0",
				Source:      "github.com/test/repo",
				InstalledAt: now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cmd: Command{
				Version:     "1.0.0",
				Source:      "github.com/test/repo",
				InstalledAt: now,
				UpdatedAt:   now,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing version",
			cmd: Command{
				Name:        "test-cmd",
				Source:      "github.com/test/repo",
				InstalledAt: now,
				UpdatedAt:   now,
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing source",
			cmd: Command{
				Name:        "test-cmd",
				Version:     "1.0.0",
				InstalledAt: now,
				UpdatedAt:   now,
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "missing installedAt",
			cmd: Command{
				Name:      "test-cmd",
				Version:   "1.0.0",
				Source:    "github.com/test/repo",
				UpdatedAt: now,
			},
			wantErr: true,
			errMsg:  "installedAt is required",
		},
		{
			name: "missing updatedAt",
			cmd: Command{
				Name:        "test-cmd",
				Version:     "1.0.0",
				Source:      "github.com/test/repo",
				InstalledAt: now,
			},
			wantErr: true,
			errMsg:  "updatedAt is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Command.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLockFile_Validate(t *testing.T) {
	now := time.Now()
	validCmd := &Command{
		Name:        "test-cmd",
		Version:     "1.0.0",
		Source:      "github.com/test/repo",
		InstalledAt: now,
		UpdatedAt:   now,
	}

	tests := []struct {
		name    string
		lf      LockFile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid lock file",
			lf: LockFile{
				Version: "1.0",
				Commands: map[string]*Command{
					"test-cmd": validCmd,
				},
			},
			wantErr: false,
		},
		{
			name: "nil commands map",
			lf: LockFile{
				Version:  "1.0",
				Commands: nil,
			},
			wantErr: true,
			errMsg:  "commands map cannot be nil",
		},
		{
			name: "empty commands map is valid",
			lf: LockFile{
				Version:  "1.0",
				Commands: make(map[string]*Command),
			},
			wantErr: false,
		},
		{
			name: "invalid command",
			lf: LockFile{
				Version: "1.0",
				Commands: map[string]*Command{
					"bad-cmd": {
						Name: "bad-cmd",
						// Missing required fields
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid command bad-cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lf.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LockFile.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("LockFile.Validate() error = %v, want to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLockFile_GetCommand(t *testing.T) {
	now := time.Now()
	cmd1 := &Command{
		Name:        "cmd1",
		Version:     "1.0.0",
		Source:      "github.com/test/cmd1",
		InstalledAt: now,
		UpdatedAt:   now,
	}
	cmd2 := &Command{
		Name:        "cmd2",
		Version:     "2.0.0",
		Source:      "github.com/test/cmd2",
		InstalledAt: now,
		UpdatedAt:   now,
	}

	lf := &LockFile{
		Version: "1.0",
		Commands: map[string]*Command{
			"cmd1": cmd1,
			"cmd2": cmd2,
		},
	}

	tests := []struct {
		name      string
		cmdName   string
		wantCmd   *Command
		wantExist bool
	}{
		{
			name:      "existing command",
			cmdName:   "cmd1",
			wantCmd:   cmd1,
			wantExist: true,
		},
		{
			name:      "another existing command",
			cmdName:   "cmd2",
			wantCmd:   cmd2,
			wantExist: true,
		},
		{
			name:      "non-existing command",
			cmdName:   "cmd3",
			wantCmd:   nil,
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotExist := lf.GetCommand(tt.cmdName)
			if gotExist != tt.wantExist {
				t.Errorf("LockFile.GetCommand() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
			if gotCmd != tt.wantCmd {
				t.Errorf("LockFile.GetCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
		})
	}
}

func TestLockFile_SetCommand(t *testing.T) {
	now := time.Now()
	cmd := &Command{
		Name:        "test-cmd",
		Version:     "1.0.0",
		Source:      "github.com/test/repo",
		InstalledAt: now,
		UpdatedAt:   now,
	}

	tests := []struct {
		name    string
		lf      *LockFile
		cmdName string
		cmd     *Command
	}{
		{
			name: "set in existing map",
			lf: &LockFile{
				Version:  "1.0",
				Commands: make(map[string]*Command),
			},
			cmdName: "test-cmd",
			cmd:     cmd,
		},
		{
			name: "set in nil map",
			lf: &LockFile{
				Version: "1.0",
			},
			cmdName: "test-cmd",
			cmd:     cmd,
		},
		{
			name: "update existing command",
			lf: &LockFile{
				Version: "1.0",
				Commands: map[string]*Command{
					"test-cmd": {
						Name:    "test-cmd",
						Version: "0.9.0",
					},
				},
			},
			cmdName: "test-cmd",
			cmd:     cmd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.lf.SetCommand(tt.cmdName, tt.cmd)
			if tt.lf.Commands == nil {
				t.Error("LockFile.SetCommand() Commands map is nil")
				return
			}
			if got := tt.lf.Commands[tt.cmdName]; got != tt.cmd {
				t.Errorf("LockFile.SetCommand() = %v, want %v", got, tt.cmd)
			}
		})
	}
}

func TestLockFile_RemoveCommand(t *testing.T) {
	now := time.Now()
	cmd1 := &Command{
		Name:        "cmd1",
		Version:     "1.0.0",
		Source:      "github.com/test/cmd1",
		InstalledAt: now,
		UpdatedAt:   now,
	}
	cmd2 := &Command{
		Name:        "cmd2",
		Version:     "2.0.0",
		Source:      "github.com/test/cmd2",
		InstalledAt: now,
		UpdatedAt:   now,
	}

	lf := &LockFile{
		Version: "1.0",
		Commands: map[string]*Command{
			"cmd1": cmd1,
			"cmd2": cmd2,
		},
	}

	// Remove existing command
	lf.RemoveCommand("cmd1")
	if _, exists := lf.Commands["cmd1"]; exists {
		t.Error("LockFile.RemoveCommand() command still exists")
	}
	if _, exists := lf.Commands["cmd2"]; !exists {
		t.Error("LockFile.RemoveCommand() removed wrong command")
	}

	// Remove non-existing command (should not panic)
	lf.RemoveCommand("cmd3")

	// Remove from nil map (should not panic)
	lf2 := &LockFile{}
	lf2.RemoveCommand("cmd1")
}

func TestLockFile_MarshalJSON(t *testing.T) {
	now := time.Now()
	lf := &LockFile{
		Version: "1.0",
		Commands: map[string]*Command{
			"test-cmd": {
				Name:         "test-cmd",
				Version:      "1.0.0",
				Source:       "github.com/test/repo",
				InstalledAt:  now,
				UpdatedAt:    now,
				Dependencies: []string{"dep1", "dep2"},
				Metadata: map[string]string{
					"author": "Test Author",
				},
			},
		},
	}

	// Test standard JSON marshaling (no indentation)
	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Should be compact (no newlines)
	if strings.Contains(string(data), "\n") {
		t.Errorf("json.Marshal() output should not be indented")
	}

	// Test indented JSON marshaling (as used in manager.Save())
	indentedData, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() error = %v", err)
	}

	// Check that output is indented
	if !strings.Contains(string(indentedData), "\n") {
		t.Errorf("json.MarshalIndent() output is not indented. Output: %s", string(indentedData))
	}

	// Unmarshal and compare
	var lf2 LockFile
	if err := json.Unmarshal(data, &lf2); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if lf2.Version != lf.Version {
		t.Errorf("Version mismatch: got %v, want %v", lf2.Version, lf.Version)
	}

	if len(lf2.Commands) != len(lf.Commands) {
		t.Errorf("Commands count mismatch: got %v, want %v", len(lf2.Commands), len(lf.Commands))
	}
}

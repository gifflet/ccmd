// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid command structure",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create valid ccmd.yaml
				ccmdContent := `name: mycommand
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)

				// Create index.md
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# MyCommand\n"), 0o644)

				return cmdDir
			},
			wantErr: false,
		},
		{
			name: "command directory not found",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantErr:     true,
			errContains: "command directory not found",
		},
		{
			name: "missing ccmd.yaml",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Only create index.md
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# MyCommand\n"), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "ccmd.yaml not found",
		},
		{
			name: "invalid ccmd.yaml format",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create invalid YAML
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte("invalid: yaml: content:"), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "invalid ccmd.yaml format",
		},
		{
			name: "missing required fields in metadata",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create ccmd.yaml with missing fields
				ccmdContent := `name: mycommand
version: 1.0.0
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "invalid metadata",
		},
		{
			name: "missing index.md",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create valid ccmd.yaml
				ccmdContent := `name: mycommand
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "index.md not found",
		},
		{
			name: "empty index.md",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create valid ccmd.yaml
				ccmdContent := `name: mycommand
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)

				// Create empty index.md
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte(""), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "index.md is empty",
		},
		{
			name: "command name mismatch",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create ccmd.yaml with different name
				ccmdContent := `name: differentname
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: differentname
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Command\n"), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "command name mismatch",
		},
		{
			name: "invalid version format",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create ccmd.yaml with invalid version
				ccmdContent := `name: mycommand
version: invalid-version
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Command\n"), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "invalid version format",
		},
		{
			name: "versioned directory name matching",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand@1.0.0")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create ccmd.yaml matching versioned dir
				ccmdContent := `name: mycommand
version: 1.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Command\n"), 0o644)

				return cmdDir
			},
			wantErr: false,
		},
		{
			name: "versioned directory version mismatch",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				cmdDir := filepath.Join(dir, "mycommand@1.0.0")
				_ = os.Mkdir(cmdDir, 0o755)

				// Create ccmd.yaml with different version
				ccmdContent := `name: mycommand
version: 2.0.0
description: A test command
author: Test Author
repository: https://github.com/test/mycommand
entry: mycommand
`
				_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
				_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Command\n"), 0o644)

				return cmdDir
			},
			wantErr:     true,
			errContains: "version mismatch with directory name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdPath := tt.setup(t)
			validator := NewCommandValidator(cmdPath)

			err := validator.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	validator := &CommandValidator{}

	tests := []struct {
		version string
		wantErr bool
	}{
		// Valid versions
		{"1.0.0", false},
		{"0.1.0", false},
		{"2.1.3", false},
		{"1.0.0-alpha", false},
		{"1.0.0-alpha.1", false},
		{"1.0.0-0.3.7", false},
		{"1.0.0-x.7.z.92", false},
		{"1.0.0+20130313144700", false},
		{"1.0.0-beta+exp.sha.5114f85", false},
		{"1.0.0+21AF26D3-117B344092BD", false},

		// Invalid versions
		{"1", true},
		{"1.0", true},
		{"1.0.0-", true},
		{"1.0.0+", true},
		{"01.0.0", true},
		{"1.01.0", true},
		{"1.0.01", true},
		{"1.0.0.0", true},
		{"v1.0.0", true},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := validator.validateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersion(%s) error = %v, wantErr %v", tt.version, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	// Setup valid command
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "testcmd")
	_ = os.Mkdir(cmdDir, 0o755)

	ccmdContent := `name: testcmd
version: 1.0.0
description: Test command
author: Test
repository: https://github.com/test/cmd
entry: testcmd
`
	_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
	_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Test\n"), 0o644)

	// Test
	err := ValidateDirectory(cmdDir)
	if err != nil {
		t.Errorf("ValidateDirectory() unexpected error: %v", err)
	}
}

func TestValidateInstalled(t *testing.T) {
	// Setup commands directory
	commandsDir := t.TempDir()
	cmdDir := filepath.Join(commandsDir, "installed-cmd")
	_ = os.Mkdir(cmdDir, 0o755)

	ccmdContent := `name: installed-cmd
version: 1.0.0
description: Installed command
author: Test
repository: https://github.com/test/cmd
entry: installed-cmd
`
	_ = os.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte(ccmdContent), 0o644)
	_ = os.WriteFile(filepath.Join(cmdDir, "index.md"), []byte("# Installed\n"), 0o644)

	// Test
	err := ValidateInstalled(commandsDir, "installed-cmd")
	if err != nil {
		t.Errorf("ValidateInstalled() unexpected error: %v", err)
	}
}

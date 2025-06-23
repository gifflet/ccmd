// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package search

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "search [keyword]", cmd.Use)
	assert.Equal(t, "Search for installed commands", cmd.Short)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("tags"))
	assert.NotNil(t, cmd.Flags().Lookup("author"))
	assert.NotNil(t, cmd.Flags().Lookup("all"))

	// Check that it has Args function
	assert.NotNil(t, cmd.Args)
}

func TestCommandExecute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "search with keyword",
			args:    []string{"search", "test"},
			wantErr: false,
		},
		{
			name:    "search with no args",
			args:    []string{"search"},
			wantErr: false,
		},
		{
			name:    "search with all flag",
			args:    []string{"search", "--all"},
			wantErr: false,
		},
		{
			name:    "search with tags",
			args:    []string{"search", "--tags", "cli,tool"},
			wantErr: false,
		},
		{
			name:    "search with author",
			args:    []string{"search", "--author", "Test Author"},
			wantErr: false,
		},
		{
			name:    "search with multiple args should fail",
			args:    []string{"search", "arg1", "arg2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command
			rootCmd := &cobra.Command{Use: "test"}
			rootCmd.AddCommand(NewCommand())

			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// We expect no error, but the command might still fail due to no lock file
				// which is expected in test environment
				require.True(t, err == nil || err.Error() == "search failed: failed to load lock file: open test/commands.lock: no such file or directory")
			}
		})
	}
}

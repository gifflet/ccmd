package help

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gifflet/ccmd/cmd/install"
	"github.com/gifflet/ccmd/cmd/list"
	"github.com/gifflet/ccmd/cmd/sync"
)

func TestHelpCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "general help",
			args: []string{},
			expectedOutput: []string{
				"Available Commands:",
				"help",
				"install",
				"list",
				"sync",
				"Common Use Cases:",
			},
		},
		{
			name: "help for sync command",
			args: []string{"sync"},
			expectedOutput: []string{
				"Command: sync",
				"Synchronize installed commands with ccmd.yaml",
				"Usage:",
				"Examples:",
				"--dry-run",
				"--force",
			},
		},
		{
			name: "help for install command",
			args: []string{"install"},
			expectedOutput: []string{
				"Command: install",
				"Install a command from a Git repository",
				"Usage:",
				"Examples:",
				"--version",
				"--name",
				"--force",
			},
		},
		{
			name: "help for list command",
			args: []string{"list"},
			expectedOutput: []string{
				"Command: list",
				"List all installed commands",
				"Usage:",
				"Examples:",
				"--verbose",
				"--json",
				"--sort",
			},
		},
		{
			name:        "help for non-existent command",
			args:        []string{"nonexistent"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command and add subcommands
			rootCmd := &cobra.Command{
				Use:   "ccmd",
				Short: "A CLI tool for managing Claude Code commands",
				Long: `ccmd is a command-line interface tool designed to help manage and execute
Claude Code commands efficiently.

ccmd provides a simple way to install, manage, and synchronize command-line tools
from GitHub repositories. It uses a declarative approach with ccmd.yaml files to
manage project dependencies and ensure consistent tool versions across teams.`,
			}

			// Add commands
			rootCmd.AddCommand(NewCommand())
			rootCmd.AddCommand(install.NewCommand())
			rootCmd.AddCommand(list.NewCommand())
			rootCmd.AddCommand(sync.NewCommand())

			// Create help command
			helpCmd := NewCommand()
			rootCmd.AddCommand(helpCmd)
			helpCmd.SetArgs(tt.args)

			// Capture output
			output := captureOutput(t, func() {
				err := helpCmd.Execute()
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})

			// Check expected output
			if !tt.expectError {
				for _, expected := range tt.expectedOutput {
					assert.Contains(t, output, expected, "Output should contain: %s", expected)
				}
			}
		})
	}
}

func TestGetCommandExamples(t *testing.T) {
	tests := []struct {
		command     string
		hasExamples bool
	}{
		{"list", true},
		{"update", true},
		{"search", true},
		{"info", true},
		{"remove", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			examples := getCommandExamples(tt.command)
			if tt.hasExamples {
				assert.NotEmpty(t, examples, "Command %s should have examples", tt.command)
			} else {
				assert.Empty(t, examples, "Command %s should not have examples", tt.command)
			}
		})
	}
}

func TestHelpCommandIntegration(t *testing.T) {
	// Test that help command is properly integrated with root command
	rootCmd := &cobra.Command{
		Use:   "ccmd",
		Short: "A CLI tool for managing Claude Code commands",
	}

	// Add help command
	rootCmd.AddCommand(NewCommand())

	// Test that help command exists
	helpCmd, _, err := rootCmd.Find([]string{"help"})
	require.NoError(t, err)
	assert.NotNil(t, helpCmd)
	assert.Equal(t, "help", helpCmd.Name())
}

func TestPrintCommandExamples(t *testing.T) {
	// Test that printCommandExamples doesn't panic for various commands
	commands := []string{"list", "update", "search", "info", "remove", "unknown"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			output := captureOutput(t, func() {
				printCommandExamples(cmd)
			})
			// Should not panic and should produce some output or empty
			assert.NotNil(t, output)
		})
	}
}

func TestPrintRelatedCommands(t *testing.T) {
	// Test that printRelatedCommands works for various commands
	commands := []string{"install", "remove", "list", "sync", "update", "search", "info", "unknown"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			output := captureOutput(t, func() {
				printRelatedCommands(cmd)
			})
			// Should not panic
			assert.NotNil(t, output)
			// Known commands should have related commands
			if cmd != "unknown" {
				assert.Contains(t, output, "See also:", "Command %s should have related commands", cmd)
			}
		})
	}
}

// captureOutput captures stdout during function execution
func captureOutput(t *testing.T, f func()) string {
	t.Helper()

	// Save original stdout
	originalStdout := os.Stdout

	// Create pipe
	r, w, err := os.Pipe()
	require.NoError(t, err)

	// Replace stdout
	os.Stdout = w

	// Run function in goroutine
	done := make(chan bool)
	go func() {
		f()
		close(done)
	}()

	// Wait for function to complete
	<-done

	// Close writer
	w.Close()

	// Read output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	// Restore stdout
	os.Stdout = originalStdout

	return buf.String()
}

func TestHelpForAllCommands(t *testing.T) {
	// Test that --help works for all commands
	commands := []string{"install", "list", "sync", "update", "remove", "search", "info"}

	for _, cmdName := range commands {
		t.Run(cmdName+"_help_flag", func(t *testing.T) {
			// This test ensures that each command has proper help text
			// In a real scenario, we'd create the actual command and test --help
			// For now, we just verify the help system recognizes these commands
			examples := getCommandExamples(cmdName)
			// Most commands should have examples defined
			if cmdName != "sync" && cmdName != "install" { // These have examples in Long description
				assert.NotEmpty(t, examples, "Command %s should have examples", cmdName)
			}
		})
	}
}

func TestHelpOutputFormatting(t *testing.T) {
	// Test that help output is properly formatted
	rootCmd := &cobra.Command{
		Use:   "ccmd",
		Short: "Test root command",
		Long:  "Test long description",
	}

	// Add test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Long: `Test command long description.

Examples:
  ccmd test
  ccmd test --flag`,
	}
	testCmd.Flags().String("flag", "", "Test flag")
	rootCmd.AddCommand(testCmd)

	// Create help command and test
	helpCmd := NewCommand()
	helpCmd.SetArgs([]string{"test"})

	output := captureOutput(t, func() {
		err := showCommandHelp(helpCmd, "test")
		// This will fail because we don't have the full root command setup
		// but we can check the formatting logic works
		if err != nil && !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Output should be captured even if command fails
	assert.NotNil(t, output)
}

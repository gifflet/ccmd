package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/cmd/remove"
	"github.com/gifflet/ccmd/internal/output"
)

// Build information, injected at build time
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ccmd",
	Short: "A CLI tool for managing Claude Code commands",
	Long: `ccmd is a command-line interface tool designed to help manage and execute
Claude Code commands efficiently.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		_ = cmd.Help()
	},
}

func init() {
	// Register subcommands
	rootCmd.AddCommand(remove.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.Fatal("Command failed: %v", err)
	}
}

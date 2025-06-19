package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/cmd/help"
	"github.com/gifflet/ccmd/cmd/info"
	"github.com/gifflet/ccmd/cmd/install"
	"github.com/gifflet/ccmd/cmd/list"
	"github.com/gifflet/ccmd/cmd/remove"
	"github.com/gifflet/ccmd/cmd/search"
	"github.com/gifflet/ccmd/cmd/sync"
	"github.com/gifflet/ccmd/cmd/update"
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
Claude Code commands efficiently.

ccmd provides a simple way to install, manage, and synchronize command-line tools
from GitHub repositories. It uses a declarative approach with ccmd.yaml files to
manage project dependencies and ensure consistent tool versions across teams.

Key Features:
  • Install commands from GitHub repositories
  • Declarative dependency management with ccmd.yaml
  • Sync installed commands with project requirements
  • Update commands to latest or specific versions
  • Search for available commands
  • List and inspect installed commands

Getting Started:
  1. Create a ccmd.yaml file in your project root
  2. Declare your command dependencies
  3. Run 'ccmd sync' to install all dependencies

For detailed help on any command, use 'ccmd help [command]' or 'ccmd [command] --help'.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		if err := cmd.Help(); err != nil {
			// Help command failed, but we can ignore this
			_ = err
		}
	},
}

func main() {
	// Register subcommands
	rootCmd.AddCommand(help.NewCommand())
	rootCmd.AddCommand(info.NewCommand())
	rootCmd.AddCommand(install.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(remove.NewCommand())
	rootCmd.AddCommand(search.NewCommand())
	rootCmd.AddCommand(sync.NewCommand())
	rootCmd.AddCommand(update.NewCommand())

	// Configure help command
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
	rootCmd.InitDefaultHelpCmd()

	if err := rootCmd.Execute(); err != nil {
		output.Fatalf("Command failed: %v", err)
	}
}

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/cmd/info"
	cmdinit "github.com/gifflet/ccmd/cmd/init"
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
Claude Code commands efficiently.`,
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
	rootCmd.AddCommand(info.NewCommand())
	rootCmd.AddCommand(cmdinit.NewCommand())
	rootCmd.AddCommand(install.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(remove.NewCommand())
	rootCmd.AddCommand(search.NewCommand())
	rootCmd.AddCommand(sync.NewCommand())
	rootCmd.AddCommand(update.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		output.Fatalf("Command failed: %v", err)
	}
}

package main

import (
	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/cmd/install"
	"github.com/gifflet/ccmd/cmd/list"
	"github.com/gifflet/ccmd/cmd/remove"
	"github.com/gifflet/ccmd/cmd/search"
	"github.com/gifflet/ccmd/internal/output"
)

var rootCmd = &cobra.Command{
	Use:   "ccmd",
	Short: "A CLI tool for managing Claude Code commands",
	Long: `ccmd is a command-line interface tool designed to help manage and execute
Claude Code commands efficiently.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		_ = cmd.Help()
	},
}

func init() {
	// Register subcommands
	rootCmd.AddCommand(install.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(remove.NewCommand())
	rootCmd.AddCommand(search.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.Fatal("Command failed: %v", err)
	}
}

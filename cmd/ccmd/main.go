package main

import (
	"github.com/gifflet/ccmd/internal/output"
	"github.com/spf13/cobra"
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.Fatal("Command failed: %v", err)
	}
}

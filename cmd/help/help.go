package help

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/errors"
)

// NewCommand creates a new help command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Show help for ccmd or a specific command",
		Long: `Show comprehensive help information for ccmd or a specific command.

When called without arguments, displays an overview of all available commands.
When called with a command name, displays detailed help for that specific command.

Examples:
  # Show general help
  ccmd help

  # Show help for the sync command
  ccmd help sync

  # Show help for the install command
  ccmd help install`,
		Args: cobra.MaximumNArgs(1),
		RunE: errors.WrapCommand("help", func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return showGeneralHelp(cmd)
			}
			return showCommandHelp(cmd, args[0])
		}),
	}

	return cmd
}

// showGeneralHelp displays general help information about ccmd.
func showGeneralHelp(cmd *cobra.Command) error {
	rootCmd := cmd.Root()

	output.Printf("%s", rootCmd.Long)
	output.Printf("")
	output.PrintInfof("Usage:")
	output.Printf("  %s [command] [flags]", rootCmd.Use)
	output.Printf("  %s [command] --help", rootCmd.Use)
	output.Printf("")

	output.PrintInfof("Available Commands:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, subCmd := range rootCmd.Commands() {
		if !subCmd.Hidden {
			fmt.Fprintf(w, "  %s\t%s\n", subCmd.Name(), subCmd.Short)
		}
	}
	w.Flush()
	output.Printf("")

	output.PrintInfof("Global Flags:")
	output.Printf("  -h, --help      Show help for any command")
	output.Printf("  -v, --version   Show version information")
	output.Printf("")

	output.PrintInfof("Common Use Cases:")
	output.Printf("")
	output.Printf("  Initialize a new project:")
	output.Printf("    $ ccmd init")
	output.Printf("    $ echo 'commands: []' > ccmd.yaml")
	output.Printf("")
	output.Printf("  Install a command:")
	output.Printf("    $ ccmd install github.com/user/repo")
	output.Printf("    $ ccmd install github.com/user/repo@v1.0.0")
	output.Printf("")
	output.Printf("  Sync commands with ccmd.yaml:")
	output.Printf("    $ ccmd sync")
	output.Printf("    $ ccmd sync --dry-run  # Preview changes")
	output.Printf("")
	output.Printf("  List installed commands:")
	output.Printf("    $ ccmd list")
	output.Printf("    $ ccmd list --json  # JSON output")
	output.Printf("")
	output.Printf("  Search for commands:")
	output.Printf("    $ ccmd search query")
	output.Printf("    $ ccmd search -t tag  # Search by tag")
	output.Printf("")

	output.PrintInfof("For more information about a specific command:")
	output.Printf("  ccmd help [command]")
	output.Printf("  ccmd [command] --help")

	return nil
}

// showCommandHelp displays detailed help for a specific command.
func showCommandHelp(cmd *cobra.Command, commandName string) error {
	rootCmd := cmd.Root()

	// Find the requested command
	targetCmd, _, err := rootCmd.Find([]string{commandName})
	if err != nil || targetCmd == rootCmd {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	// Print command-specific help with enhanced formatting
	output.PrintInfof("Command: %s", targetCmd.Name())
	output.Printf("")

	if targetCmd.Long != "" {
		output.Printf("%s", targetCmd.Long)
	} else if targetCmd.Short != "" {
		output.Printf("%s", targetCmd.Short)
	}
	output.Printf("")

	// Usage
	if targetCmd.Use != "" {
		output.PrintInfof("Usage:")
		output.Printf("  ccmd %s", targetCmd.Use)
		output.Printf("")
	}

	// Aliases
	if len(targetCmd.Aliases) > 0 {
		output.PrintInfof("Aliases:")
		output.Printf("  %s", strings.Join(targetCmd.Aliases, ", "))
		output.Printf("")
	}

	// Examples from Long description
	if strings.Contains(targetCmd.Long, "Examples:") || strings.Contains(targetCmd.Long, "Example:") {
		// Examples are already included in Long description
		// Just ensure proper formatting
	} else {
		// Add command-specific examples if not in Long description
		printCommandExamples(targetCmd.Name())
	}

	// Flags
	if targetCmd.HasAvailableLocalFlags() {
		output.PrintInfof("Flags:")
		targetCmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			flagInfo := fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
			if flag.Value.Type() != "bool" {
				flagInfo += fmt.Sprintf(" %s", flag.Value.Type())
			}

			// Add padding for alignment
			padding := 25 - len(flagInfo)
			if padding < 1 {
				padding = 1
			}

			fmt.Printf("%s%s%s\n", flagInfo, strings.Repeat(" ", padding), flag.Usage)

			if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "[]" {
				fmt.Printf("%sDefault: %s\n", strings.Repeat(" ", 27), flag.DefValue)
			}
		})
		output.Printf("")
	}

	// Global flags reminder
	output.PrintInfof("Global Flags:")
	output.Printf("  -h, --help      Show help for this command")
	output.Printf("")

	// Related commands
	printRelatedCommands(targetCmd.Name())

	return nil
}

// printCommandExamples prints examples for commands that don't have them in Long description.
func printCommandExamples(commandName string) {
	examples := getCommandExamples(commandName)
	if len(examples) > 0 {
		output.Printf("")
		output.PrintInfof("Examples:")
		for _, example := range examples {
			output.Printf("  %s", example)
		}
		output.Printf("")
	}
}

// getCommandExamples returns examples for a specific command.
func getCommandExamples(commandName string) []string {
	examplesMap := map[string][]string{
		"list": {
			"# List all installed commands",
			"ccmd list",
			"",
			"# List commands with details",
			"ccmd list --long",
			"",
			"# List commands in JSON format",
			"ccmd list --json",
			"",
			"# List and sort by a specific field",
			"ccmd list --sort name",
			"ccmd list --sort installed",
		},
		"update": {
			"# Update all commands",
			"ccmd update --all",
			"",
			"# Update a specific command",
			"ccmd update command-name",
			"",
			"# Update to a specific version",
			"ccmd update command-name --version v2.0.0",
			"",
			"# Force update even if up to date",
			"ccmd update command-name --force",
		},
		"search": {
			"# Search for commands by keyword",
			"ccmd search cli-tool",
			"",
			"# Search by tag",
			"ccmd search -t productivity",
			"",
			"# Search by author",
			"ccmd search -a username",
			"",
			"# Search with multiple filters",
			"ccmd search query -t golang -a author",
		},
		"info": {
			"# Show info about an installed command",
			"ccmd info command-name",
			"",
			"# Show brief info",
			"ccmd info command-name --brief",
			"",
			"# Show info in JSON format",
			"ccmd info command-name --json",
		},
		"remove": {
			"# Remove a command",
			"ccmd remove command-name",
			"",
			"# Remove a command without confirmation",
			"ccmd remove command-name --force",
			"",
			"# Remove multiple commands",
			"ccmd remove cmd1 cmd2 cmd3",
		},
	}

	return examplesMap[commandName]
}

// printRelatedCommands suggests related commands based on the current command.
func printRelatedCommands(commandName string) {
	relatedMap := map[string][]string{
		"install": {"list", "sync", "remove"},
		"remove":  {"list", "install"},
		"list":    {"info", "install", "remove"},
		"sync":    {"list", "install", "update"},
		"update":  {"list", "info", "sync"},
		"search":  {"install", "info"},
		"info":    {"list", "update", "remove"},
	}

	if related, exists := relatedMap[commandName]; exists && len(related) > 0 {
		output.PrintInfof("See also:")
		output.Printf("  ccmd help %s", strings.Join(related, ", ccmd help "))
		output.Printf("")
	}
}

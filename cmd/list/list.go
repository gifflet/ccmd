package list

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
)

// NewCommand creates a new list command.
func NewCommand() *cobra.Command {
	var (
		verbose bool
		jsonOut bool
		sortBy  string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all installed commands",
		Long: `List all installed commands with their versions, sources, and last update dates.

The list command provides a quick overview of all ccmd-managed commands currently
installed on your system. Use various flags to customize the output format.

Examples:
  # List all commands in table format
  ccmd list

  # Show detailed information for each command
  ccmd list --verbose

  # Output in JSON format
  ccmd list --json

  # Sort by different fields
  ccmd list --sort name
  ccmd list --sort version
  ccmd list --sort updated`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(verbose, jsonOut, sortBy)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().BoolVar(&verbose, "long", false, "Alias for --verbose")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by field: name, version, updated, installed")

	return cmd
}

func runList(verbose, jsonOut bool, sortBy string) error {
	// Get detailed command information
	opts := commands.ListOptions{}
	details, err := commands.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(details) == 0 {
		if jsonOut {
			output.Printf("[]")
		} else {
			output.PrintInfof("No commands installed yet.")
			output.PrintInfof("Use 'ccmd install' to install commands.")
		}
		return nil
	}

	// Sort based on sortBy parameter
	switch sortBy {
	case "name":
		sort.Slice(details, func(i, j int) bool {
			return details[i].Name < details[j].Name
		})
	case "version":
		sort.Slice(details, func(i, j int) bool {
			return details[i].Version < details[j].Version
		})
	case "updated":
		sort.Slice(details, func(i, j int) bool {
			return details[i].UpdatedAt.After(details[j].UpdatedAt)
		})
	case "installed":
		sort.Slice(details, func(i, j int) bool {
			return details[i].InstalledAt.After(details[j].InstalledAt)
		})
	default:
		// Default to name
		sort.Slice(details, func(i, j int) bool {
			return details[i].Name < details[j].Name
		})
	}

	// Check for structure issues
	hasStructureIssues := false
	for _, detail := range details {
		if !detail.StructureValid {
			hasStructureIssues = true
			break
		}
	}

	// Output based on format
	if jsonOut {
		return output.PrintJSON(details)
	}

	// Print table
	if verbose {
		printVerboseList(details)
	} else {
		printSimpleList(details)
	}

	// Show warning if there are structure issues
	if hasStructureIssues {
		output.PrintWarningf("\nSome commands have broken dual structure (missing directory or .md file).")
		output.PrintWarningf("Run with --verbose flag to see details.")
	}

	return nil
}

func printSimpleList(commands []*commands.CommandDetail) {
	output.PrintInfof("Found %d installed command(s):\n", len(commands))

	// Define column widths
	const (
		nameWidth    = 20
		versionWidth = 10
		sourceWidth  = 30
		updatedWidth = 20
	)

	// Print header - Bold adds ANSI codes, so we need to pad the content, not the formatted string
	fmt.Printf("%s%s  %s%s  %s%s  %s\n",
		output.Bold("NAME"), strings.Repeat(" ", nameWidth-4),
		output.Bold("VERSION"), strings.Repeat(" ", versionWidth-7),
		output.Bold("SOURCE"), strings.Repeat(" ", sourceWidth-6),
		output.Bold("UPDATED"))

	// Print separator line
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", versionWidth),
		strings.Repeat("-", sourceWidth),
		strings.Repeat("-", updatedWidth))

	// Print commands
	for _, detail := range commands {
		name := detail.Name
		if !detail.StructureValid {
			name += output.Warning(" ⚠")
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			nameWidth, name,
			versionWidth, detail.Version,
			sourceWidth, truncateSource(detail.Source, sourceWidth),
			updatedWidth, formatTime(detail.UpdatedAt))
	}
}

func printVerboseList(commands []*commands.CommandDetail) {
	output.PrintInfof("Found %d installed command(s):\n", len(commands))

	for i, detail := range commands {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("%s %s\n", output.Bold("Command:"), detail.Name)
		fmt.Printf("  Version:     %s\n", detail.Version)
		fmt.Printf("  Source:      %s\n", detail.Source)
		fmt.Printf("  Installed:   %s\n", formatTimeVerbose(detail.InstalledAt))
		fmt.Printf("  Updated:     %s\n", formatTimeVerbose(detail.UpdatedAt))

		// Structure status
		fmt.Print("  Structure:   ")
		if detail.StructureValid {
			fmt.Printf("%s\n", output.Success("OK"))
		} else {
			fmt.Printf("%s", output.Error("BROKEN"))
			issues := []string{}
			if !detail.HasDirectory {
				issues = append(issues, "missing directory")
			}
			if !detail.HasMarkdownFile {
				issues = append(issues, "missing .md file")
			}
			fmt.Printf(" (%s)\n", strings.Join(issues, ", "))
		}

		// Dependencies
		if len(detail.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %s\n", strings.Join(detail.Dependencies, ", "))
		}

		// Metadata
		if len(detail.Metadata) > 0 {
			fmt.Println("  Metadata:")
			keys := make([]string, 0, len(detail.Metadata))
			for k := range detail.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("    %s: %s\n", k, detail.Metadata[k])
			}
		}
	}
}

func truncateSource(source string, maxLen int) string {
	if len(source) <= maxLen {
		return source
	}
	return source[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

func formatTimeVerbose(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

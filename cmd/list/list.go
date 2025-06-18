package list

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
)

// NewCommand creates a new list command.
func NewCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed commands",
		Long:  `List all installed commands with their versions, sources, and last update dates.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	return cmd
}

func runList(verbose bool) error {
	// Get detailed command information
	opts := commands.ListOptions{}
	details, err := commands.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(details) == 0 {
		output.PrintInfo("No commands installed yet.")
		output.PrintInfo("Use 'ccmd install' to install commands.")
		return nil
	}

	// Sort by name
	sort.Slice(details, func(i, j int) bool {
		return details[i].Name < details[j].Name
	})

	// Check for structure issues
	hasStructureIssues := false
	for _, detail := range details {
		if !detail.StructureValid {
			hasStructureIssues = true
			break
		}
	}

	// Print table
	if verbose {
		printVerboseList(details)
	} else {
		printSimpleList(details)
	}

	// Show warning if there are structure issues
	if hasStructureIssues {
		output.PrintWarning("\nSome commands have broken dual structure (missing directory or .md file).")
		output.PrintWarning("Run with --verbose flag to see details.")
	}

	return nil
}

func printSimpleList(commands []*commands.CommandDetail) {
	output.PrintInfo(fmt.Sprintf("Found %d installed command(s):\n", len(commands)))

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		output.Bold("NAME"),
		output.Bold("VERSION"),
		output.Bold("SOURCE"),
		output.Bold("UPDATED"))
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 30),
		strings.Repeat("-", 20))

	// Print commands
	for _, detail := range commands {
		status := ""
		if !detail.StructureValid {
			status = output.Warning(" ⚠")
		}

		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n",
			detail.Name,
			status,
			detail.Version,
			truncateSource(detail.Source, 30),
			formatTime(detail.UpdatedAt))
	}
}

func printVerboseList(commands []*commands.CommandDetail) {
	output.PrintInfo(fmt.Sprintf("Found %d installed command(s):\n", len(commands)))

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

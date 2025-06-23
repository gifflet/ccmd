/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

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
	var long bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all commands managed by ccmd",
		Long: `List all commands managed by ccmd with their versions, sources, and metadata.

This command shows only commands that are tracked in the ccmd-lock.yaml file
and have entries in the .claude/commands/ directory.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(long)
		},
	}

	cmd.Flags().BoolVarP(&long, "long", "l", false, "Show detailed output including metadata")

	return cmd
}

func runList(long bool) error {
	// Get detailed command information
	opts := commands.ListOptions{}
	details, err := commands.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(details) == 0 {
		output.PrintInfof("No commands installed yet.")
		output.PrintInfof("Use 'ccmd install' to install commands.")
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
	if long {
		printLongList(details)
	} else {
		printSimpleList(details)
	}

	// Show warning if there are structure issues
	if hasStructureIssues {
		output.PrintWarningf("\nSome commands have broken dual structure (missing directory or .md file).")
		output.PrintWarningf("Run with --long flag to see details.")
	}

	return nil
}

func printSimpleList(commands []*commands.CommandDetail) {
	output.PrintInfof("Found %d command(s) managed by ccmd:\n", len(commands))

	// Define column widths
	const (
		nameWidth        = 20
		versionWidth     = 10
		descriptionWidth = 40
		updatedWidth     = 20
	)

	// Print header - Bold adds ANSI codes, so we need to pad the content, not the formatted string
	fmt.Printf("%s%s  %s%s  %s%s  %s\n",
		output.Bold("NAME"), strings.Repeat(" ", nameWidth-4),
		output.Bold("VERSION"), strings.Repeat(" ", versionWidth-7),
		output.Bold("DESCRIPTION"), strings.Repeat(" ", descriptionWidth-11),
		output.Bold("UPDATED"))

	// Print separator line
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", versionWidth),
		strings.Repeat("-", descriptionWidth),
		strings.Repeat("-", updatedWidth))

	// Print commands
	for _, detail := range commands {
		name := detail.Name
		if !detail.StructureValid {
			name += output.Warning(" ⚠")
		}

		// Use metadata version if available, otherwise use lock file version
		version := detail.Version
		if detail.CommandMetadata != nil && detail.CommandMetadata.Version != "" {
			version = detail.CommandMetadata.Version
		}

		// Get description from metadata if available
		description := "(no description)"
		if detail.CommandMetadata != nil && detail.CommandMetadata.Description != "" {
			description = detail.CommandMetadata.Description
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			nameWidth, name,
			versionWidth, version,
			descriptionWidth, truncateText(description, descriptionWidth),
			updatedWidth, formatTime(detail.UpdatedAt))
	}
}

func printLongList(commands []*commands.CommandDetail) {
	output.PrintInfof("Found %d command(s) managed by ccmd:\n", len(commands))

	for i, detail := range commands {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("%s %s\n", output.Bold("Command:"), detail.Name)

		// Use metadata version if available
		version := detail.Version
		if detail.CommandMetadata != nil && detail.CommandMetadata.Version != "" {
			version = detail.CommandMetadata.Version
		}
		fmt.Printf("  Version:     %s\n", version)

		// Show description from metadata
		if detail.CommandMetadata != nil && detail.CommandMetadata.Description != "" {
			fmt.Printf("  Description: %s\n", detail.CommandMetadata.Description)
		}

		// Show author from metadata
		if detail.CommandMetadata != nil && detail.CommandMetadata.Author != "" {
			fmt.Printf("  Author:      %s\n", detail.CommandMetadata.Author)
		}

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

		// Show metadata details if available
		if detail.CommandMetadata != nil {
			// Entry point
			if detail.CommandMetadata.Entry != "" {
				fmt.Printf("  Entry:       %s\n", detail.CommandMetadata.Entry)
			}

			// Tags
			if len(detail.CommandMetadata.Tags) > 0 {
				fmt.Printf("  Tags:        %s\n", strings.Join(detail.CommandMetadata.Tags, ", "))
			}

			// License
			if detail.CommandMetadata.License != "" {
				fmt.Printf("  License:     %s\n", detail.CommandMetadata.License)
			}

			// Homepage
			if detail.CommandMetadata.Homepage != "" {
				fmt.Printf("  Homepage:    %s\n", detail.CommandMetadata.Homepage)
			}

			// Repository (from metadata)
			if detail.CommandMetadata.Repository != "" && detail.CommandMetadata.Repository != detail.Source {
				fmt.Printf("  Repository:  %s\n", detail.CommandMetadata.Repository)
			}
		}

		// Dependencies from lock file
		if len(detail.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %s\n", strings.Join(detail.Dependencies, ", "))
		}

		// Additional metadata from lock file
		if len(detail.Metadata) > 0 {
			fmt.Println("  Lock Metadata:")
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

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
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

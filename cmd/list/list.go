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
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
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
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Get detailed command information
	opts := core.ListOptions{
		ProjectPath: cwd,
	}
	details, err := core.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(details) == 0 {
		output.PrintInfof("No commands installed yet.")
		output.PrintInfof("Use 'ccmd install' to install commands.")
		return nil
	}

	// Check for structure issues
	hasStructureIssues := false
	for _, detail := range details {
		if detail.BrokenStructure {
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

func printSimpleList(commands []core.CommandDetail) {
	output.PrintInfof("Found %d command(s) managed by ccmd:\n", len(commands))

	// Define column widths
	const (
		nameWidth        = 20
		versionWidth     = 10
		descriptionWidth = 40
		updatedWidth     = 20
	)

	// Print header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s",
		nameWidth, "NAME",
		versionWidth, "VERSION",
		descriptionWidth, "DESCRIPTION",
		updatedWidth, "UPDATED")
	output.Printf(header)
	output.Printf(strings.Repeat("-", len(header)))

	// Print each command
	for _, cmd := range commands {
		// Format name with warning icon if structure is broken
		name := cmd.Name
		if cmd.BrokenStructure {
			name = "⚠ " + name
		}
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		// Format version
		version := cmd.Version
		if version == "" {
			version = "unknown"
		}
		if len(version) > versionWidth {
			version = version[:versionWidth-3] + "..."
		}

		// Format description
		description := cmd.Description
		if description == "" {
			description = "-"
		}
		if len(description) > descriptionWidth {
			description = description[:descriptionWidth-3] + "..."
		}

		// Format updated time
		updated := formatTimeAgo(cmd.UpdatedAt)
		if len(updated) > updatedWidth {
			updated = updated[:updatedWidth-3] + "..."
		}

		// Print row
		row := fmt.Sprintf("%-*s %-*s %-*s %-*s",
			nameWidth, name,
			versionWidth, version,
			descriptionWidth, description,
			updatedWidth, updated)
		output.Printf(row)
	}
}

func printLongList(commands []core.CommandDetail) {
	output.PrintInfof("Found %d command(s) managed by ccmd:\n", len(commands))

	for i, cmd := range commands {
		if i > 0 {
			output.Printf(strings.Repeat("-", 60))
		}

		// Basic info
		output.Printf("Name:        %s", cmd.Name)
		output.Printf("Version:     %s", formatOrDash(cmd.Version))
		output.Printf("Source:      %s", formatOrDash(cmd.Repository))
		output.Printf("Description: %s", formatOrDash(cmd.Description))

		// Metadata
		if cmd.Author != "" {
			output.Printf("Author:      %s", cmd.Author)
		}
		if len(cmd.Tags) > 0 {
			output.Printf("Tags:        %s", strings.Join(cmd.Tags, ", "))
		}
		if cmd.License != "" {
			output.Printf("License:     %s", cmd.License)
		}
		if cmd.Homepage != "" {
			output.Printf("Homepage:    %s", cmd.Homepage)
		}

		// Structure status
		if cmd.BrokenStructure {
			output.Printf("Status:      ⚠ BROKEN - %s", cmd.StructureError)
		} else {
			output.Printf("Status:      OK")
		}

		// Timestamps
		output.Printf("Installed:   %s", formatTimestamp(cmd.InstalledAt))
		output.Printf("Updated:     %s", formatTimestamp(cmd.UpdatedAt))
	}
}

func formatOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func formatTimeAgo(timestamp string) string {
	if timestamp == "" {
		return "unknown"
	}

	// Parse timestamp (assuming RFC3339 format)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try other formats
		t, err = time.Parse("2006-01-02T15:04:05Z", timestamp)
		if err != nil {
			return timestamp
		}
	}

	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 30*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

func formatTimestamp(timestamp string) string {
	if timestamp == "" {
		return "unknown"
	}

	// Parse timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try other formats
		t, err = time.Parse("2006-01-02T15:04:05Z", timestamp)
		if err != nil {
			return timestamp
		}
	}

	return t.Format("2006-01-02 15:04:05")
}

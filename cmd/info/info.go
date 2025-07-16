/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package info

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/output"
)

// NewCommand creates a new info command.
func NewCommand() *cobra.Command {
	var jsonFormat bool

	cmd := &cobra.Command{
		Use:   "info <command-name>",
		Short: "Display detailed information about an installed command",
		Long: `Display detailed information about a specific installed command,
including metadata and structure verification.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(args[0], jsonFormat)
		},
	}

	cmd.Flags().BoolVar(&jsonFormat, "json", false, "Output in JSON format")

	return cmd
}

func runInfo(commandName string, jsonFormat bool) error {
	return runInfoWithFS(commandName, jsonFormat, nil)
}

func runInfoWithFS(commandName string, jsonFormat bool, filesystem fs.FileSystem) error {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	// Get detailed command information from core
	info, err := core.GetCommandDetails(commandName, ".", filesystem)
	if err != nil {
		if jsonFormat {
			return err
		}
		output.PrintErrorf("Error: %s", err.Error())
		return fmt.Errorf("command not found")
	}

	if jsonFormat {
		// Output as JSON
		jsonStr, err := core.FormatCommandInfoJSON(info)
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	} else {
		// Output as formatted text
		displayCommandInfo(info, commandName, filesystem)
	}

	return nil
}

func displayCommandInfo(info *core.CommandInfo, commandName string, filesystem fs.FileSystem) {
	// Header
	fmt.Println()
	output.PrintInfof("=== Command Information ===")
	fmt.Println()

	// Basic info
	fmt.Printf("%s %s\n", color.CyanString("Name:"), info.Name)
	fmt.Printf("%s %s\n", color.CyanString("Version:"), info.Version)

	if info.Author != "" {
		fmt.Printf("%s %s\n", color.CyanString("Author:"), info.Author)
	}

	if info.Description != "" {
		fmt.Printf("%s %s\n", color.CyanString("Description:"), info.Description)
	}

	if info.Repository != "" {
		fmt.Printf("%s %s\n", color.CyanString("Repository:"), info.Repository)
	}

	if info.Homepage != "" {
		fmt.Printf("%s %s\n", color.CyanString("Homepage:"), info.Homepage)
	}

	if info.License != "" {
		fmt.Printf("%s %s\n", color.CyanString("License:"), info.License)
	}

	if len(info.Tags) > 0 {
		fmt.Printf("%s %s\n", color.CyanString("Tags:"), strings.Join(info.Tags, ", "))
	}

	if info.Entry != "" {
		fmt.Printf("%s %s\n", color.CyanString("Entry Point:"), info.Entry)
	}

	// Installation info
	fmt.Println()
	output.PrintInfof("=== Installation Details ===")
	fmt.Println()

	fmt.Printf("%s %s\n", color.CyanString("Source:"), info.Source)
	fmt.Printf("%s %s\n", color.CyanString("Installed:"), info.InstalledAt)
	fmt.Printf("%s %s\n", color.CyanString("Updated:"), info.UpdatedAt)

	// Structure verification
	fmt.Println()
	output.PrintInfof("=== Structure Verification ===")
	fmt.Println()

	printStatus("Command directory", info.Structure.DirectoryExists)
	printStatus("Standalone .md file", info.Structure.MarkdownExists)
	printStatus("ccmd.yaml", info.Structure.HasCcmdYaml)
	printStatus("index.md", info.Structure.HasIndexMd)

	if len(info.Structure.Issues) > 0 {
		fmt.Println()
		output.PrintWarningf("Issues found:")
		for _, issue := range info.Structure.Issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	// Preview content if available
	if info.Structure.HasIndexMd {
		fmt.Println()
		output.PrintInfof("=== Content Preview ===")
		fmt.Println()

		preview, totalLines, err := core.ReadCommandContentPreview(commandName, ".claude", filesystem, 10)
		if err == nil {
			fmt.Print(preview)
			if totalLines > 10 {
				fmt.Printf("\n... (showing first 10 lines of %d total)\n", totalLines)
			}
		}
	}

	fmt.Println()
}

func printStatus(label string, ok bool) {
	status := color.GreenString("✓")
	if !ok {
		status = color.RedString("✗")
	}
	fmt.Printf("  %s %s\n", status, label)
}

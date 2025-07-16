/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package init

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
)

// NewCommand creates a new init command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Claude Code Command project",
		Long: `Initialize a new Claude Code Command project by creating the necessary 
configuration files and directory structure.

This interactive command guides you through setting up a new ccmd project. It will
prompt you for essential metadata about your command, including name, version,
description, author, and repository information. The command then generates a
properly formatted ccmd.yaml file with your specifications.

Additionally, it creates the .claude/commands directory structure required for
storing and managing Claude Code commands in your project.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}

	return cmd
}

func runInit() error {
	scanner := bufio.NewScanner(os.Stdin)

	output.Printf("This utility will walk you through creating a ccmd.yaml file.")
	output.Printf("Press ^C at any time to quit.\n")

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load existing config if any
	defaults, existingCommands, err := core.LoadExistingConfig(currentDir)
	if err != nil {
		output.PrintWarningf("Warning: %v", err)
		// Continue with fresh defaults
		defaults = core.InitDefaults(currentDir)
	} else if existingCommands != nil {
		output.Printf("Loaded existing ccmd.yaml file.")
	}

	// Prompt for each field
	name := promptUser(scanner, "name", defaults.Name)
	version := promptUser(scanner, "version", defaults.Version)
	description := promptUser(scanner, "description", defaults.Description)
	author := promptUser(scanner, "author", defaults.Author)
	repository := promptUser(scanner, "repository", defaults.Repository)
	entry := promptUser(scanner, "entry", defaults.Entry)
	tagsInput := promptUser(scanner, "tags (comma-separated)", core.FormatTags(defaults.Tags))

	// Create options
	opts := core.InitOptions{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Repository:  repository,
		Entry:       entry,
		Tags:        core.ParseTags(tagsInput),
		ProjectPath: currentDir,
	}

	// Generate preview
	preview, err := core.GenerateConfigPreview(opts, existingCommands)
	if err != nil {
		return err
	}

	output.Printf("\nAbout to write to %s:\n", filepath.Join(currentDir, "ccmd.yaml"))
	output.Printf("%s", preview)

	// Confirm
	confirm := promptUser(scanner, "\nIs this OK?", "yes")
	if !isConfirmation(confirm) {
		output.PrintWarningf("Canceled.")
		return nil
	}

	// Initialize project
	if existingCommands != nil {
		err = core.InitProjectWithCommands(opts, existingCommands)
	} else {
		err = core.InitProject(opts)
	}

	if err != nil {
		return err
	}

	output.PrintSuccessf("✓ Created .claude/commands directory")
	output.PrintSuccessf("✓ Created ccmd.yaml")
	output.Printf("\n🎉 ccmd project initialized!")

	// Show next steps
	showNextSteps(opts)

	return nil
}

func showNextSteps(opts core.InitOptions) {
	// Check if entry file exists
	entryPath := filepath.Join(opts.ProjectPath, opts.Entry)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		output.Printf("\n📝 Create %s with your command instructions:", opts.Entry)
		output.Printf("```markdown")
		output.Printf("# %s", opts.Name)
		output.Printf("")
		output.Printf("You are an AI assistant. When invoked, you should...")
		output.Printf("```")
	}

	output.Printf("\n🚀 Publish your command:")
	output.Printf("  1. git add ccmd.yaml %s  # add ccmd-lock.yaml if you installed commands", opts.Entry)
	output.Printf("  2. git commit -m \"feat: add %s command\"", opts.Name)
	output.Printf("  3. git push origin main")

	output.Printf("\n✨ Then install with:")
	output.PrintInfof(core.GetInstallCommand(opts.Repository))
}

func promptUser(scanner *bufio.Scanner, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s: (%s) ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" && defaultValue != "" {
		return defaultValue
	}

	return input
}

func isConfirmation(input string) bool {
	lower := strings.ToLower(input)
	return lower == "yes" || lower == "y" || lower == ""
}

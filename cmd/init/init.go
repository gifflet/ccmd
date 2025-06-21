package init

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/output"
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

	// Get current directory name for default
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	dirName := filepath.Base(currentDir)

	// Prompt for each field
	name := promptUser(scanner, "name", dirName)
	version := promptUser(scanner, "version", "1.0.0")
	description := promptUser(scanner, "description", "")
	author := promptUser(scanner, "author", "")
	repository := promptUser(scanner, "repository", "")
	entry := promptUser(scanner, "entry", "index.md")
	tagsStr := promptUser(scanner, "tags (comma-separated)", "")

	// Parse tags
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	// Create metadata structure
	metadata := models.CommandMetadata{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Repository:  repository,
		Entry:       entry,
		Tags:        tags,
	}

	// Show preview
	yamlData, err := yaml.Marshal(&metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	output.Printf("\nAbout to write to %s:\n", filepath.Join(currentDir, "ccmd.yaml"))
	output.Printf("%s", string(yamlData))

	// Confirm
	confirm := promptUser(scanner, "\nIs this OK?", "yes")
	if !isConfirmation(confirm) {
		output.PrintWarningf("Canceled.")
		return nil
	}

	// Create .claude directory
	claudeDir := filepath.Join(currentDir, ".claude", "commands")
	if err := fs.CreateDir(claudeDir); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Write ccmd.yaml
	ccmdPath := filepath.Join(currentDir, "ccmd.yaml")
	if err := fs.WriteYAMLFile(ccmdPath, &metadata); err != nil {
		return fmt.Errorf("failed to write ccmd.yaml: %w", err)
	}

	output.PrintSuccessf("✓ Created .claude/commands directory")
	output.PrintSuccessf("✓ Created ccmd.yaml")

	output.Printf("\n🎉 ccmd project initialized!")

	// Create the entry file if it doesn't exist
	entryPath := filepath.Join(currentDir, metadata.Entry)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		output.Printf("\n📝 Create %s with your command instructions:", metadata.Entry)
		output.Printf("```markdown")
		output.Printf("# %s", metadata.Name)
		output.Printf("")
		output.Printf("You are an AI assistant. When invoked, you should...")
		output.Printf("```")
	}

	output.Printf("\n🚀 Publish your command:")
	output.Printf("  1. git add ccmd.yaml %s  # add ccmd-lock.yaml if you installed commands", metadata.Entry)
	output.Printf("  2. git commit -m \"feat: add %s command\"", metadata.Name)
	output.Printf("  3. git push origin main")

	// Use the actual repository if provided, otherwise show placeholder
	installCmd := "ccmd install "
	if metadata.Repository != "" {
		// Extract GitHub path from full URL
		if strings.HasPrefix(metadata.Repository, "https://github.com/") {
			repoPath := strings.TrimPrefix(metadata.Repository, "https://github.com/")
			repoPath = strings.TrimSuffix(repoPath, ".git")
			installCmd += repoPath
		} else {
			installCmd += metadata.Repository
		}
	} else {
		installCmd += "github.com/your-username/your-repo"
	}

	output.Printf("\n✨ Then install with:")
	output.PrintInfof(installCmd)

	return nil
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

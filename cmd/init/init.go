// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
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
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/project"
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

	// Set defaults
	nameDefault := dirName
	versionDefault := "1.0.0"
	descriptionDefault := ""
	authorDefault := ""
	repositoryDefault := ""
	entryDefault := "index.md"
	tagsDefault := ""

	// Try to load existing ccmd.yaml
	ccmdPath := filepath.Join(currentDir, "ccmd.yaml")
	var existingCommands interface{}

	if _, err := os.Stat(ccmdPath); err == nil {
		// File exists, try to load it using raw YAML to preserve structure
		data, err := os.ReadFile(ccmdPath) //nolint:gosec // ccmdPath is constructed from known safe values
		if err == nil {
			var rawConfig map[string]interface{}
			if err := yaml.Unmarshal(data, &rawConfig); err == nil {
				// Extract basic fields for defaults
				if name, ok := rawConfig["name"].(string); ok && name != "" {
					nameDefault = name
				}
				if version, ok := rawConfig["version"].(string); ok && version != "" {
					versionDefault = version
				}
				if desc, ok := rawConfig["description"].(string); ok {
					descriptionDefault = desc
				}
				if author, ok := rawConfig["author"].(string); ok {
					authorDefault = author
				}
				if repo, ok := rawConfig["repository"].(string); ok {
					repositoryDefault = repo
				}
				if entry, ok := rawConfig["entry"].(string); ok && entry != "" {
					entryDefault = entry
				}
				if tags, ok := rawConfig["tags"].([]interface{}); ok {
					var tagStrings []string
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok {
							tagStrings = append(tagStrings, tagStr)
						}
					}
					if len(tagStrings) > 0 {
						tagsDefault = strings.Join(tagStrings, ", ")
					}
				}
				// Preserve commands field as-is
				existingCommands = rawConfig["commands"]
				output.Printf("Loaded existing ccmd.yaml file.")
			}
		}
	}

	// Prompt for each field
	name := promptUser(scanner, "name", nameDefault)
	version := promptUser(scanner, "version", versionDefault)
	description := promptUser(scanner, "description", descriptionDefault)
	author := promptUser(scanner, "author", authorDefault)
	repository := promptUser(scanner, "repository", repositoryDefault)
	entry := promptUser(scanner, "entry", entryDefault)
	tagsInput := promptUser(scanner, "tags (comma-separated)", tagsDefault)

	// Parse tags
	var tags []string
	if tagsInput != "" {
		for _, tag := range strings.Split(tagsInput, ",") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	// Create config structure
	config := project.Config{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Repository:  repository,
		Entry:       entry,
		Tags:        tags,
		Commands:    existingCommands, // Preserve existing commands
	}

	// Create a custom structure to ensure commands comes last
	type orderedConfig struct {
		Name        string      `yaml:"name,omitempty"`
		Version     string      `yaml:"version,omitempty"`
		Description string      `yaml:"description,omitempty"`
		Author      string      `yaml:"author,omitempty"`
		Repository  string      `yaml:"repository,omitempty"`
		Entry       string      `yaml:"entry,omitempty"`
		Tags        []string    `yaml:"tags,omitempty"`
		Commands    interface{} `yaml:"commands,omitempty"`
	}

	// Show preview
	preview := orderedConfig{
		Name:        config.Name,
		Version:     config.Version,
		Description: config.Description,
		Author:      config.Author,
		Repository:  config.Repository,
		Entry:       config.Entry,
		Tags:        config.Tags,
		Commands:    config.Commands,
	}

	yamlData, err := yaml.Marshal(&preview)
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

	// Write ccmd.yaml using the ordered structure
	if err := fs.WriteYAMLFile(ccmdPath, &preview); err != nil {
		return fmt.Errorf("failed to write ccmd.yaml: %w", err)
	}

	output.PrintSuccessf("✓ Created .claude/commands directory")
	output.PrintSuccessf("✓ Created ccmd.yaml")

	output.Printf("\n🎉 ccmd project initialized!")

	// Create the entry file if it doesn't exist
	entryPath := filepath.Join(currentDir, config.Entry)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		output.Printf("\n📝 Create %s with your command instructions:", config.Entry)
		output.Printf("```markdown")
		output.Printf("# %s", config.Name)
		output.Printf("")
		output.Printf("You are an AI assistant. When invoked, you should...")
		output.Printf("```")
	}

	output.Printf("\n🚀 Publish your command:")
	output.Printf("  1. git add ccmd.yaml %s  # add ccmd-lock.yaml if you installed commands", config.Entry)
	output.Printf("  2. git commit -m \"feat: add %s command\"", config.Name)
	output.Printf("  3. git push origin main")

	// Use the actual repository if provided, otherwise show placeholder
	installCmd := "ccmd install "
	if config.Repository != "" {
		// Extract GitHub path from full URL
		if strings.HasPrefix(config.Repository, "https://github.com/") {
			repoPath := strings.TrimPrefix(config.Repository, "https://github.com/")
			repoPath = strings.TrimSuffix(repoPath, ".git")
			installCmd += repoPath
		} else {
			installCmd += config.Repository
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

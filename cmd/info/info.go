package info

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
)

// Output represents the structured output format for JSON
type Output struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Repository  string            `json:"repository"`
	License     string            `json:"license,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Entry       string            `json:"entry,omitempty"`
	Source      string            `json:"source"`
	InstalledAt time.Time         `json:"installed_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Structure   StructureInfo     `json:"structure"`
}

// StructureInfo contains information about command structure integrity
type StructureInfo struct {
	DirectoryExists bool     `json:"directory_exists"`
	MarkdownExists  bool     `json:"markdown_exists"`
	HasCcmdYaml     bool     `json:"has_ccmd_yaml"`
	HasIndexMd      bool     `json:"has_index_md"`
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues,omitempty"`
}

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

	// Check if command exists
	exists, err := commands.CommandExists(commandName, "", filesystem)
	if err != nil {
		return fmt.Errorf("failed to check command existence: %w", err)
	}

	if !exists {
		if jsonFormat {
			return fmt.Errorf("command '%s' is not installed", commandName)
		}
		output.PrintErrorf("Command '%s' is not installed", commandName)
		return fmt.Errorf("command not found")
	}

	// Get command info from lock file
	cmdInfo, err := commands.GetCommandInfo(commandName, "", filesystem)
	if err != nil {
		return fmt.Errorf("failed to get command info: %w", err)
	}

	// Get base directory (project-local)
	baseDir := ".claude"

	// Check structure and get metadata
	structureInfo, metadata := checkCommandStructure(commandName, baseDir, filesystem)

	// Prepare output data
	infoData := Output{
		Name:        cmdInfo.Name,
		Version:     cmdInfo.Version,
		Source:      cmdInfo.Source,
		InstalledAt: cmdInfo.InstalledAt,
		UpdatedAt:   cmdInfo.UpdatedAt,
		Metadata:    cmdInfo.Metadata,
		Structure:   structureInfo,
	}

	// Fill in data from ccmd.yaml if available
	if metadata != nil {
		infoData.Author = metadata.Author
		infoData.Description = metadata.Description
		infoData.Repository = metadata.Repository
		infoData.License = metadata.License
		infoData.Homepage = metadata.Homepage
		infoData.Tags = metadata.Tags
		infoData.Entry = metadata.Entry
	} else if desc, ok := cmdInfo.Metadata["description"]; ok {
		// Fallback to lock file metadata
		infoData.Description = desc
	}

	if jsonFormat {
		// Output as JSON
		data, err := json.MarshalIndent(infoData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		// Output as formatted text
		displayCommandInfo(infoData, baseDir, commandName, filesystem)
	}

	return nil
}

func checkCommandStructure(commandName, baseDir string,
	filesystem fs.FileSystem) (StructureInfo, *models.CommandMetadata) {
	info := StructureInfo{
		DirectoryExists: false,
		MarkdownExists:  false,
		HasCcmdYaml:     false,
		HasIndexMd:      false,
		IsValid:         false,
		Issues:          []string{},
	}

	commandDir := filepath.Join(baseDir, "commands", commandName)
	markdownFile := filepath.Join(baseDir, "commands", commandName+".md")
	ccmdYamlFile := filepath.Join(commandDir, "ccmd.yaml")
	indexMdFile := filepath.Join(commandDir, "index.md")

	// Check directory
	if dirInfo, err := filesystem.Stat(commandDir); err == nil && dirInfo.IsDir() {
		info.DirectoryExists = true
	} else {
		info.Issues = append(info.Issues, "Command directory is missing")
	}

	// Check markdown file
	if fileInfo, err := filesystem.Stat(markdownFile); err == nil && !fileInfo.IsDir() {
		info.MarkdownExists = true
	} else {
		info.Issues = append(info.Issues, "Standalone markdown file is missing")
	}

	// Check ccmd.yaml
	var metadata *models.CommandMetadata
	if info.DirectoryExists {
		if data, err := filesystem.ReadFile(ccmdYamlFile); err == nil {
			info.HasCcmdYaml = true
			metadata = &models.CommandMetadata{}
			if err := yaml.Unmarshal(data, metadata); err != nil {
				info.Issues = append(info.Issues, "ccmd.yaml is malformed")
				metadata = nil
			}
		} else {
			info.Issues = append(info.Issues, "ccmd.yaml is missing")
		}

		// Check index.md
		if _, err := filesystem.Stat(indexMdFile); err == nil {
			info.HasIndexMd = true
		}
	}

	// Determine if structure is valid
	info.IsValid = info.DirectoryExists && info.MarkdownExists && info.HasCcmdYaml
	if !info.IsValid && len(info.Issues) == 0 {
		info.Issues = append(info.Issues, "Incomplete command structure")
	}

	return info, metadata
}

func displayCommandInfo(info Output, baseDir, commandName string, filesystem fs.FileSystem) {
	// Header
	fmt.Println()
	fmt.Println(output.Info("=== Command Information ==="))
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
	fmt.Println(output.Info("=== Installation Details ==="))
	fmt.Println()

	fmt.Printf("%s %s\n", color.CyanString("Source:"), info.Source)
	fmt.Printf("%s %s\n", color.CyanString("Installed:"), info.InstalledAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s %s\n", color.CyanString("Updated:"), info.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Structure verification
	fmt.Println()
	fmt.Println(output.Info("=== Structure Verification ==="))
	fmt.Println()

	printStatus("Command directory", info.Structure.DirectoryExists)
	printStatus("Standalone .md file", info.Structure.MarkdownExists)
	printStatus("ccmd.yaml", info.Structure.HasCcmdYaml)
	printStatus("index.md", info.Structure.HasIndexMd)

	if len(info.Structure.Issues) > 0 {
		fmt.Println()
		fmt.Println(output.Warning("Issues found:"))
		for _, issue := range info.Structure.Issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	// Preview content if available
	if info.Structure.HasIndexMd {
		fmt.Println()
		fmt.Println(output.Info("=== Content Preview ==="))
		fmt.Println()

		indexPath := filepath.Join(baseDir, "commands", commandName, "index.md")
		if content, err := filesystem.ReadFile(indexPath); err == nil {
			lines := strings.Split(string(content), "\n")
			previewLines := 10
			if len(lines) < previewLines {
				previewLines = len(lines)
			}

			for i := 0; i < previewLines; i++ {
				fmt.Println(lines[i])
			}

			if len(lines) > previewLines {
				fmt.Printf("\n... (showing first %d lines of %d total)\n", previewLines, len(lines))
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

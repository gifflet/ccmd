// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/project"
)

// ListOptions contains options for listing commands.
type ListOptions struct {
	BaseDir    string
	FileSystem fs.FileSystem
}

// CommandDetail contains detailed information about a command including structure validation.
type CommandDetail struct {
	*project.CommandLockInfo
	HasDirectory     bool
	HasMarkdownFile  bool
	StructureValid   bool
	StructureMessage string
	// Metadata from command's ccmd.yaml file
	CommandMetadata *models.CommandMetadata
}

// List returns detailed information about all installed commands.
func List(opts ListOptions) ([]*CommandDetail, error) {
	if opts.FileSystem == nil {
		opts.FileSystem = fs.OS{}
	}

	if opts.BaseDir == "" {
		opts.BaseDir = "."
	}

	// Load lock file from project root
	lockPath := filepath.Join(opts.BaseDir, "ccmd-lock.yaml")
	lockManager := project.NewLockManagerWithFS(lockPath, opts.FileSystem)
	if err := lockManager.Load(); err != nil {
		if os.IsNotExist(err) {
			return []*CommandDetail{}, nil
		}
		return nil, fmt.Errorf("failed to load lock file: %w", err)
	}

	// Get all commands from lock file
	commands, err := lockManager.ListCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}

	// Check structure for each command
	details := make([]*CommandDetail, 0, len(commands))
	for _, cmd := range commands {
		detail := &CommandDetail{
			CommandLockInfo: cmd,
		}

		// Check directory existence
		commandDir := filepath.Join(opts.BaseDir, ".claude", "commands", cmd.Name)
		if stat, err := opts.FileSystem.Stat(commandDir); err == nil && stat.IsDir() {
			detail.HasDirectory = true

			// Try to read command metadata from ccmd.yaml
			metadataPath := filepath.Join(commandDir, "ccmd.yaml")
			if data, err := opts.FileSystem.ReadFile(metadataPath); err == nil {
				var metadata models.CommandMetadata
				if err := metadata.UnmarshalYAML(data); err == nil {
					// Only use valid metadata
					if err := metadata.Validate(); err == nil {
						detail.CommandMetadata = &metadata
					}
				}
			}
		}

		// Check markdown file existence
		markdownFile := filepath.Join(opts.BaseDir, ".claude", "commands", cmd.Name+".md")
		if stat, err := opts.FileSystem.Stat(markdownFile); err == nil && !stat.IsDir() {
			detail.HasMarkdownFile = true
		}

		// Validate structure
		detail.StructureValid = detail.HasDirectory && detail.HasMarkdownFile
		if !detail.StructureValid {
			messages := []string{}
			if !detail.HasDirectory {
				messages = append(messages, "missing directory")
			}
			if !detail.HasMarkdownFile {
				messages = append(messages, "missing .md file")
			}
			detail.StructureMessage = fmt.Sprintf("broken structure: %v", messages)
		}

		details = append(details, detail)
	}

	return details, nil
}

// VerifyCommandStructure checks if a specific command has valid dual structure.
func VerifyCommandStructure(name, baseDir string, filesystem fs.FileSystem) (valid bool, status string, err error) {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	if baseDir == "" {
		baseDir = "."
	}

	// Check if command exists in lock file
	lockPath := filepath.Join(baseDir, "ccmd-lock.yaml")
	lockManager := project.NewLockManagerWithFS(lockPath, filesystem)
	if err := lockManager.Load(); err != nil {
		return false, "", fmt.Errorf("failed to load lock file: %w", err)
	}

	if !lockManager.HasCommand(name) {
		return false, "command not found in lock file", nil
	}

	// Check directory
	hasDir := false
	commandDir := filepath.Join(baseDir, ".claude", "commands", name)
	if stat, err := filesystem.Stat(commandDir); err == nil && stat.IsDir() {
		hasDir = true
	}

	// Check markdown file
	hasMarkdown := false
	markdownFile := filepath.Join(baseDir, ".claude", "commands", name+".md")
	if stat, err := filesystem.Stat(markdownFile); err == nil && !stat.IsDir() {
		hasMarkdown = true
	}

	if hasDir && hasMarkdown {
		return true, "", nil
	}

	// Build error message
	issues := []string{}
	if !hasDir {
		issues = append(issues, "missing directory")
	}
	if !hasMarkdown {
		issues = append(issues, "missing .md file")
	}

	return false, fmt.Sprintf("broken structure: %v", issues), nil
}

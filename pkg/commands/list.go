package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
)

// ListOptions contains options for listing commands.
type ListOptions struct {
	BaseDir    string
	FileSystem fs.FileSystem
}

// CommandDetail contains detailed information about a command including structure validation.
type CommandDetail struct {
	*models.Command
	HasDirectory     bool
	HasMarkdownFile  bool
	StructureValid   bool
	StructureMessage string
}

// List returns detailed information about all installed commands.
func List(opts ListOptions) ([]*CommandDetail, error) {
	if opts.FileSystem == nil {
		opts.FileSystem = fs.OS{}
	}

	if opts.BaseDir == "" {
		opts.BaseDir = ".claude"
	}

	// Load lock file
	lockManager := lock.NewManagerWithFS(opts.BaseDir, opts.FileSystem)
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
			Command: cmd,
		}

		// Check directory existence
		commandDir := filepath.Join(opts.BaseDir, "commands", cmd.Name)
		if stat, err := opts.FileSystem.Stat(commandDir); err == nil && stat.IsDir() {
			detail.HasDirectory = true
		}

		// Check markdown file existence
		markdownFile := filepath.Join(opts.BaseDir, "commands", cmd.Name+".md")
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
		baseDir = ".claude"
	}

	// Check if command exists in lock file
	lockManager := lock.NewManagerWithFS(baseDir, filesystem)
	if err := lockManager.Load(); err != nil {
		return false, "", fmt.Errorf("failed to load lock file: %w", err)
	}

	if !lockManager.HasCommand(name) {
		return false, "command not found in lock file", nil
	}

	// Check directory
	hasDir := false
	commandDir := filepath.Join(baseDir, "commands", name)
	if stat, err := filesystem.Stat(commandDir); err == nil && stat.IsDir() {
		hasDir = true
	}

	// Check markdown file
	hasMarkdown := false
	markdownFile := filepath.Join(baseDir, "commands", name+".md")
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

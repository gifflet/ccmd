package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
)

// RemoveOptions contains options for removing a command.
type RemoveOptions struct {
	Name       string
	BaseDir    string
	FileSystem fs.FileSystem
}

// Remove removes a command by name and cleans up associated files.
func Remove(opts RemoveOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("command name is required")
	}

	if opts.FileSystem == nil {
		opts.FileSystem = fs.OS{}
	}

	if opts.BaseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.BaseDir = filepath.Join(homeDir, ".claude")
	}

	lockManager := lock.NewManagerWithFS(opts.BaseDir, opts.FileSystem)

	// Load current lock file
	if err := lockManager.Load(); err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Check if command exists
	if !lockManager.HasCommand(opts.Name) {
		return fmt.Errorf("command '%s' not found", opts.Name)
	}

	// Remove command directory (ignore if doesn't exist)
	commandDir := filepath.Join(opts.BaseDir, "commands", opts.Name)
	if err := opts.FileSystem.RemoveAll(commandDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove command directory: %w", err)
	}

	// Remove command markdown file (ignore if doesn't exist)
	commandFile := filepath.Join(opts.BaseDir, "commands", opts.Name+".md")
	if err := opts.FileSystem.Remove(commandFile); err != nil && !os.IsNotExist(err) {
		// Try to restore the command directory if markdown removal fails
		if mkdirErr := opts.FileSystem.MkdirAll(commandDir, 0755); mkdirErr != nil {
			// Log error but don't fail the operation
			_ = mkdirErr
		}
		return fmt.Errorf("failed to remove command markdown file: %w", err)
	}

	// Update lock file
	if err := lockManager.RemoveCommand(opts.Name); err != nil {
		// Try to restore the command directory and file if lock file update fails
		if mkdirErr := opts.FileSystem.MkdirAll(commandDir, 0755); mkdirErr != nil {
			// Log error but don't fail the operation
			_ = mkdirErr
		}
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	if err := lockManager.Save(); err != nil {
		// Try to restore the command directory and file if save fails
		if mkdirErr := opts.FileSystem.MkdirAll(commandDir, 0755); mkdirErr != nil {
			// Log error but don't fail the operation
			_ = mkdirErr
		}
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	return nil
}

// ListCommands returns a list of all installed commands.
func ListCommands(baseDir string, filesystem fs.FileSystem) ([]string, error) {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".claude")
	}

	lockManager := lock.NewManagerWithFS(baseDir, filesystem)
	if err := lockManager.Load(); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to load lock file: %w", err)
	}

	cmds, err := lockManager.ListCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}

	commands := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		commands = append(commands, cmd.Name)
	}

	return commands, nil
}

// CommandExists checks if a command exists.
func CommandExists(name, baseDir string, filesystem fs.FileSystem) (bool, error) {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".claude")
	}

	lockManager := lock.NewManagerWithFS(baseDir, filesystem)
	if err := lockManager.Load(); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to load lock file: %w", err)
	}

	return lockManager.HasCommand(name), nil
}

// GetCommandInfo retrieves information about a specific command.
func GetCommandInfo(name, baseDir string, filesystem fs.FileSystem) (*models.Command, error) {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".claude")
	}

	lockManager := lock.NewManagerWithFS(baseDir, filesystem)
	if err := lockManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load lock file: %w", err)
	}

	cmd, err := lockManager.GetCommand(name)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

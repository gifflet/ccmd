// Package remove provides the remove command for ccmd.
package remove

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/project"
)

// NewCommand creates a new remove command.
func NewCommand() *cobra.Command {
	var (
		force bool
		save  bool
	)

	cmd := &cobra.Command{
		Use:   "remove <command-name>",
		Short: "Remove an installed command",
		Long:  `Remove an installed command and clean up all associated files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(args[0], force, save)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal without confirmation")
	cmd.Flags().BoolVarP(&save, "save", "s", false, "Update ccmd.yaml and ccmd-lock.yaml files")

	return cmd
}

func runRemove(commandName string, force, save bool) error {
	// Check if command exists in lock file
	lockManager := project.NewLockManagerWithFS("ccmd-lock.yaml", fs.OS{})
	if err := lockManager.Load(); err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	if !lockManager.HasCommand(commandName) {
		return fmt.Errorf("command '%s' is not installed", commandName)
	}

	// Get command info for display
	cmdInfo, err := commands.GetCommandInfo(commandName, "", nil)
	if err != nil {
		return fmt.Errorf("failed to get command info: %w", err)
	}

	// Confirm removal if not forced
	if !force {
		output.PrintInfof("Command details:")
		output.PrintInfof("  Name: %s", cmdInfo.Name)
		output.PrintInfof("  Version: %s", cmdInfo.Version)
		if desc, ok := cmdInfo.Metadata["description"]; ok && desc != "" {
			output.PrintInfof("  Description: %s", desc)
		}

		output.PrintWarningf("\nThis will permanently remove the command and all its files.")
		output.PrintInfof("Are you sure you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if !isConfirmation(response) {
			output.PrintInfof("Removal canceled")
			return nil
		}
	}

	// Create spinner for removal process
	spinner := output.NewSpinner(fmt.Sprintf("Removing command '%s'...", commandName))
	spinner.Start()

	// Remove the command
	removeOpts := commands.RemoveOptions{
		Name: commandName,
	}

	if err := commands.Remove(removeOpts); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to remove command: %w", err)
	}

	spinner.Stop()
	output.PrintSuccessf("Command '%s' has been successfully removed", commandName)

	// Update project files if --save flag is used
	if save {
		if err := updateProjectFiles(commandName, cmdInfo); err != nil {
			output.PrintWarningf("Command removed but failed to update project files: %v", err)
			// Don't return error as the command was already removed successfully
		} else {
			output.PrintSuccessf("Updated ccmd.yaml and ccmd-lock.yaml")
		}
	}

	return nil
}

func isConfirmation(response string) bool {
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func updateProjectFiles(commandName string, cmdInfo *project.CommandLockInfo) error {
	// Get the repository from command info
	repo := ""
	if cmdInfo != nil && cmdInfo.Metadata != nil {
		if r, ok := cmdInfo.Metadata["repository"]; ok {
			repo = r
		}
	}

	if repo == "" {
		return fmt.Errorf("cannot determine repository for command %s", commandName)
	}

	// Create project manager for current directory
	manager := project.NewManager(".")

	// Update ccmd.yaml if it exists
	if manager.ConfigExists() {
		// Remove the command from ccmd.yaml
		if err := manager.RemoveCommand(repo); err != nil {
			// If command is not in ccmd.yaml, that's okay
			if !strings.Contains(err.Error(), "not found in configuration") {
				return fmt.Errorf("failed to update ccmd.yaml: %w", err)
			}
		}
	}

	// Update ccmd-lock.yaml if it exists
	if manager.LockExists() {
		lockFile, err := manager.LoadLockFile()
		if err != nil {
			return fmt.Errorf("failed to load lock file: %w", err)
		}

		// Remove from lock file
		if lockFile.RemoveCommand(commandName) {
			if err := manager.SaveLockFile(lockFile); err != nil {
				return fmt.Errorf("failed to save lock file: %w", err)
			}
		}
	}

	return nil
}

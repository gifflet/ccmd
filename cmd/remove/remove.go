package remove

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
)

// NewCommand creates a new remove command.
func NewCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <command-name>",
		Short: "Remove an installed command",
		Long:  `Remove an installed command and clean up all associated files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal without confirmation")

	return cmd
}

func runRemove(commandName string, force bool) error {
	// Check if command exists
	exists, err := commands.CommandExists(commandName, "", nil)
	if err != nil {
		return fmt.Errorf("failed to check command existence: %w", err)
	}

	if !exists {
		output.Error("Command '%s' is not installed", commandName)
		return fmt.Errorf("command not found")
	}

	// Get command info for display
	cmdInfo, err := commands.GetCommandInfo(commandName, "", nil)
	if err != nil {
		return fmt.Errorf("failed to get command info: %w", err)
	}

	// Confirm removal if not forced
	if !force {
		output.Info("Command details:")
		output.Info("  Name: %s", cmdInfo.Name)
		output.Info("  Version: %s", cmdInfo.Version)
		if desc, ok := cmdInfo.Metadata["description"]; ok && desc != "" {
			output.Info("  Description: %s", desc)
		}

		output.Warning("\nThis will permanently remove the command and all its files.")
		output.Info("Are you sure you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if !isConfirmation(response) {
			output.Info("Removal cancelled")
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
	output.Success("Command '%s' has been successfully removed", commandName)

	return nil
}

func isConfirmation(response string) bool {
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}


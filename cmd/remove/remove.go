/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package remove

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
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
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Get command info
	cmdInfo, err := core.GetCommandInfo(commandName, cwd)
	if err != nil {
		return fmt.Errorf("command %q is not installed", commandName)
	}

	// Confirm removal if not forced
	if !force {
		output.PrintInfof("Command details:")
		output.PrintInfof("  Name: %s", cmdInfo.Name)
		output.PrintInfof("  Version: %s", cmdInfo.Version)
		if cmdInfo.Description != "" {
			output.PrintInfof("  Description: %s", cmdInfo.Description)
		}

		output.PrintWarningf("\nThis will permanently remove the command and all its files.")
		output.Printf("Are you sure you want to continue? [y/N]: ")

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
	removeOpts := core.RemoveOptions{
		Name:        commandName,
		Force:       force,
		UpdateFiles: save,
	}

	if err := core.Remove(removeOpts); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to remove command: %w", err)
	}

	spinner.Stop()

	// Success message
	output.PrintSuccessf("Command '%s' has been removed", commandName)

	if save {
		output.PrintInfof("Updated ccmd.yaml and ccmd-lock.yaml")
	}

	return nil
}

func isConfirmation(response string) bool {
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

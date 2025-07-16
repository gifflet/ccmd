/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package update

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
)

// NewCommand creates the update command
func NewCommand() *cobra.Command {
	var (
		all       bool
		checkOnly bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "update [command]",
		Short: "Update installed commands to their latest versions",
		Long: `Update installed commands to their latest versions.

With --all flag, it updates all installed commands.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

			opts := core.UpdateOptions{
				Name:      name,
				All:       all,
				CheckOnly: checkOnly,
				Force:     force,
			}

			_, err := core.Update(context.Background(), opts)
			return err
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Update all installed commands")
	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates without installing")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if version appears current")

	return cmd
}

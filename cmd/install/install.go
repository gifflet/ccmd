/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package install

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
)

// NewCommand creates a new install command.
func NewCommand() *cobra.Command {
	var (
		version string
		name    string
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "install [repository]",
		Short: "Install a command from a Git repository or from ccmd.yaml",
		Long: `Install a command from a Git repository or install all commands from ccmd.yaml.

When no repository is provided, installs all commands defined in the project's ccmd.yaml file.
When a repository is provided, installs the command and adds it to ccmd.yaml and ccmd-lock.yaml.

Examples:
  # Install all commands from ccmd.yaml
  ccmd install

  # Install latest version
  ccmd install github.com/user/repo

  # Install specific version
  ccmd install github.com/user/repo@v1.0.0

  # Install with custom name
  ccmd install github.com/user/repo --name mycommand

  # Force reinstall
  ccmd install github.com/user/repo --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) == 0 {
				// Install from config
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				return core.InstallFromConfig(ctx, cwd, force)
			}

			// Install specific repository
			opts := core.InstallOptions{
				Repository: args[0],
				Version:    version,
				Name:       name,
				Force:      force,
			}

			if err := core.Install(ctx, opts); err != nil {
				return err
			}

			// Extract command name for usage info
			commandName := name
			if commandName == "" {
				repo, _ := core.ParseRepositorySpec(args[0])
				path := core.ExtractRepoPath(core.NormalizeRepositoryURL(repo))
				parts := strings.Split(path, "/")
				if len(parts) > 0 {
					commandName = parts[len(parts)-1]
				}
			}

			output.PrintInfof("\nTo use the command, run:")
			output.PrintInfof("/%s", commandName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version/tag to install")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Override command name")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall if already exists")

	return cmd
}

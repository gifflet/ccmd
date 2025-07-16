/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package sync

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
)

// NewCommand creates the sync command
func NewCommand() *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize installed commands with ccmd.yaml",
		Long: `Synchronize installed commands with ccmd.yaml configuration.

This command will:
- Install commands listed in ccmd.yaml but not installed
- Remove commands installed but not in ccmd.yaml
- Update ccmd-lock.yaml to reflect current state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(dryRun, force)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force sync without confirmation")

	return cmd
}

func runSync(dryRun, force bool) error {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Analyze what needs to be done
	analysis, err := core.AnalyzeSync(cwd)
	if err != nil {
		return err
	}

	// Show analysis
	if analysis.InSync {
		output.PrintInfof("✓ Commands are already in sync with ccmd.yaml")
		return nil
	}

	output.PrintInfof("=== Sync Analysis ===")
	if len(analysis.ToInstall) > 0 {
		output.PrintInfof("\nCommands to install:")
		for _, cmd := range analysis.ToInstall {
			output.Printf("  + %s", cmd.Repo)
		}
	}

	if len(analysis.ToRemove) > 0 {
		output.PrintInfof("\nCommands to remove:")
		for _, name := range analysis.ToRemove {
			output.Printf("  - %s", name)
		}
	}

	if dryRun {
		output.PrintInfof("\n(dry-run mode - no changes made)")
		return nil
	}

	// Execute sync
	opts := core.SyncOptions{
		ProjectPath: cwd,
		DryRun:      dryRun,
		Force:       force,
	}

	result, err := core.Sync(context.Background(), opts)
	if err != nil {
		return err
	}

	// Show results
	if len(result.Installed) > 0 {
		output.PrintInfof("\nInstalled commands:")
		for _, cmd := range result.Installed {
			output.PrintSuccessf("  ✓ %s", cmd)
		}
	}

	if len(result.Removed) > 0 {
		output.PrintInfof("\nRemoved commands:")
		for _, name := range result.Removed {
			output.PrintSuccessf("  ✓ %s", name)
		}
	}

	if len(result.Failed) > 0 {
		output.PrintErrorf("\nFailed operations:")
		for _, failure := range result.Failed {
			output.PrintErrorf("  ✗ %s %s: %v", failure.Operation, failure.Command, failure.Error)
		}
	}

	if len(result.Failed) == 0 {
		output.PrintSuccessf("\n✓ Sync completed successfully")
	} else {
		output.PrintWarningf("\n⚠ Sync completed with %d error(s)", len(result.Failed))
	}

	return nil
}

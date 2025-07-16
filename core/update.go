/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/output"
)

// UpdateOptions represents options for updating commands
type UpdateOptions struct {
	Name      string // Command name (empty for all)
	All       bool   // Update all commands
	CheckOnly bool   // Only check for updates without installing
	Force     bool   // Force update even if version appears current
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	UpdatedCount int
	FailedCount  int
	CheckedCount int
}

// Update updates one or more installed commands
func Update(ctx context.Context, opts UpdateOptions) (*UpdateResult, error) {
	if opts.All && opts.Name != "" {
		return nil, errors.InvalidInput("cannot specify command name with --all flag")
	}

	if !opts.All && opts.Name == "" {
		return nil, errors.InvalidInput("command name required (or use --all)")
	}

	if opts.All {
		return updateAllCommands(ctx, opts.CheckOnly, opts.Force)
	}

	return updateSingleCommand(ctx, opts.Name, opts.CheckOnly, opts.Force)
}

func updateAllCommands(ctx context.Context, checkOnly, force bool) (*UpdateResult, error) {
	// List all commands
	commands, err := List(ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}

	if len(commands) == 0 {
		output.PrintInfof("No commands installed")
		return &UpdateResult{}, nil
	}

	output.PrintInfof("Checking %d commands for updates...", len(commands))

	result := &UpdateResult{}

	for _, cmd := range commands {
		output.PrintInfof("\nChecking %s...", cmd.Name)
		result.CheckedCount++

		_, version := ParseCommandSpec(cmd.Resolved)

		needsUpdate, reason := shouldUpdateCommand(cmd.Name, version, force)

		if checkOnly {
			if strings.Contains(reason, "pinned to commit") {
				output.PrintInfof("Installed with commit %.7s (no updates for commits)", version)
			} else if needsUpdate {
				output.PrintWarningf("Update available for %s", cmd.Name)
			} else {
				output.PrintInfof("%s is up to date", cmd.Name)
			}
			continue
		}

		if !needsUpdate {
			output.PrintInfof("%s is already up to date", cmd.Name)
			continue
		}

		if force && isCommitHash(version) {
			output.PrintWarningf("Force updating command installed with commit %.7s", version)
		}

		// Force reinstall to update
		// Note: We don't pass Name to allow Install to use the name from ccmd.yaml
		// This handles cases where the command name changed in the remote repository
		opts := InstallOptions{
			Repository: cmd.Repository,
			Version:    version,
			Force:      true,
		}

		if err := Install(ctx, opts); err != nil {
			output.PrintErrorf("Failed to update %s: %v", cmd.Name, err)
			result.FailedCount++
		} else {
			output.PrintSuccessf("Updated %s", cmd.Name)
			result.UpdatedCount++
		}
	}

	// Summary
	output.PrintInfof("\n=== Update Summary ===")
	if result.UpdatedCount > 0 {
		output.PrintSuccessf("%d command(s) updated", result.UpdatedCount)
	}
	if result.FailedCount > 0 {
		output.PrintErrorf("%d command(s) failed to update", result.FailedCount)
	}

	return result, nil
}

func checkIfUpdateNeeded(commandName, version string) (bool, error) {
	// Get project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return false, err
	}

	commandDir := filepath.Join(projectRoot, ".claude", "commands", commandName)
	if !dirExists(commandDir) {
		// Command directory doesn't exist, needs update
		return true, nil
	}

	localCommit, err := gitGetRefCommit(commandDir, version)
	if err != nil {
		// Local ref doesn't exist or error getting it
		return true, nil
	}

	remoteCommit, err := gitGetRemoteRefCommit(commandDir, version)
	if err != nil {
		// Can't get remote ref, assume update needed
		return true, nil
	}

	return localCommit != remoteCommit, nil
}

// shouldUpdateCommand determines if a command needs updating based on version and flags
func shouldUpdateCommand(commandName, version string, force bool) (needsUpdate bool, reason string) {
	if force {
		return true, "forced update"
	}

	if version == "" {
		return true, "tracks latest version"
	}

	if isCommitHash(version) {
		return false, fmt.Sprintf("pinned to commit %.7s", version)
	}

	updateNeeded, err := checkIfUpdateNeeded(commandName, version)
	if err != nil {
		return true, fmt.Sprintf("check failed: %v", err)
	}

	if !updateNeeded {
		return false, "already up to date"
	}

	return true, "update available"
}

func updateSingleCommand(ctx context.Context, name string, checkOnly, force bool) (*UpdateResult, error) {
	// Get command info
	cmdInfo, err := GetCommandInfo(name, "")
	if err != nil {
		return nil, errors.NotFound(fmt.Sprintf("command %q", name))
	}

	output.PrintInfof("Checking %s for updates...", name)

	result := &UpdateResult{CheckedCount: 1}

	// Extract version from Resolved field (repo@version or repo@commit)
	_, version := ParseCommandSpec(cmdInfo.Resolved)

	// Check if update is needed
	needsUpdate, reason := shouldUpdateCommand(name, version, force)

	if checkOnly {
		// Report status based on reason
		if strings.Contains(reason, "pinned to commit") {
			output.PrintInfof("Command %q is installed with commit %.7s (no updates for commits)", name, version)
		} else if strings.Contains(reason, "tracks latest") {
			output.PrintWarningf("Command %q tracks latest version, update may be available", name)
		} else if strings.Contains(reason, "check failed") {
			output.PrintWarningf("Could not check for updates: %s", reason)
		} else if needsUpdate {
			output.PrintWarningf("Update available for %q", name)
		} else {
			output.PrintInfof("Command %q is up to date", name)
		}
		output.PrintInfof("Current version: %s", cmdInfo.Version)
		return result, nil
	}

	// Handle commit hash updates
	if version != "" && isCommitHash(version) && !force {
		output.PrintWarningf("Command %q is installed with commit hash %.7s and cannot be updated.", name, version)
		output.PrintWarningf("To change versions, reinstall with a different tag, branch, or commit.")
		return result, nil
	}

	// Check if update is needed
	if !needsUpdate {
		output.PrintInfof("Command %q is already up to date", name)
		return result, nil
	}

	// Show update reason if special case
	if force && isCommitHash(version) {
		output.PrintWarningf("Force updating command installed with commit hash %.7s", version)
	} else if strings.Contains(reason, "check failed") {
		output.PrintWarningf("Proceeding with update: %s", reason)
	}

	// Force reinstall to update
	// Note: We don't pass Name to allow Install to use the name from ccmd.yaml
	// This handles cases where the command name changed in the remote repository
	opts := InstallOptions{
		Repository: cmdInfo.Repository,
		Version:    version,
		Force:      true,
	}

	if err := Install(ctx, opts); err != nil {
		result.FailedCount = 1
		return result, fmt.Errorf("failed to update: %w", err)
	}

	// Get the current name of the command after installation
	projectRoot, err := findProjectRoot()
	if err == nil {
		currentName, _ := findExistingCommandByRepo(projectRoot, ExtractRepoPath(cmdInfo.Repository))
		if currentName != "" && currentName != name {
			output.PrintSuccessf("Command %q updated to %q successfully", name, currentName)
		} else {
			output.PrintSuccessf("Command %q updated successfully", name)
		}
	} else {
		// Fallback if we can't determine the project root
		output.PrintSuccessf("Command %q updated successfully", name)
	}

	result.UpdatedCount = 1
	return result, nil
}

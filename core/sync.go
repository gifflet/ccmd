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
	"strings"
)

// SyncOptions represents options for syncing commands
type SyncOptions struct {
	ProjectPath string
	DryRun      bool
	Force       bool
}

// SyncAnalysis represents the analysis of what needs to be synced
type SyncAnalysis struct {
	ToInstall []ConfigCommand
	ToRemove  []string
	InSync    bool
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Installed []string
	Removed   []string
	Failed    []SyncError
}

// SyncError represents an error during sync operation
type SyncError struct {
	Command   string
	Operation string // "install" or "remove"
	Error     error
}

// AnalyzeSync analyzes what needs to be synced between config and installed commands
func AnalyzeSync(projectPath string) (*SyncAnalysis, error) {
	// Load project config
	config, err := LoadProjectConfig(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ccmd.yaml: %w", err)
	}

	// Get installed commands
	installed, err := List(ListOptions{ProjectPath: projectPath})
	if err != nil {
		return nil, fmt.Errorf("failed to list installed commands: %w", err)
	}

	// Create maps for easy lookup
	installedMap := make(map[string]CommandDetail)
	for _, cmd := range installed {
		installedMap[cmd.Name] = cmd
	}

	configCommands := config.GetConfigCommands()
	configMap := make(map[string]ConfigCommand)
	for _, cmd := range configCommands {
		// Extract name from repo
		name := extractCommandName(cmd.Repo)
		configMap[name] = cmd
	}

	// Analyze what needs to be done
	var toInstall []ConfigCommand
	var toRemove []string

	// Find commands to install
	for name, cmd := range configMap {
		if _, exists := installedMap[name]; !exists {
			toInstall = append(toInstall, cmd)
		}
	}

	// Find commands to remove
	for name := range installedMap {
		if _, exists := configMap[name]; !exists {
			toRemove = append(toRemove, name)
		}
	}

	return &SyncAnalysis{
		ToInstall: toInstall,
		ToRemove:  toRemove,
		InSync:    len(toInstall) == 0 && len(toRemove) == 0,
	}, nil
}

// Sync synchronizes installed commands with the project configuration
func Sync(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	// Analyze what needs to be done
	analysis, err := AnalyzeSync(opts.ProjectPath)
	if err != nil {
		return nil, err
	}

	// If in sync, return empty result
	if analysis.InSync {
		return &SyncResult{}, nil
	}

	// If dry run, return without making changes
	if opts.DryRun {
		return &SyncResult{}, nil
	}

	result := &SyncResult{
		Installed: []string{},
		Removed:   []string{},
		Failed:    []SyncError{},
	}

	// Install missing commands
	for _, cmd := range analysis.ToInstall {
		repository := normalizeRepository(cmd.Repo)

		installOpts := InstallOptions{
			Repository: repository,
			Version:    cmd.Version,
			Force:      false,
		}

		if err := Install(ctx, installOpts); err != nil {
			result.Failed = append(result.Failed, SyncError{
				Command:   cmd.Repo,
				Operation: "install",
				Error:     err,
			})
		} else {
			result.Installed = append(result.Installed, cmd.Repo)
		}
	}

	// Remove extra commands
	for _, name := range analysis.ToRemove {
		removeOpts := RemoveOptions{
			Name:        name,
			Force:       opts.Force,
			UpdateFiles: false, // Don't update ccmd.yaml since we're syncing from it
		}

		if err := Remove(removeOpts); err != nil {
			result.Failed = append(result.Failed, SyncError{
				Command:   name,
				Operation: "remove",
				Error:     err,
			})
		} else {
			result.Removed = append(result.Removed, name)
		}
	}

	return result, nil
}

// normalizeRepository converts a short repo reference to a full URL
func normalizeRepository(repo string) string {
	if isFullURL(repo) {
		return repo
	}
	return "https://github.com/" + repo + ".git"
}

// isFullURL checks if a string is a full URL
func isFullURL(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "git@")
}

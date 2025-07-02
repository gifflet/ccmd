/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gifflet/ccmd/internal/installer"
	"github.com/gifflet/ccmd/pkg/errors"
)

// InstallOptions represents options for installing a command
type InstallOptions struct {
	Repository string // Git repository URL
	Version    string // Version/tag to install (optional)
	Name       string // Override command name (optional)
	Force      bool   // Force reinstall if already exists
}

// Install installs a command from a Git repository
func Install(opts InstallOptions) error {
	if opts.Repository == "" {
		return errors.InvalidInput("repository URL is required")
	}

	// Get config directory (project-local)
	configDir := ".claude"
	commandsDir := filepath.Join(configDir, "commands")

	// Create installer options
	installerOpts := installer.Options{
		Repository: opts.Repository,
		Version:    opts.Version,
		Name:       opts.Name,
		Force:      opts.Force,
		InstallDir: commandsDir,
	}

	// Create installer
	inst, err := installer.New(installerOpts)
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Perform installation
	ctx := context.Background()
	if err := inst.Install(ctx); err != nil {
		return err
	}

	return nil
}

// ParseRepositorySpec parses a repository specification (URL[@version])
func ParseRepositorySpec(spec string) (repository, version string) {
	return installer.ParseRepositorySpec(spec)
}

// InstallFromProject installs all commands defined in the project's ccmd.yaml
func InstallFromProject(projectPath string, force bool) error {
	ctx := context.Background()
	return installer.InstallFromConfig(ctx, projectPath, force)
}

// NormalizeRepositoryURL normalizes various repository formats to a full URL
func NormalizeRepositoryURL(url string) string {
	return installer.NormalizeRepositoryURL(url)
}

// ExtractRepoPath extracts owner/repo from a Git URL
func ExtractRepoPath(gitURL string) string {
	return installer.ExtractRepoPath(gitURL)
}

// GetInstalledCommands returns a list of installed commands
func GetInstalledCommands(projectPath string) ([]InstalledCommand, error) {
	cm := installer.NewCommandManager(projectPath)

	commands, err := cm.GetInstalledCommands()
	if err != nil {
		return nil, err
	}

	// Convert to our public type
	result := make([]InstalledCommand, len(commands))
	for i, cmd := range commands {
		result[i] = InstalledCommand{
			Name:        cmd.Name,
			Version:     cmd.Version,
			Description: cmd.Description,
			Author:      cmd.Author,
			Repository:  cmd.Repository,
			Path:        cmd.Path,
		}
	}

	return result, nil
}

// InstalledCommand represents an installed command
type InstalledCommand struct {
	Name        string
	Version     string
	Description string
	Author      string
	Repository  string
	Path        string
}

// copyDirectory recursively copies a directory (deprecated - use installer package)
// Kept for backward compatibility
func copyDirectory(_ interface{}, _, _ string) error {
	return fmt.Errorf("copyDirectory is deprecated, use installer package")
}

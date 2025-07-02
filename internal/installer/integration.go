/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
	"github.com/gifflet/ccmd/pkg/project"
)

// IntegrationOptions provides options for integrated installation
type IntegrationOptions struct {
	Repository    string // Git repository URL or shorthand (e.g., "user/repo")
	Version       string // Version/tag to install
	Name          string // Override command name
	Force         bool   // Force reinstall
	ProjectPath   string // Path to project (for ccmd.yaml updates)
	GlobalInstall bool   // Install globally vs project-local
}

// ParseRepositorySpec parses a repository specification (URL[@version])
func ParseRepositorySpec(spec string) (repository, version string) {
	// Find the last @ that could be a version separator
	lastAt := strings.LastIndex(spec, "@")

	// If no @ found, it's all repository
	if lastAt == -1 {
		return spec, ""
	}

	// Check if this is an SSH URL (git@host:...)
	beforeAt := spec[:lastAt]
	afterAt := spec[lastAt+1:]

	// If the part before @ looks like a protocol or is very short,
	// and the part after @ contains a colon, it's likely an SSH URL
	if (strings.HasPrefix(beforeAt, "git") || strings.HasPrefix(beforeAt, "ssh") || len(beforeAt) < 5) &&
		strings.Contains(afterAt, ":") && !strings.Contains(afterAt, "://") {
		// This is likely an SSH URL like git@github.com:user/repo
		return spec, ""
	}

	// Otherwise, treat everything after the last @ as a version
	if afterAt != "" {
		return beforeAt, afterAt
	}

	return spec, ""
}

// NormalizeRepositoryURL normalizes various repository formats to a full URL
func NormalizeRepositoryURL(url string) string {
	url = strings.TrimSpace(url)

	// Handle GitHub shorthand (user/repo)
	if !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, "/")
		if len(parts) == 2 && !strings.Contains(parts[0], ".") {
			// Assume GitHub shorthand
			url = fmt.Sprintf("https://github.com/%s.git", strings.TrimSuffix(url, ".git"))
			return url
		}
	}

	// If URL doesn't have a protocol, add https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "git@") {
		// Check if it looks like a GitHub/GitLab/etc URL
		if strings.Contains(url, "github.com") ||
			strings.Contains(url, "gitlab.com") ||
			strings.Contains(url, "bitbucket.org") {
			url = "https://" + url
		}
	}

	// Ensure .git suffix for consistency
	if !strings.HasSuffix(url, ".git") && !strings.Contains(url, ".git@") {
		url += ".git"
	}

	return url
}

// ExtractRepoPath extracts owner/repo from a Git URL
func ExtractRepoPath(gitURL string) string {
	// Remove protocol
	url := strings.TrimPrefix(gitURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git://")

	// Handle SSH URLs
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
	}

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Extract path after domain
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		// Return owner/repo
		return parts[1] + "/" + parts[2]
	}

	return ""
}

// InstallCommand performs an integrated installation with project management
func InstallCommand(ctx context.Context, opts IntegrationOptions, addToConfig bool) error {
	log := logger.WithField("component", "installer-integration")

	// Parse repository spec if version is included
	repo, specVersion := ParseRepositorySpec(opts.Repository)
	if specVersion != "" && opts.Version == "" {
		opts.Version = specVersion
	}

	// Normalize repository URL
	repo = NormalizeRepositoryURL(repo)

	log.WithFields(logger.Fields{
		"repository": repo,
		"version":    opts.Version,
		"name":       opts.Name,
		"force":      opts.Force,
		"global":     opts.GlobalInstall,
	}).Debug("starting integrated installation")

	// Determine installation directory
	installDir := ".claude/commands"
	if opts.GlobalInstall {
		// Global installation path support will be added in a future release
		return fmt.Errorf("global installation not yet supported")
	}

	// Create installer
	installerOpts := Options{
		Repository:  repo,
		Version:     opts.Version,
		Name:        opts.Name,
		Force:       opts.Force,
		InstallDir:  installDir,
		ProjectPath: opts.ProjectPath,
	}

	installer, err := New(installerOpts)
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Perform installation
	if err := installer.Install(ctx); err != nil {
		return err
	}

	// Update project files if in a project context
	if opts.ProjectPath != "" && addToConfig {
		if err := updateProjectFiles(opts.ProjectPath, repo, opts.Version, opts.Name); err != nil {
			// Check if it's just "already exists" error
			if !strings.Contains(err.Error(), "already exists in configuration") {
				log.WithError(err).Warn("failed to update project files")
			}
		}
	}

	return nil
}

// updateProjectFiles updates ccmd.yaml and ccmd-lock.yaml in the project
func updateProjectFiles(projectPath, repository, version, _ string) error {
	pm := project.NewManager(projectPath)

	// Check if project has ccmd.yaml
	if !pm.ConfigExists() {
		if err := pm.InitializeConfig(); err != nil {
			return errors.FileError("initialize ccmd.yaml", filepath.Join(projectPath, "ccmd.yaml"), err)
		}
	}

	// Extract owner/repo from repository URL
	repoPath := ExtractRepoPath(repository)
	if repoPath == "" {
		return errors.InvalidInput("failed to extract repository path")
	}

	// Add command to ccmd.yaml
	if err := pm.AddCommand(repoPath, version); err != nil {
		return errors.FileError("update ccmd.yaml", filepath.Join(projectPath, "ccmd.yaml"), err)
	}

	// Update lock file
	lockFile, err := pm.LoadLockFile()
	if err != nil {
		// Create new lock file if it doesn't exist
		if os.IsNotExist(err) {
			lockFile = project.NewLockFile()
		} else {
			return errors.FileError("load lock file", filepath.Join(projectPath, "ccmd-lock.yaml"), err)
		}
	}

	// Save the lock file (ensures it exists for future operations)
	if err := pm.SaveLockFile(lockFile); err != nil {
		return errors.FileError("save lock file", filepath.Join(projectPath, "ccmd-lock.yaml"), err)
	}

	// Note: The actual lock file update is handled by the installer itself
	// This is just for project-level tracking

	return nil
}

// InstallFromConfig installs all commands defined in a project's ccmd.yaml
func InstallFromConfig(ctx context.Context, projectPath string, force bool) error {
	log := logger.WithField("component", "installer-integration")

	pm := project.NewManager(projectPath)

	// Check if config exists
	if !pm.ConfigExists() {
		return errors.NotFound("no ccmd.yaml found in project")
	}

	// Load config
	config, err := pm.LoadConfig()
	if err != nil {
		return errors.FileError("load ccmd.yaml", filepath.Join(projectPath, "ccmd.yaml"), err)
	}

	// Get normalized commands
	configCommands, err := config.GetCommands()
	if err != nil {
		return errors.InvalidInput(fmt.Sprintf("failed to parse commands: %v", err))
	}

	if len(configCommands) == 0 {
		log.Info("no commands found in ccmd.yaml")
		output.PrintInfof("No commands found in ccmd.yaml")
		return nil
	}

	// Load existing lock file to preserve original URLs
	lockFile, lockErr := pm.LoadLockFile()
	if lockErr != nil && !os.IsNotExist(lockErr) {
		log.WithError(lockErr).Warn("failed to load lock file")
		lockFile = nil
	}

	log.WithField("count", len(configCommands)).Info("installing commands from ccmd.yaml")
	output.PrintInfof("Installing %d command(s) from ccmd.yaml", len(configCommands))

	// Install each command
	var installErrors []error
	successCount := 0

	for _, cmd := range configCommands {
		_, repoName, err := cmd.ParseOwnerRepo()
		if err != nil {
			output.Error("Failed to parse repository %s: %v", cmd.Repo, err)
			installErrors = append(installErrors,
				errors.InvalidInput(fmt.Sprintf("failed to parse repository %s: %v", cmd.Repo, err)))
			continue
		}

		manager := project.NewManager(".claude")
		exists, err := manager.CommandExists(repoName)
		if err == nil && exists && !force {
			output.PrintInfof("Command '%s' is already installed", repoName)
			successCount++
			continue
		}

		// Determine repository URL - prefer lock file URL if available
		var repository string
		if lockFile != nil {
			if cmdLock, found := lockFile.GetCommand(repoName); found && cmdLock.Source != "" {
				// Use original source URL from lock file
				repository = cmdLock.Source
				log.WithFields(logger.Fields{
					"command": repoName,
					"source":  repository,
				}).Debug("using repository URL from lock file")
			}
		}

		// If not found in lock file, normalize the repository URL
		if repository == "" {
			repository = NormalizeRepositoryURL(fmt.Sprintf("github.com/%s", cmd.Repo))
			log.WithFields(logger.Fields{
				"command": repoName,
				"source":  repository,
			}).Debug("using normalized repository URL")
		}

		log.WithFields(logger.Fields{
			"repository": cmd.Repo,
			"version":    cmd.Version,
		}).Debug("installing command")

		// Show user output for installation
		output.PrintInfof("\nInstalling %s", cmd.Repo)
		if cmd.Version != "" {
			output.PrintInfof("Version: %s", cmd.Version)
		}

		// Create spinner for installation process
		spinner := output.NewSpinner(fmt.Sprintf("Installing %s...", cmd.Repo))
		spinner.Start()

		// Parse repo to get command name
		_, repoName, err = cmd.ParseOwnerRepo()
		if err != nil {
			spinner.Stop()
			output.Error("Failed to parse repository %s: %v", cmd.Repo, err)
			installErrors = append(installErrors,
				errors.InvalidInput(fmt.Sprintf("failed to parse repository %s: %v", cmd.Repo, err)))
			continue
		}

		// Install the command
		opts := IntegrationOptions{
			Repository:  repository,
			Version:     cmd.Version,
			Name:        repoName,
			Force:       force,
			ProjectPath: projectPath,
		}

		if err := InstallCommand(ctx, opts, false); err != nil {
			spinner.Stop()
			output.Error("Failed to install %s: %v", cmd.Repo, err)
			installErrors = append(installErrors, fmt.Errorf("failed to install command %s: %w", cmd.Repo, err))
			continue
		}

		spinner.Stop()
		output.PrintSuccessf("Command '%s' has been successfully installed", repoName)
		successCount++
	}

	// Report results
	log.WithFields(logger.Fields{
		"success": successCount,
		"failed":  len(installErrors),
		"total":   len(configCommands),
	}).Info("installation complete")

	// Show final summary to user
	output.PrintInfof("\nSuccessfully installed %d out of %d command(s)", successCount, len(configCommands))

	// Return error if any installations failed
	if len(installErrors) > 0 {
		// Create a combined error message
		return fmt.Errorf("some commands failed to install: %d errors occurred", len(installErrors))
	}

	return nil
}

// CommandManager provides high-level command management operations
type CommandManager struct {
	projectPath string
	fileSystem  fs.FileSystem
	logger      logger.Logger
}

// NewCommandManager creates a new command manager
func NewCommandManager(projectPath string) *CommandManager {
	return &CommandManager{
		projectPath: projectPath,
		fileSystem:  fs.NewOSFileSystem(),
		logger:      logger.WithField("component", "command-manager"),
	}
}

// Install installs a command with the given options
func (cm *CommandManager) Install(ctx context.Context, repository, version, name string, force bool) error {
	opts := IntegrationOptions{
		Repository:  repository,
		Version:     version,
		Name:        name,
		Force:       force,
		ProjectPath: cm.projectPath,
	}

	return InstallCommand(ctx, opts, true)
}

// InstallFromProject installs all commands from the project's ccmd.yaml
func (cm *CommandManager) InstallFromProject(ctx context.Context, force bool) error {
	return InstallFromConfig(ctx, cm.projectPath, force)
}

// GetInstalledCommands returns a list of installed commands
func (cm *CommandManager) GetInstalledCommands() ([]InstalledCommand, error) {
	installDir := filepath.Join(cm.projectPath, ".claude", "commands")

	entries, err := cm.fileSystem.ReadDir(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledCommand{}, nil
		}
		return nil, errors.FileError("read commands directory", installDir, err)
	}

	commands := make([]InstalledCommand, 0, len(entries))
	seen := make(map[string]bool) // Track seen commands to avoid duplicates

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip if we've already seen this command
		if seen[entry.Name()] {
			continue
		}
		seen[entry.Name()] = true

		commandPath := filepath.Join(installDir, entry.Name())
		metadataPath := filepath.Join(commandPath, "ccmd.yaml")

		// Check if metadata exists
		if _, err := cm.fileSystem.Stat(metadataPath); err != nil {
			continue
		}

		// Read metadata
		data, err := cm.fileSystem.ReadFile(metadataPath)
		if err != nil {
			cm.logger.WithError(err).WithField("command", entry.Name()).Warn("failed to read metadata")
			continue
		}

		var metadata models.CommandMetadata
		if err := yaml.Unmarshal(data, &metadata); err != nil {
			cm.logger.WithError(err).WithField("command", entry.Name()).Warn("failed to parse metadata")
			continue
		}

		commands = append(commands, InstalledCommand{
			Name:        entry.Name(),
			Version:     metadata.Version,
			Description: metadata.Description,
			Author:      metadata.Author,
			Repository:  metadata.Repository,
			Path:        commandPath,
		})
	}

	return commands, nil
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

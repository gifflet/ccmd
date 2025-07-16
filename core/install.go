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
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
	"github.com/gifflet/ccmd/pkg/output"
)

// InstallOptions represents options for installing a command
type InstallOptions struct {
	Repository string // Git repository URL or shorthand
	Version    string // Version/tag to install (optional)
	Commit     string // Specific commit to install (used when different from Version)
	Name       string // Override command name (optional)
	Force      bool   // Force reinstall if already exists
}

// Install installs a command from a Git repository
func Install(_ context.Context, opts InstallOptions) error {
	log := logger.New()

	// Validate input
	if opts.Repository == "" {
		return errors.InvalidInput("repository URL is required")
	}

	// Parse repository spec (repo@version)
	repo, version := ParseRepositorySpec(opts.Repository)
	if version != "" && opts.Version == "" {
		opts.Version = version
	}
	opts.Repository = repo

	// Normalize repository URL
	repoURL := NormalizeRepositoryURL(opts.Repository)
	log.WithField("repository", repoURL).Debug("Installing command")

	// Get project paths
	projectRoot, err := findProjectRoot()
	if err != nil {
		return errors.FileError("find project root", "", err)
	}

	ccmdDir := filepath.Join(projectRoot, ".claude")
	commandsDir := filepath.Join(ccmdDir, "commands")

	// Create commands directory if needed
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return errors.FileError("create commands directory", commandsDir, err)
	}

	// Create temp directory for cloning
	tempDir, err := os.MkdirTemp("", "ccmd-install-*")
	if err != nil {
		return errors.FileError("create temp directory", "", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	output.PrintInfof("Cloning repository %s...", repoURL)
	// Use commit if specified, otherwise use version
	cloneVersion := opts.Version
	if opts.Commit != "" {
		cloneVersion = opts.Commit
	}
	if err := gitClone(repoURL, tempDir, cloneVersion); err != nil {
		return errors.GitError("clone", err)
	}

	// Read and validate metadata
	metadataPath := filepath.Join(tempDir, "ccmd.yaml")
	metadata, err := readCommandMetadata(metadataPath)
	if err != nil {
		return err
	}

	// Determine command name
	commandName := opts.Name
	if commandName == "" {
		commandName = metadata.Name
		if commandName == "" {
			// Extract from repository path
			commandName = extractCommandName(repoURL)
		}
	}

	// Validate command name
	if err := validateCommandName(commandName); err != nil {
		return err
	}

	// Check for existing installations from the same repository
	targetRepoPath := ExtractRepoPath(repoURL)
	existingCommand, err := findExistingCommandByRepo(projectRoot, targetRepoPath)
	if err != nil {
		return errors.FileError("check existing commands", "", err)
	}

	// If force installing and found existing command from same repo
	if opts.Force && existingCommand != "" && existingCommand != commandName {
		output.PrintInfof("Removing previous installation %q from the same repository...", existingCommand)

		// Remove the old command directory
		oldCommandDir := filepath.Join(commandsDir, existingCommand)
		if err := os.RemoveAll(oldCommandDir); err != nil {
			return errors.FileError("remove previous command", oldCommandDir, err)
		}

		// Remove the old standalone .md file
		oldStandalonePath := filepath.Join(ccmdDir, "commands", existingCommand+".md")
		if fileExists(oldStandalonePath) {
			if err := os.Remove(oldStandalonePath); err != nil {
				log.WithError(err).Warn("Failed to remove old standalone documentation")
			}
		}
	}

	// Check if command already exists (by name)
	destDir := filepath.Join(commandsDir, commandName)
	if !opts.Force {
		if _, err := os.Stat(destDir); err == nil {
			return errors.AlreadyExists(fmt.Sprintf("command %q", commandName))
		}
	}

	// Remove existing command if force installing (same name)
	if opts.Force && dirExists(destDir) {
		output.PrintInfof("Removing existing command %q...", commandName)
		if err := os.RemoveAll(destDir); err != nil {
			return errors.FileError("remove existing command", destDir, err)
		}
	}

	// Install command files
	output.PrintInfof("Installing command %q...", commandName)
	if err := copyDirectory(tempDir, destDir); err != nil {
		return errors.FileError("copy command files", destDir, err)
	}

	// Preserve original version from ccmd.yaml before updating metadata
	originalVersion := metadata.Version

	// Update metadata with actual values
	metadata.Name = commandName
	metadata.Repository = repoURL

	// Write updated metadata
	if err := writeCommandMetadata(filepath.Join(destDir, "ccmd.yaml"), metadata); err != nil {
		os.RemoveAll(destDir) // Rollback on error
		return err
	}

	// Create standalone .md file in .claude/commands directory
	standalonePath := filepath.Join(ccmdDir, "commands", commandName+".md")
	if err := createStandaloneDoc(destDir, standalonePath, metadata); err != nil {
		log.WithError(err).Warn("Failed to create standalone documentation")
	}

	// Update lock file
	if err := updateLockFile(projectRoot, commandName, metadata, originalVersion, opts.Version); err != nil {
		log.WithError(err).Warn("Failed to update lock file")
	}

	// Update ccmd.yaml with proper repository format
	repoSpec := opts.Repository
	if strings.Contains(repoSpec, "://") || strings.HasPrefix(repoSpec, "git@") {
		// If it's a full URL, extract owner/repo format
		repoSpec = ExtractRepoPath(repoSpec)
	}
	// Normalize commit hash to 7 characters for config
	versionForConfig := opts.Version
	if isCommitHash(versionForConfig) && len(versionForConfig) > 7 {
		versionForConfig = versionForConfig[:7]
	}
	if err := addToConfig(projectRoot, commandName, repoSpec, versionForConfig); err != nil {
		log.WithError(err).Warn("Failed to update ccmd.yaml")
	}

	output.PrintSuccessf("Command %q installed successfully", commandName)
	return nil
}

// InstallFromConfig installs all commands from project's ccmd.yaml
func InstallFromConfig(ctx context.Context, projectPath string, force bool) error {
	config, err := LoadProjectConfig(projectPath)
	if err != nil {
		return err
	}

	if len(config.Commands) == 0 {
		output.PrintInfof("No commands found in ccmd.yaml")
		return nil
	}

	// Try to load lock file to get exact commits
	lockPath := filepath.Join(projectPath, LockFileName)
	var lockFile *LockFile
	if fileExists(lockPath) {
		lockFile, _ = ReadLockFile(lockPath)
	}

	var installErrors []error
	for _, cmdSpec := range config.Commands {
		repo, version := ParseCommandSpec(cmdSpec)

		// Find exact commit from lock file if available
		commitToInstall := ""
		if lockFile != nil {
			// Find the command by matching the repository source
			normalizedRepo := NormalizeRepositoryURL(repo)
			for _, lockCmd := range lockFile.Commands {
				// Compare normalized URLs
				if NormalizeRepositoryURL(lockCmd.Source) == normalizedRepo {
					// Use the exact commit from lock file
					commitToInstall = lockCmd.Commit
					break
				}
			}
		}

		opts := InstallOptions{
			Repository: repo,
			Version:    version,
			Commit:     commitToInstall,
			Force:      force,
		}

		output.PrintInfof("Installing %s...", cmdSpec)
		if err := Install(ctx, opts); err != nil {
			installErrors = append(installErrors, fmt.Errorf("%s: %w", repo, err))
			output.PrintErrorf("Failed to install %s: %v", repo, err)
		}
	}

	if len(installErrors) > 0 {
		return fmt.Errorf("failed to install %d commands", len(installErrors))
	}

	return nil
}

// Helper functions

func readCommandMetadata(path string) (*ProjectConfig, error) {
	if !fileExists(path) {
		return nil, errors.NotFound("ccmd.yaml not found in repository")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.FileError("read metadata", path, err)
	}

	var metadata ProjectConfig
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, errors.FileError("parse metadata", path, err)
	}

	// Validate required fields
	if err := validateMetadata(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func writeCommandMetadata(path string, metadata *ProjectConfig) error {
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return errors.FileError("marshal metadata", path, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.FileError("write metadata", path, err)
	}

	return nil
}

func validateMetadata(metadata *ProjectConfig) error {
	// Validate metadata fields if it's being used as command metadata
	if metadata.Name != "" || metadata.Version != "" {
		return metadata.Validate()
	}
	// Check for required index.md will be validated after installation
	return nil
}

func validateCommandName(name string) error {
	if name == "" {
		return errors.InvalidInput("command name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return errors.InvalidInput("command name contains invalid characters")
	}

	return nil
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}

func createStandaloneDoc(commandDir, standalonePath string, metadata *ProjectConfig) error {
	// Read index.md
	indexPath := filepath.Join(commandDir, "index.md")
	if !fileExists(indexPath) {
		return errors.NotFound("index.md not found")
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	// Create standalone content with metadata header
	standalone := fmt.Sprintf(`# %s

**Version:** %s
**Author:** %s
**Repository:** %s

%s
`, metadata.Name, metadata.Version, metadata.Author, metadata.Repository, string(content))

	return os.WriteFile(standalonePath, []byte(standalone), 0644)
}

func updateLockFile(projectRoot, commandName string, metadata *ProjectConfig, originalVersion string, requestedVersion string) error {
	lockPath := filepath.Join(projectRoot, LockFileName)
	now := time.Now()

	// Read existing lock file or create new
	var lockFile *LockFile
	if fileExists(lockPath) {
		var err error
		lockFile, err = ReadLockFile(lockPath)
		if err != nil {
			return err
		}
	} else {
		// Initialize new lock file
		lockFile = &LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands:        make(map[string]*LockCommand),
		}
	}

	// Get current commit hash
	commitHash := "unknown"
	commandPath := filepath.Join(projectRoot, ".claude", "commands", commandName)
	if hash, err := gitGetCurrentCommit(commandPath); err == nil {
		commitHash = hash
	}

	// Create resolved URL (source@version or source@commit)
	resolved := metadata.Repository
	if requestedVersion != "" {
		resolved = fmt.Sprintf("%s@%s", metadata.Repository, requestedVersion)
	} else {
		// Try to detect default branch when no version specified
		if defaultBranch, err := gitGetDefaultBranch(commandPath); err == nil {
			resolved = fmt.Sprintf("%s@%s", metadata.Repository, defaultBranch)
		} else if commitHash != "unknown" && len(commitHash) >= 7 {
			resolved = fmt.Sprintf("%s@%s", metadata.Repository, commitHash[:7])
		}
	}

	// Look for existing entry by repository (not just by name)
	repoPath := ExtractRepoPath(metadata.Repository)
	var existingKey string
	var existingCmd *LockCommand

	// Find if this repository is already tracked under a different name
	for key, cmd := range lockFile.Commands {
		if ExtractRepoPath(cmd.Source) == repoPath {
			existingKey = key
			existingCmd = cmd
			break
		}
	}

	// Determine installed_at time
	installedAt := now
	if existingCmd != nil && !existingCmd.InstalledAt.IsZero() {
		installedAt = existingCmd.InstalledAt
	}

	// If found under a different name, remove the old entry
	if existingKey != "" && existingKey != commandName {
		delete(lockFile.Commands, existingKey)
	}

	// Update or create command entry with current name
	lockFile.Commands[commandName] = &LockCommand{
		Name:        commandName,
		Version:     originalVersion,
		Source:      metadata.Repository,
		Resolved:    resolved,
		Commit:      commitHash,
		InstalledAt: installedAt,
		UpdatedAt:   now,
	}

	// Write updated lock file
	return WriteLockFile(lockPath, lockFile)
}

func getInstalledCommands(projectRoot string) (map[string]string, error) {
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")
	installedCommands := make(map[string]string) // commandName -> repository path

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return installedCommands, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Read ccmd.yaml to get repository info
			metadataPath := filepath.Join(commandsDir, entry.Name(), "ccmd.yaml")
			if metadata, err := readCommandMetadata(metadataPath); err == nil && metadata.Repository != "" {
				// Store the simplified repository path
				installedCommands[entry.Name()] = ExtractRepoPath(metadata.Repository)
			}
		}
	}

	return installedCommands, nil
}

// findExistingCommandByRepo finds an existing command installed from the same repository
func findExistingCommandByRepo(projectRoot, targetRepoPath string) (string, error) {
	installedCommands, err := getInstalledCommands(projectRoot)
	if err != nil {
		return "", err
	}

	for commandName, repoPath := range installedCommands {
		if repoPath == targetRepoPath {
			return commandName, nil
		}
	}

	return "", nil
}

func addToConfig(projectRoot, commandName, repository, version string) error {
	// Load existing config or create new one
	var config *ProjectConfig
	if ProjectConfigExists(projectRoot) {
		var err error
		config, err = LoadProjectConfig(projectRoot)
		if err != nil {
			return err
		}
	} else {
		// Create new config with just commands
		config = &ProjectConfig{
			Commands: []string{},
		}
	}

	// Get installed commands mapping
	installedCommands, err := getInstalledCommands(projectRoot)
	if err != nil {
		return err
	}

	// Create command spec
	commandSpec := repository
	if version != "" {
		commandSpec = fmt.Sprintf("%s@%s", repository, version)
	}

	// Check if this specific command is already in config
	found := false
	currentRepo := ExtractRepoPath(repository)

	for i, cmd := range config.Commands {
		repo, _ := ParseCommandSpec(cmd)
		repoPath := ExtractRepoPath(repo)

		// Check if this entry corresponds to our command
		// Either by matching the repository directly or by checking installed commands
		if repoPath == currentRepo {
			// Direct repository match
			config.Commands[i] = commandSpec
			found = true
			break
		} else if installedRepo, exists := installedCommands[commandName]; exists && repoPath == installedRepo {
			// Command is installed and this entry matches its repository
			config.Commands[i] = commandSpec
			found = true
			break
		}
	}

	// Add new command if not found
	if !found {
		config.Commands = append(config.Commands, commandSpec)
	}

	// Save config
	return SaveProjectConfig(projectRoot, config)
}

// Utility functions

func ParseRepositorySpec(spec string) (repository, version string) {
	// For SSH URLs, we need to be careful not to split on @ that's part of the path
	if strings.HasPrefix(spec, "git@") {
		// Find the position after .git or at the end of string
		gitSuffix := ".git@"
		gitIndex := strings.LastIndex(spec, gitSuffix)
		if gitIndex > -1 {
			// Found .git@version pattern
			repository = spec[:gitIndex+4] // Include .git
			version = spec[gitIndex+len(gitSuffix):]
			return repository, version
		}

		// For SSH URLs without .git, we need to find the @ that separates repo from version
		// We'll look for the pattern: owner/repo@version where @ comes right after the repo name
		colonIndex := strings.Index(spec, ":")
		if colonIndex > -1 {
			// Split the URL into parts to analyze
			// Example: git@github.com:owner/repo@version
			// After colon: owner/repo@version
			afterColon := spec[colonIndex+1:]

			// Find the first @ in the part after the colon
			// This handles cases like: owner/repo@version or owner/repo@branch/name
			firstAtIndex := strings.Index(afterColon, "@")
			if firstAtIndex > -1 {
				// Check if this @ is part of a username (like user@company/repo)
				// by checking if there's a / before the @
				beforeAt := afterColon[:firstAtIndex]
				if strings.Contains(beforeAt, "/") {
					// There's a / before @, so this @ is likely a version separator
					repository = spec[:colonIndex+1+firstAtIndex]
					version = spec[colonIndex+1+firstAtIndex+1:]
					return repository, version
				}
				// No / before @, might be user@company pattern, look for next @
				remainingPart := afterColon[firstAtIndex+1:]
				nextAtIndex := strings.Index(remainingPart, "@")
				if nextAtIndex > -1 {
					// Found another @, use it as version separator
					totalIndex := colonIndex + 1 + firstAtIndex + 1 + nextAtIndex
					repository = spec[:totalIndex]
					version = spec[totalIndex+1:]
					return repository, version
				}
			}
		}

		// No version found
		return spec, ""
	}

	// For HTTPS/HTTP URLs and shorthand format, simple split on last @
	lastAtIndex := strings.LastIndex(spec, "@")
	if lastAtIndex == -1 {
		return spec, ""
	}

	repository = spec[:lastAtIndex]
	version = spec[lastAtIndex+1:]
	return repository, version
}

func NormalizeRepositoryURL(url string) string {
	// Handle GitHub shorthand (owner/repo)
	if !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@") {
		if strings.Count(url, "/") == 1 {
			return fmt.Sprintf("https://github.com/%s.git", url)
		}
	}

	// Add .git suffix if missing
	if !strings.HasSuffix(url, ".git") && strings.Contains(url, "github.com") {
		url += ".git"
	}

	return url
}

func extractCommandName(repoURL string) string {
	// Use ExtractRepoPath and get just the repo name
	path := ExtractRepoPath(repoURL)
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for ccmd.yaml
	for {
		if fileExists(filepath.Join(dir, "ccmd.yaml")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, use current directory
			return os.Getwd()
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

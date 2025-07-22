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
func Install(_ context.Context, opts InstallOptions) (string, error) {
	log := logger.New()

	if opts.Repository == "" {
		return "", errors.InvalidInput("repository URL is required")
	}

	repo, version := ParseRepositorySpec(opts.Repository)
	if version != "" && opts.Version == "" {
		opts.Version = version
	}
	opts.Repository = repo

	repoURL := NormalizeRepositoryURL(opts.Repository)
	log.WithField("repository", repoURL).Debug("Installing command")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", errors.FileError("find project root", "", err)
	}

	ccmdDir := filepath.Join(projectRoot, ".claude")
	commandsDir := filepath.Join(ccmdDir, "commands")

	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return "", errors.FileError("create commands directory", commandsDir, err)
	}

	tempDir, err := os.MkdirTemp("", "ccmd-install-*")
	if err != nil {
		return "", errors.FileError("create temp directory", "", err)
	}
	defer os.RemoveAll(tempDir)

	output.PrintInfof("Cloning repository %s...", repoURL)
	cloneVersion := opts.Version
	if opts.Commit != "" {
		cloneVersion = opts.Commit
	}
	if err := gitClone(repoURL, tempDir, cloneVersion); err != nil {
		return "", errors.GitError("clone", err)
	}

	metadataPath := filepath.Join(tempDir, "ccmd.yaml")
	metadata, err := readCommandMetadata(metadataPath)
	if err != nil {
		return "", err
	}

	commandName := opts.Name
	if commandName == "" {
		commandName = metadata.Name
		if commandName == "" {
			commandName = extractCommandName(repoURL)
		}
	}

	if err := validateCommandName(commandName); err != nil {
		return "", err
	}

	targetRepoPath := ExtractRepoPath(repoURL)
	existingCommand, err := findExistingCommandByRepo(projectRoot, targetRepoPath)
	if err != nil {
		return "", errors.FileError("check existing commands", "", err)
	}

	if existingCommand != "" && !opts.Force {
		return "", errors.AlreadyExists(fmt.Sprintf(
			"repository already installed as command %q, use --force to reinstall",
			existingCommand))
	}

	commandNameChanged := existingCommand != "" && existingCommand != commandName

	if opts.Force {
		output.PrintInfof("Removing previous installation %q...", existingCommand)
		if err := removeCommandFiles(projectRoot, existingCommand); err != nil {
			return "", err
		}
	}

	destDir := filepath.Join(commandsDir, commandName)

	output.PrintInfof("Installing command %q...", commandName)
	if err := copyDirectory(tempDir, destDir); err != nil {
		return "", errors.FileError("copy command files", destDir, err)
	}

	originalVersion := metadata.Version

	metadata.Name = commandName
	metadata.Repository = repoURL

	if err := writeCommandMetadata(filepath.Join(destDir, "ccmd.yaml"), metadata); err != nil {
		os.RemoveAll(destDir)
		return "", err
	}

	standalonePath := filepath.Join(ccmdDir, "commands", commandName+".md")
	if err := createStandaloneDoc(destDir, standalonePath, metadata); err != nil {
		log.WithError(err).Warn("Failed to create standalone documentation")
	}

	if err := updateLockFile(projectRoot, commandName, metadata, originalVersion, opts.Version); err != nil {
		log.WithError(err).Warn("Failed to update lock file")
	}

	repoSpec := opts.Repository
	if strings.Contains(repoSpec, "://") || strings.HasPrefix(repoSpec, "git@") {
		repoSpec = ExtractRepoPath(repoSpec)
	}
	versionForConfig := opts.Version
	if isCommitHash(versionForConfig) && len(versionForConfig) > 7 {
		versionForConfig = versionForConfig[:7]
	}
	if err := addToConfig(projectRoot, commandName, repoSpec, versionForConfig); err != nil {
		log.WithError(err).Warn("Failed to update ccmd.yaml")
	}

	if commandNameChanged {
		output.PrintSuccessf("Installed command %q renamed to %q successfully", existingCommand, commandName)
	} else {
		output.PrintSuccessf("Command %q installed successfully", commandName)
	}

	return commandName, nil
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

	lockPath := filepath.Join(projectPath, LockFileName)
	var lockFile *LockFile
	if fileExists(lockPath) {
		lockFile, _ = ReadLockFile(lockPath)
	}

	var installErrors []error
	for _, cmdSpec := range config.Commands {
		repo, version := ParseCommandSpec(cmdSpec)

		commitToInstall := ""
		if lockFile != nil {
			normalizedRepo := NormalizeRepositoryURL(repo)
			for _, lockCmd := range lockFile.Commands {
				if NormalizeRepositoryURL(lockCmd.Source) == normalizedRepo {
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
		if _, err := Install(ctx, opts); err != nil {
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
	if metadata.Name != "" || metadata.Version != "" {
		return metadata.Validate()
	}
	return nil
}

func validateCommandName(name string) error {
	if name == "" {
		return errors.InvalidInput("command name cannot be empty")
	}

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
	indexPath := filepath.Join(commandDir, "index.md")
	if !fileExists(indexPath) {
		return errors.NotFound("index.md not found")
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

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

	var lockFile *LockFile
	if fileExists(lockPath) {
		var err error
		lockFile, err = ReadLockFile(lockPath)
		if err != nil {
			return err
		}
	} else {
		lockFile = &LockFile{
			Version:         "1.0",
			LockfileVersion: 1,
			Commands:        make(map[string]*LockCommand),
		}
	}

	commitHash := "unknown"
	commandPath := filepath.Join(projectRoot, ".claude", "commands", commandName)
	if hash, err := gitGetCurrentCommit(commandPath); err == nil {
		commitHash = hash
	}

	resolved := metadata.Repository
	if requestedVersion != "" {
		resolved = fmt.Sprintf("%s@%s", metadata.Repository, requestedVersion)
	} else {
		if defaultBranch, err := gitGetDefaultBranch(commandPath); err == nil {
			resolved = fmt.Sprintf("%s@%s", metadata.Repository, defaultBranch)
		} else if commitHash != "unknown" && len(commitHash) >= 7 {
			resolved = fmt.Sprintf("%s@%s", metadata.Repository, commitHash[:7])
		}
	}

	repoPath := ExtractRepoPath(metadata.Repository)
	var existingKey string
	var existingCmd *LockCommand

	for key, cmd := range lockFile.Commands {
		if ExtractRepoPath(cmd.Source) == repoPath {
			existingKey = key
			existingCmd = cmd
			break
		}
	}

	installedAt := now
	if existingCmd != nil && !existingCmd.InstalledAt.IsZero() {
		installedAt = existingCmd.InstalledAt
	}

	if existingKey != "" && existingKey != commandName {
		delete(lockFile.Commands, existingKey)
	}

	lockFile.Commands[commandName] = &LockCommand{
		Name:        commandName,
		Version:     originalVersion,
		Source:      metadata.Repository,
		Resolved:    resolved,
		Commit:      commitHash,
		InstalledAt: installedAt,
		UpdatedAt:   now,
	}

	return WriteLockFile(lockPath, lockFile)
}

func getInstalledCommands(projectRoot string) (map[string]string, error) {
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")
	installedCommands := make(map[string]string)

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return installedCommands, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			metadataPath := filepath.Join(commandsDir, entry.Name(), "ccmd.yaml")
			if metadata, err := readCommandMetadata(metadataPath); err == nil && metadata.Repository != "" {
				installedCommands[entry.Name()] = ExtractRepoPath(metadata.Repository)
			}
		}
	}

	return installedCommands, nil
}

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
	var config *ProjectConfig
	if ProjectConfigExists(projectRoot) {
		var err error
		config, err = LoadProjectConfig(projectRoot)
		if err != nil {
			return err
		}
	} else {
		config = &ProjectConfig{
			Commands: []string{},
		}
	}

	installedCommands, err := getInstalledCommands(projectRoot)
	if err != nil {
		return err
	}

	commandSpec := repository
	if version != "" {
		commandSpec = fmt.Sprintf("%s@%s", repository, version)
	}

	found := false
	currentRepo := ExtractRepoPath(repository)

	for i, cmd := range config.Commands {
		repo, _ := ParseCommandSpec(cmd)
		repoPath := ExtractRepoPath(repo)

		if repoPath == currentRepo {
			config.Commands[i] = commandSpec
			found = true
			break
		} else if installedRepo, exists := installedCommands[commandName]; exists && repoPath == installedRepo {
			config.Commands[i] = commandSpec
			found = true
			break
		}
	}

	if !found {
		config.Commands = append(config.Commands, commandSpec)
	}

	return SaveProjectConfig(projectRoot, config)
}

func ParseRepositorySpec(spec string) (repository, version string) {
	if strings.HasPrefix(spec, "git@") {
		gitSuffix := ".git@"
		gitIndex := strings.LastIndex(spec, gitSuffix)
		if gitIndex > -1 {
			repository = spec[:gitIndex+4]
			version = spec[gitIndex+len(gitSuffix):]
			return repository, version
		}

		colonIndex := strings.Index(spec, ":")
		if colonIndex > -1 {
			afterColon := spec[colonIndex+1:]

			firstAtIndex := strings.Index(afterColon, "@")
			if firstAtIndex > -1 {
				beforeAt := afterColon[:firstAtIndex]
				if strings.Contains(beforeAt, "/") {
					repository = spec[:colonIndex+1+firstAtIndex]
					version = spec[colonIndex+1+firstAtIndex+1:]
					return repository, version
				}
				remainingPart := afterColon[firstAtIndex+1:]
				nextAtIndex := strings.Index(remainingPart, "@")
				if nextAtIndex > -1 {
					totalIndex := colonIndex + 1 + firstAtIndex + 1 + nextAtIndex
					repository = spec[:totalIndex]
					version = spec[totalIndex+1:]
					return repository, version
				}
			}
		}

		return spec, ""
	}

	lastAtIndex := strings.LastIndex(spec, "@")
	if lastAtIndex == -1 {
		return spec, ""
	}

	repository = spec[:lastAtIndex]
	version = spec[lastAtIndex+1:]
	return repository, version
}

func NormalizeRepositoryURL(url string) string {
	if !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@") {
		if strings.Count(url, "/") == 1 {
			return fmt.Sprintf("https://github.com/%s.git", url)
		}
	}

	if !strings.HasSuffix(url, ".git") && strings.Contains(url, "github.com") {
		url += ".git"
	}

	return url
}

func extractCommandName(repoURL string) string {
	path := ExtractRepoPath(repoURL)
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if fileExists(filepath.Join(dir, "ccmd.yaml")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
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

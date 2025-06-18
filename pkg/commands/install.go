package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/validation"
)

// InstallOptions represents options for installing a command
type InstallOptions struct {
	Repository string        // Git repository URL
	Version    string        // Version/tag to install (optional)
	Name       string        // Override command name (optional)
	Force      bool          // Force reinstall if already exists
	FileSystem fs.FileSystem // File system to use (defaults to OS)
}

// GitClient interface for git operations
type GitClient interface {
	Clone(opts git.CloneOptions) error
	ValidateRemoteRepository(url string) error
	GetLatestTag(repoPath string) (string, error)
	GetCurrentCommit(repoPath string) (string, error)
}

// Install installs a command from a Git repository
func Install(opts InstallOptions) error {
	if opts.Repository == "" {
		return fmt.Errorf("repository URL is required")
	}

	// Use default file system if not provided
	if opts.FileSystem == nil {
		opts.FileSystem = fs.NewOSFileSystem()
	}

	// Get config directory (project-local)
	configDir := ".claude"

	// Create config directory if it doesn't exist
	if err := opts.FileSystem.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create commands directory
	commandsDir := filepath.Join(configDir, "commands")
	if err := opts.FileSystem.MkdirAll(commandsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	// Parse repository URL to get default command name
	repoName, err := git.ParseRepositoryURL(opts.Repository)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Create git client
	gitClient := git.NewClient(configDir)

	// Validate remote repository exists
	if err := gitClient.ValidateRemoteRepository(opts.Repository); err != nil {
		return fmt.Errorf("repository not accessible: %w", err)
	}

	// Create temporary directory for cloning
	tempDir := filepath.Join(configDir, "tmp", fmt.Sprintf("install-%s-%d", repoName, time.Now().Unix()))
	if err := opts.FileSystem.MkdirAll(tempDir, 0o755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if removeErr := opts.FileSystem.RemoveAll(tempDir); removeErr != nil {
			// Log error but don't fail the operation
			_ = removeErr
		}
	}()

	// Clone repository
	cloneOpts := git.CloneOptions{
		URL:     opts.Repository,
		Target:  tempDir,
		Tag:     opts.Version,
		Shallow: true,
		Depth:   1,
	}
	if err := gitClient.Clone(cloneOpts); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get version if not specified
	version := opts.Version
	if version == "" {
		// Try to get latest tag
		latestTag, err := gitClient.GetLatestTag(tempDir)
		if err != nil {
			// Fall back to current commit
			commit, err := gitClient.GetCurrentCommit(tempDir)
			if err != nil {
				return fmt.Errorf("failed to determine version: %w", err)
			}
			version = commit[:7] // Use short commit hash
		} else {
			version = latestTag
		}
	}

	// Read and validate ccmd.yaml
	metadataPath := filepath.Join(tempDir, "ccmd.yaml")
	metadataData, err := opts.FileSystem.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read ccmd.yaml: %w", err)
	}

	var metadata models.CommandMetadata
	if err := metadata.UnmarshalYAML(metadataData); err != nil {
		return fmt.Errorf("failed to parse ccmd.yaml: %w", err)
	}

	if err := metadata.Validate(); err != nil {
		return fmt.Errorf("invalid ccmd.yaml: %w", err)
	}

	// Use override name if provided
	commandName := metadata.Name
	if opts.Name != "" {
		commandName = opts.Name
	}

	// Check if command already exists
	commandDir := filepath.Join(commandsDir, commandName)
	exists, existsErr := opts.FileSystem.Exists(commandDir)
	if existsErr != nil {
		return fmt.Errorf("failed to check if command exists: %w", existsErr)
	}
	if exists && !opts.Force {
		return fmt.Errorf("command '%s' already exists (use --force to reinstall)", commandName)
	}

	// Validate command structure
	validator := validation.NewValidator(opts.FileSystem)
	if err := validator.ValidateCommandStructure(tempDir); err != nil {
		return fmt.Errorf("invalid command structure: %w", err)
	}

	// Remove existing command if force install
	if exists && opts.Force {
		if err := opts.FileSystem.RemoveAll(commandDir); err != nil {
			return fmt.Errorf("failed to remove existing command: %w", err)
		}
	}

	// Create command directory
	if err := opts.FileSystem.MkdirAll(commandDir, 0o755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Copy files to command directory
	if err := copyDirectory(opts.FileSystem, tempDir, commandDir); err != nil {
		if removeErr := opts.FileSystem.RemoveAll(commandDir); removeErr != nil {
			// Log error but don't fail the operation
			_ = removeErr
		}
		return fmt.Errorf("failed to install command: %w", err)
	}

	// Create standalone .md file for dual structure
	standalonePath := filepath.Join(commandsDir, fmt.Sprintf("%s.md", commandName))
	indexPath := filepath.Join(commandDir, "index.md")
	if indexData, err := opts.FileSystem.ReadFile(indexPath); err == nil {
		if err := opts.FileSystem.WriteFile(standalonePath, indexData, 0o644); err != nil {
			if removeErr := opts.FileSystem.RemoveAll(commandDir); removeErr != nil {
				// Log error but don't fail the operation
				_ = removeErr
			}
			return fmt.Errorf("failed to create standalone file: %w", err)
		}
	}

	// Update lock file
	lockManager := lock.NewManagerWithFS(configDir, opts.FileSystem)
	if err := lockManager.Load(); err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Add command to lock file
	cmd := &models.Command{
		Name:        commandName,
		Version:     version,
		Source:      opts.Repository,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]string{
			"description": metadata.Description,
			"author":      metadata.Author,
		},
	}

	if err := lockManager.AddCommand(cmd); err != nil {
		return fmt.Errorf("failed to add command to lock file: %w", err)
	}

	// Save lock file
	if err := lockManager.Save(); err != nil {
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(fs fs.FileSystem, src, dst string) error {
	// Get source directory info
	srcInfo, err := fs.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory with same permissions
	if err := fs.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Read source directory
	entries, err := fs.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDirectory(fs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := fs.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", srcPath, err)
			}

			info, err := fs.Stat(srcPath)
			if err != nil {
				return fmt.Errorf("failed to stat file %s: %w", srcPath, err)
			}

			if err := fs.WriteFile(dstPath, data, info.Mode()); err != nil {
				return fmt.Errorf("failed to write file %s: %w", dstPath, err)
			}
		}
	}

	return nil
}

// ParseRepositorySpec parses a repository specification (URL[@version])
func ParseRepositorySpec(spec string) (repository, version string) {
	// Find the last @ that could be a version separator
	lastAt := strings.LastIndex(spec, "@")

	// If no @ found, it's all repository
	if lastAt == -1 {
		return spec, ""
	}

	// Check if what follows @ looks like a version (starts with v, digit, or is a short hash)
	possibleVersion := spec[lastAt+1:]
	if possibleVersion != "" && (strings.HasPrefix(possibleVersion, "v") ||
		(len(possibleVersion) >= 1 && possibleVersion[0] >= '0' && possibleVersion[0] <= '9') ||
		(len(possibleVersion) >= 6 && len(possibleVersion) <= 40 && isHex(possibleVersion))) {
		return spec[:lastAt], possibleVersion
	}

	// Otherwise, the @ is part of the repository name
	return spec, ""
}

// isHex checks if a string contains only hexadecimal characters
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

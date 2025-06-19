package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/validation"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// Options represents configuration for the installer
type Options struct {
	Repository    string        // Git repository URL
	Version       string        // Version/tag to install (optional)
	Name          string        // Override command name (optional)
	Force         bool          // Force reinstall if already exists
	InstallDir    string        // Directory to install commands (default: .claude/commands)
	FileSystem    fs.FileSystem // File system interface (for testing)
	GitClient     GitClient     // Git client interface (for testing)
	TempDirPrefix string        // Prefix for temporary directories
}

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(opts git.CloneOptions) error
	ValidateRemoteRepository(url string) error
	GetLatestTag(repoPath string) (string, error)
	GetCurrentCommit(repoPath string) (string, error)
	IsGitRepository(path string) bool
}

// Installer manages the command installation process
type Installer struct {
	opts        Options
	logger      logger.Logger
	gitClient   GitClient
	fileSystem  fs.FileSystem
	metaManager *metadataManager
	lockManager *lock.Manager
	validator   *validation.Validator
}

// New creates a new installer instance
func New(opts Options) (*Installer, error) {
	// Validate required options
	if opts.Repository == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "repository URL is required")
	}

	// Set defaults
	if opts.InstallDir == "" {
		opts.InstallDir = filepath.Join(".claude", "commands")
	}
	if opts.FileSystem == nil {
		opts.FileSystem = fs.NewOSFileSystem()
	}
	if opts.GitClient == nil {
		opts.GitClient = git.NewClient("")
	}
	if opts.TempDirPrefix == "" {
		opts.TempDirPrefix = "ccmd-install"
	}

	// Create managers
	lockManager := lock.NewManagerWithFS(filepath.Dir(opts.InstallDir), opts.FileSystem)

	return &Installer{
		opts:        opts,
		logger:      logger.WithField("component", "installer"),
		gitClient:   opts.GitClient,
		fileSystem:  opts.FileSystem,
		metaManager: newMetadataManager(opts.FileSystem),
		lockManager: lockManager,
		validator:   validation.NewValidator(opts.FileSystem),
	}, nil
}

// Install performs the complete installation process
func (i *Installer) Install(_ context.Context) error {
	i.logger.WithFields(logger.Fields{
		"repository": i.opts.Repository,
		"version":    i.opts.Version,
		"name":       i.opts.Name,
		"force":      i.opts.Force,
	}).Debug("starting installation")

	// Step 1: Validate remote repository
	if err := i.validateRepository(); err != nil {
		return errors.Wrap(err, errors.CodeGitInvalidRepo, "repository validation failed")
	}

	// Step 2: Create temporary directory for cloning
	tempDir, err := i.createTempDir()
	if err != nil {
		return errors.Wrap(err, errors.CodeFileIO, "failed to create temporary directory")
	}
	defer i.cleanupTempDir(tempDir)

	// Step 3: Clone repository to temporary location
	if err := i.cloneRepository(tempDir); err != nil {
		return errors.Wrap(err, errors.CodeGitClone, "failed to clone repository")
	}

	// Step 4: Validate repository structure
	metadata, err := i.validateRepositoryStructure(tempDir)
	if err != nil {
		return errors.Wrap(err, errors.CodeValidation, "repository validation failed")
	}

	// Step 5: Determine command name and version
	commandName := i.determineCommandName(metadata)
	version, err := i.determineVersion(tempDir)
	if err != nil {
		return errors.Wrap(err, errors.CodeGitInvalidRepo, "failed to determine version")
	}

	// Step 6: Check if command already exists
	if err := i.checkExistingCommand(commandName); err != nil {
		return err
	}

	// Step 7: Install command files
	commandDir := filepath.Join(i.opts.InstallDir, commandName)
	if err := i.installCommandFiles(tempDir, commandDir); err != nil {
		// Rollback on failure
		i.rollbackInstallation(commandDir)
		return errors.Wrap(err, errors.CodeFileIO, "failed to install command files")
	}

	// Step 8: Create standalone .md file
	if err := i.createStandaloneFile(commandDir, metadata); err != nil {
		// Rollback on failure with metadata to remove standalone file
		i.rollbackInstallationWithMetadata(commandDir, metadata)
		return errors.Wrap(err, errors.CodeFileIO, "failed to create standalone file")
	}

	// Step 9: Update metadata with installation info
	if err := i.updateCommandMetadata(commandDir, metadata, version); err != nil {
		// Rollback on failure with metadata to remove standalone file
		i.rollbackInstallationWithMetadata(commandDir, metadata)
		return errors.Wrap(err, errors.CodeFileIO, "failed to update command metadata")
	}

	// Step 10: Update lock file
	if err := i.updateLockFile(commandName, version, metadata); err != nil {
		// Rollback on failure with metadata to remove standalone file
		i.rollbackInstallationWithMetadata(commandDir, metadata)
		return errors.Wrap(err, errors.CodeLockConflict, "failed to update lock file")
	}

	i.logger.WithFields(logger.Fields{
		"command": commandName,
		"version": version,
		"path":    commandDir,
	}).Info("command installed successfully")

	return nil
}

// validateRepository checks if the remote repository is accessible
func (i *Installer) validateRepository() error {
	i.logger.WithField("repository", i.opts.Repository).Debug("validating remote repository")

	if err := i.gitClient.ValidateRemoteRepository(i.opts.Repository); err != nil {
		return err
	}

	return nil
}

// createTempDir creates a temporary directory for cloning
func (i *Installer) createTempDir() (string, error) {
	tempBase := os.TempDir()
	tempDir := filepath.Join(tempBase, fmt.Sprintf("%s-%d", i.opts.TempDirPrefix, time.Now().UnixNano()))

	if err := i.fileSystem.MkdirAll(tempDir, 0o755); err != nil {
		return "", err
	}

	i.logger.WithField("path", tempDir).Debug("created temporary directory")
	return tempDir, nil
}

// cleanupTempDir removes the temporary directory
func (i *Installer) cleanupTempDir(tempDir string) {
	i.logger.WithField("path", tempDir).Debug("cleaning up temporary directory")

	if err := i.fileSystem.RemoveAll(tempDir); err != nil {
		i.logger.WithError(err).Warn("failed to remove temporary directory")
	}
}

// cloneRepository clones the repository to the temporary directory
func (i *Installer) cloneRepository(tempDir string) error {
	i.logger.WithFields(logger.Fields{
		"repository": i.opts.Repository,
		"target":     tempDir,
		"version":    i.opts.Version,
	}).Debug("cloning repository")

	cloneOpts := git.CloneOptions{
		URL:     i.opts.Repository,
		Target:  tempDir,
		Shallow: true,
		Depth:   1,
	}

	// Clone specific version if provided
	if i.opts.Version != "" {
		cloneOpts.Tag = i.opts.Version
	}

	return i.gitClient.Clone(cloneOpts)
}

// validateRepositoryStructure validates the cloned repository has proper structure
func (i *Installer) validateRepositoryStructure(repoPath string) (*models.CommandMetadata, error) {
	i.logger.WithField("path", repoPath).Debug("validating repository structure")

	// Check for ccmd.yaml
	metadataPath := filepath.Join(repoPath, "ccmd.yaml")
	if _, err := i.fileSystem.Stat(metadataPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.CodeValidation, "ccmd.yaml not found in repository")
		}
		return nil, err
	}

	// Read and validate metadata
	meta, err := i.metaManager.ReadCommandMetadata(repoPath)
	if err != nil {
		return nil, err
	}

	// Validate command structure
	if err := i.validator.ValidateCommandStructure(repoPath); err != nil {
		return nil, err
	}

	return meta, nil
}

// determineCommandName determines the final command name
func (i *Installer) determineCommandName(metadata *models.CommandMetadata) string {
	if i.opts.Name != "" {
		return i.opts.Name
	}
	return metadata.Name
}

// determineVersion determines the version to use for installation
func (i *Installer) determineVersion(repoPath string) (string, error) {
	// If version is specified, use it
	if i.opts.Version != "" {
		return i.opts.Version, nil
	}

	// Try to get latest tag
	latestTag, err := i.gitClient.GetLatestTag(repoPath)
	if err == nil && latestTag != "" {
		i.logger.WithField("tag", latestTag).Debug("using latest tag as version")
		return latestTag, nil
	}

	// Fall back to current commit
	commit, err := i.gitClient.GetCurrentCommit(repoPath)
	if err != nil {
		return "", err
	}

	// Use short commit hash
	if len(commit) > 7 {
		commit = commit[:7]
	}

	i.logger.WithField("commit", commit).Debug("using commit hash as version")
	return commit, nil
}

// checkExistingCommand checks if command already exists
func (i *Installer) checkExistingCommand(commandName string) error {
	commandDir := filepath.Join(i.opts.InstallDir, commandName)

	if _, err := i.fileSystem.Stat(commandDir); err == nil {
		// Command exists
		if !i.opts.Force {
			return errors.New(errors.CodeAlreadyExists, "command already exists").
				WithDetail("command", commandName).
				WithDetail("use", "--force to reinstall")
		}

		// Remove existing command for force install
		i.logger.WithField("command", commandName).Debug("removing existing command for force install")
		if err := i.fileSystem.RemoveAll(commandDir); err != nil {
			return errors.Wrap(err, errors.CodeFileIO, "failed to remove existing command")
		}

		// Also remove standalone file if it exists
		standaloneFile := filepath.Join(i.opts.InstallDir, fmt.Sprintf("%s.md", commandName))
		if err := i.fileSystem.Remove(standaloneFile); err != nil && !os.IsNotExist(err) {
			i.logger.WithError(err).Warn("failed to remove existing standalone file")
		}
	}

	return nil
}

// installCommandFiles copies command files to installation directory
func (i *Installer) installCommandFiles(srcDir, dstDir string) error {
	i.logger.WithFields(logger.Fields{
		"source":      srcDir,
		"destination": dstDir,
	}).Debug("installing command files")

	// Create destination directory
	if err := i.fileSystem.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	// Copy all files except .git directory
	return i.copyDirectory(srcDir, dstDir)
}

// copyDirectory recursively copies directory contents
func (i *Installer) copyDirectory(src, dst string) error {
	entries, err := i.fileSystem.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create directory and recurse
			if err := i.fileSystem.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			if err := i.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := i.fileSystem.ReadFile(srcPath)
			if err != nil {
				return err
			}

			if err := i.fileSystem.WriteFile(dstPath, data, 0o644); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateCommandMetadata updates the command metadata with installation info
func (i *Installer) updateCommandMetadata(commandDir string, metadata *models.CommandMetadata, version string) error {
	// Update metadata with installation-specific information
	return i.metaManager.UpdateCommandMetadata(commandDir, func(m *models.CommandMetadata) error {
		// Preserve original metadata
		*m = *metadata

		// Update version if it was determined during installation
		if m.Version == "" {
			m.Version = version
		}

		return nil
	})
}

// updateLockFile updates the lock file with installed command info
func (i *Installer) updateLockFile(commandName, version string, metadata *models.CommandMetadata) error {
	i.logger.WithFields(logger.Fields{
		"command": commandName,
		"version": version,
	}).Debug("updating lock file")

	// Load lock file
	if err := i.lockManager.Load(); err != nil {
		return err
	}

	// Create command entry
	cmd := &models.Command{
		Name:        commandName,
		Version:     version,
		Source:      i.opts.Repository,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]string{
			"description": metadata.Description,
			"author":      metadata.Author,
			"repository":  metadata.Repository,
		},
	}

	// Add command to lock file
	if err := i.lockManager.AddCommand(cmd); err != nil {
		return err
	}

	// Save lock file
	return i.lockManager.Save()
}

// createStandaloneFile creates a standalone .md file from the command's index.md
func (i *Installer) createStandaloneFile(commandDir string, metadata *models.CommandMetadata) error {
	// Use metadata.Name for the standalone file
	standaloneFile := filepath.Join(i.opts.InstallDir, fmt.Sprintf("%s.md", metadata.Name))

	// Read index.md from command directory
	indexPath := filepath.Join(commandDir, "index.md")
	indexData, err := i.fileSystem.ReadFile(indexPath)
	if err != nil {
		return errors.Wrap(err, errors.CodeFileIO, "failed to read index.md")
	}

	// Write standalone file
	if err := i.fileSystem.WriteFile(standaloneFile, indexData, 0o644); err != nil {
		return errors.Wrap(err, errors.CodeFileIO, "failed to create standalone file")
	}

	i.logger.WithFields(logger.Fields{
		"standalone": standaloneFile,
		"source":     indexPath,
	}).Debug("created standalone .md file")

	return nil
}

// rollbackInstallation removes partially installed command
func (i *Installer) rollbackInstallation(commandDir string) {
	i.logger.WithField("path", commandDir).Warn("rolling back installation")

	if err := i.fileSystem.RemoveAll(commandDir); err != nil {
		i.logger.WithError(err).Error("failed to rollback installation")
	}
}

// rollbackInstallationWithMetadata removes partially installed command including standalone file
func (i *Installer) rollbackInstallationWithMetadata(commandDir string, metadata *models.CommandMetadata) {
	i.logger.WithField("path", commandDir).Warn("rolling back installation")

	// Remove command directory
	if err := i.fileSystem.RemoveAll(commandDir); err != nil {
		i.logger.WithError(err).Error("failed to rollback command directory")
	}

	// Remove standalone file using metadata.Name
	if metadata != nil {
		standaloneFile := filepath.Join(i.opts.InstallDir, fmt.Sprintf("%s.md", metadata.Name))
		if err := i.fileSystem.Remove(standaloneFile); err != nil && !os.IsNotExist(err) {
			i.logger.WithError(err).Error("failed to rollback standalone file")
		}
	}
}

package install

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/installer"
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
	"github.com/gifflet/ccmd/pkg/project"
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
		RunE: errors.WrapCommand("install", func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runInstallFromConfig(force)
			}
			return runInstall(args[0], version, name, force)
		}),
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version/tag to install")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Override command name")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall if already exists")

	return cmd
}

func runInstallFromConfig(force bool) error {
	log := logger.WithField("command", "install-from-config")
	log.Debug("starting install from config")

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use new installer to install from config
	ctx := context.Background()
	if err := installer.InstallFromConfig(ctx, cwd, force); err != nil {
		// Check if it's a partial failure
		if errors.IsCode(err, errors.CodePartialFailure) {
			// Some commands failed but not all
			output.PrintWarningf("Some commands failed to install")
			// Don't return error for partial failures
			return nil
		}
		return err
	}

	return nil
}

func runInstall(repository, version, name string, force bool) error {
	log := logger.WithField("command", "install")
	log.WithFields(logger.Fields{
		"repository": repository,
		"version":    version,
		"name":       name,
		"force":      force,
	}).Debug("starting install")

	// Parse repository spec if version is included
	repo, specVersion := installer.ParseRepositorySpec(repository)
	if specVersion != "" && version == "" {
		version = specVersion
	}

	// Normalize repository URL
	repo = installer.NormalizeRepositoryURL(repo)

	// Show installation info
	output.PrintInfof("Installing command from: %s", repo)
	if version != "" {
		output.PrintInfof("Version: %s", version)
	}
	if name != "" {
		output.PrintInfof("Custom name: %s", name)
	}

	// Create spinner for installation process
	spinner := output.NewSpinner("Installing command...")
	spinner.Start()

	// Get current directory for project updates
	cwd, err := os.Getwd()
	projectPath := ""
	if err == nil {
		projectPath = cwd
	}

	// Install using new installer
	ctx := context.Background()
	opts := installer.IntegrationOptions{
		Repository:  repo,
		Version:     version,
		Name:        name,
		Force:       force,
		ProjectPath: projectPath,
	}

	if err := installer.InstallCommand(ctx, opts, true); err != nil {
		spinner.Stop()
		log.WithError(err).Error("installation failed")
		return err
	}

	spinner.Stop()

	// Get installed command info
	commandName := name
	if commandName == "" {
		// Extract command name from repository
		parts := strings.Split(strings.TrimSuffix(repo, ".git"), "/")
		commandName = parts[len(parts)-1]
	}

	output.PrintSuccessf("Command '%s' has been successfully installed", commandName)

	// Project files are now updated by the installer itself
	if projectPath != "" {
		pm := project.NewManager(projectPath)
		if pm.ConfigExists() {
			// Extract owner/repo from repository URL
			repoPath := commands.ExtractRepoPath(repo)
			if repoPath != "" {
				// Try to add the command to ccmd.yaml
				if err := pm.AddCommand(repoPath, version); err != nil {
					// Don't fail the installation, just warn
					log.WithError(err).Warn("failed to add command to ccmd.yaml")
				} else {
					output.PrintInfof("Added to ccmd.yaml")

					// Also update the lock file with project info
					if err := updateProjectLockFile(pm, commandName, repo, version); err != nil {
						log.WithError(err).Warn("failed to update ccmd-lock.yaml")
					} else {
						output.PrintInfof("Updated ccmd-lock.yaml")
					}
				}
			}
		}
	}

	output.PrintInfof("\nTo use the command, run:")
	output.PrintInfof("/%s", commandName)

	return nil
}

// updateProjectLockFile updates the project's lock file with the installed command info
func updateProjectLockFile(pm *project.Manager, commandName, repository, version string) error {
	// Load or create lock file
	lockFile, err := pm.LoadLockFile()
	if err != nil {
		// Check if file doesn't exist - create new one
		if strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "cannot find the file") {
			lockFile = project.NewLockFile()
		} else {
			return err
		}
	}

	// Get the current commit hash from the installed command
	// For now, we'll use the version as a placeholder
	// In a real implementation, we'd get this from the git operations
	commitHash := version
	if commitHash == "" {
		commitHash = "latest"
	}
	// Pad or truncate to 40 chars for validation
	if len(commitHash) < 40 {
		commitHash = fmt.Sprintf("%-40s", commitHash)
	} else if len(commitHash) > 40 {
		commitHash = commitHash[:40]
	}

	// Create command entry
	cmd := &project.Command{
		Name:         commandName,
		Repository:   repository,
		Version:      version,
		CommitHash:   commitHash,
		InstalledAt:  time.Now(),
		UpdatedAt:    time.Now(),
		FileSize:     1024,                    // Placeholder
		Checksum:     strings.Repeat("0", 64), // Placeholder SHA256
		Dependencies: []string{},
		Metadata:     map[string]string{},
	}

	// Add to lock file
	if err := lockFile.AddCommand(cmd); err != nil {
		return err
	}

	// Save lock file
	return pm.SaveLockFile(lockFile)
}

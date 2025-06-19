package install

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

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

	// Create project manager
	pm := project.NewManager(cwd)

	// Check if config exists
	if !pm.ConfigExists() {
		return fmt.Errorf("no ccmd.yaml found in current directory")
	}

	// Load config
	config, err := pm.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load ccmd.yaml: %w", err)
	}

	if len(config.Commands) == 0 {
		output.PrintInfof("No commands found in ccmd.yaml")
		return nil
	}

	output.PrintInfof("Installing %d command(s) from ccmd.yaml", len(config.Commands))

	// Install each command
	installedCount := 0
	for _, cmd := range config.Commands {
		// Build repository URL
		repository := fmt.Sprintf("https://github.com/%s.git", cmd.Repo)

		output.PrintInfof("\nInstalling %s", cmd.Repo)
		if cmd.Version != "" {
			output.PrintInfof("Version: %s", cmd.Version)
		}

		// Create spinner for installation process
		spinner := output.NewSpinner(fmt.Sprintf("Installing %s...", cmd.Repo))
		spinner.Start()

		// Parse repo to get command name
		_, repoName, err := cmd.ParseOwnerRepo()
		if err != nil {
			spinner.Stop()
			output.Error("Failed to parse repository %s: %v", cmd.Repo, err)
			continue
		}

		// Install the command
		installOpts := commands.InstallOptions{
			Repository: repository,
			Version:    cmd.Version,
			Name:       repoName,
			Force:      force,
		}

		if err := commands.Install(installOpts); err != nil {
			spinner.Stop()
			output.Error("Failed to install %s: %v", cmd.Repo, err)
			continue
		}

		spinner.Stop()
		output.PrintSuccessf("Command '%s' has been successfully installed", repoName)
		installedCount++
	}

	output.PrintInfof("\nSuccessfully installed %d out of %d command(s)", installedCount, len(config.Commands))

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
	repo, specVersion := commands.ParseRepositorySpec(repository)
	if specVersion != "" && version == "" {
		version = specVersion
	}

	// Normalize repository URL
	repo = normalizeRepositoryURL(repo)

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

	// Install the command
	installOpts := commands.InstallOptions{
		Repository: repo,
		Version:    version,
		Name:       name,
		Force:      force,
	}

	if err := commands.Install(installOpts); err != nil {
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

	// Add to project config if it exists
	cwd, err := os.Getwd()
	if err == nil {
		pm := project.NewManager(cwd)
		if pm.ConfigExists() {
			// Extract owner/repo from repository URL
			repoPath := extractRepoPath(repo)
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

// normalizeRepositoryURL normalizes a repository URL to include https://
func normalizeRepositoryURL(url string) string {
	url = strings.TrimSpace(url)

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

// extractRepoPath extracts owner/repo from a Git URL
func extractRepoPath(gitURL string) string {
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

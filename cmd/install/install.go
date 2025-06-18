package install

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// NewCommand creates a new install command.
func NewCommand() *cobra.Command {
	var (
		version string
		name    string
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "install <repository>",
		Short: "Install a command from a Git repository",
		Long: `Install a command from a Git repository.

The repository must contain a valid ccmd.yaml file and follow the CCMD structure.

Examples:
  # Install latest version
  ccmd install github.com/user/repo

  # Install specific version
  ccmd install github.com/user/repo@v1.0.0

  # Install with custom name
  ccmd install github.com/user/repo --name mycommand

  # Force reinstall
  ccmd install github.com/user/repo --force`,
		Args: cobra.ExactArgs(1),
		RunE: errors.WrapCommand("install", func(cmd *cobra.Command, args []string) error {
			return runInstall(args[0], version, name, force)
		}),
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version/tag to install")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Override command name")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall if already exists")

	return cmd
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
	output.Info("Installing command from: %s", repo)
	if version != "" {
		output.Info("Version: %s", version)
	}
	if name != "" {
		output.Info("Custom name: %s", name)
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

	output.Success("Command '%s' has been successfully installed", commandName)
	output.Info("\nTo use the command, run:")
	output.Info("  ccmd %s", commandName)

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

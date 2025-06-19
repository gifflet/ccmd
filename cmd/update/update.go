package update

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
)

var (
	// ErrCommandNameRequired is returned when no command name is provided and --all is not set.
	ErrCommandNameRequired = errors.New("command name required (or use --all to update all commands)")
	// ErrCannotSpecifyWithAll is returned when a command name is provided with --all flag.
	ErrCannotSpecifyWithAll = errors.New("cannot specify command name with --all flag")
	// ErrSomeUpdatesFailed is returned when some updates fail in batch update.
	ErrSomeUpdatesFailed = errors.New("some updates failed")
)

// Result represents the result of an update operation.
type Result struct {
	Error          error
	Name           string
	CurrentVersion string
	NewVersion     string
	Updated        bool
}

// NewCommand creates a new update command.
func NewCommand() *cobra.Command {
	var updateAll bool

	cmd := &cobra.Command{
		Use:   "update [command-name]",
		Short: "Update installed commands to their latest versions",
		Long: `Update installed commands to their latest versions.

Without arguments, it updates the specified command.
With --all flag, it updates all installed commands.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 && !updateAll {
				return ErrCommandNameRequired
			}

			if len(args) > 0 && updateAll {
				return ErrCannotSpecifyWithAll
			}

			return runUpdate(args, updateAll)
		},
	}

	cmd.Flags().BoolVar(&updateAll, "all", false, "Update all installed commands")

	return cmd
}

func runUpdate(args []string, updateAll bool) error {
	return runUpdateWithFS(args, updateAll, nil)
}

func runUpdateWithFS(args []string, updateAll bool, filesystem fs.FileSystem) error {
	if filesystem == nil {
		filesystem = fs.OS{}
	}

	// Validate arguments
	if len(args) == 0 && !updateAll {
		return ErrCommandNameRequired
	}

	if len(args) > 0 && updateAll {
		return ErrCannotSpecifyWithAll
	}

	// Get config directory (project-local)
	baseDir := ".claude"

	if updateAll {
		return updateAllCommands(baseDir, filesystem)
	}

	// Update single command
	commandName := args[0]
	result := updateCommand(commandName, baseDir, filesystem)

	if result.Error != nil {
		output.PrintErrorf("Failed to update %s: %v", commandName, result.Error)
		return result.Error
	}

	displayResult(result)
	return nil
}

func updateAllCommands(baseDir string, filesystem fs.FileSystem) error {
	// List all installed commands
	listOpts := commands.ListOptions{
		BaseDir:    baseDir,
		FileSystem: filesystem,
	}

	commandList, err := commands.List(listOpts)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(commandList) == 0 {
		output.PrintInfof("No commands installed")
		return nil
	}

	// Show progress
	output.PrintInfof("Checking %d commands for updates...", len(commandList))
	fmt.Println()

	// Update each command
	results := make([]Result, 0, len(commandList))
	for _, cmd := range commandList {
		// Only show spinner in non-test environments
		var spinner *output.Spinner
		if os.Getenv("GO_TEST") != "1" {
			spinner = output.NewSpinner(fmt.Sprintf("Checking %s...", cmd.Name))
			spinner.Start()
		}

		result := updateCommand(cmd.Name, baseDir, filesystem)
		results = append(results, result)

		if spinner != nil {
			spinner.Stop()
		}

		switch {
		case result.Error != nil:
			output.PrintErrorf("Failed to update %s: %v", cmd.Name, result.Error)
		case result.Updated:
			output.PrintSuccessf("Updated %s from %s to %s", cmd.Name, result.CurrentVersion, result.NewVersion)
		default:
			output.PrintInfof("%s is already up to date (%s)", cmd.Name, result.CurrentVersion)
		}
	}

	// Summary
	fmt.Println()
	output.PrintInfof("=== Update Summary ===")

	updatedCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Error != nil {
			failedCount++
		} else if result.Updated {
			updatedCount++
		}
	}

	if updatedCount > 0 {
		output.PrintSuccessf("%d command(s) updated", updatedCount)
	}
	if failedCount > 0 {
		output.PrintErrorf("%d command(s) failed to update", failedCount)
		return ErrSomeUpdatesFailed
	}
	if updatedCount == 0 && failedCount == 0 {
		output.PrintInfof("All commands are up to date")
	}

	return nil
}

func updateCommand(commandName, baseDir string, filesystem fs.FileSystem) Result {
	result := Result{
		Name: commandName,
	}

	// Check if command exists
	exists, err := commands.CommandExists(commandName, baseDir, filesystem)
	if err != nil {
		result.Error = fmt.Errorf("failed to check command existence: %w", err)
		return result
	}

	if !exists {
		result.Error = fmt.Errorf("command '%s' is not installed", commandName)
		return result
	}

	// Get current command info
	cmdInfo, err := commands.GetCommandInfo(commandName, baseDir, filesystem)
	if err != nil {
		result.Error = fmt.Errorf("failed to get command info: %w", err)
		return result
	}

	result.CurrentVersion = cmdInfo.Version

	// Create git client
	gitClient := git.NewClient(baseDir)

	// Create temporary directory for checking updates
	tempDir := filepath.Join(baseDir, "tmp", fmt.Sprintf("update-%s-%d", commandName, time.Now().Unix()))
	if mkdirErr := filesystem.MkdirAll(tempDir, 0o755); mkdirErr != nil {
		result.Error = fmt.Errorf("failed to create temp directory: %w", mkdirErr)
		return result
	}
	defer func() {
		if removeErr := filesystem.RemoveAll(tempDir); removeErr != nil {
			// Log error but don't fail the operation
			_ = removeErr
		}
	}()

	// Clone repository to check latest version
	cloneOpts := git.CloneOptions{
		URL:     cmdInfo.Source,
		Target:  tempDir,
		Shallow: true,
		Depth:   1,
	}
	if cloneErr := gitClient.Clone(cloneOpts); cloneErr != nil {
		result.Error = fmt.Errorf("failed to clone repository: %w", cloneErr)
		return result
	}

	// Get latest version
	latestVersion, err := getLatestVersion(gitClient, tempDir)
	if err != nil {
		result.Error = fmt.Errorf("failed to get latest version: %w", err)
		return result
	}

	// Compare versions
	needsUpdate, err := versionNeedsUpdate(cmdInfo.Version, latestVersion)
	if err != nil {
		// If we can't compare versions, check if they're different
		needsUpdate = cmdInfo.Version != latestVersion
	}

	if !needsUpdate {
		result.NewVersion = latestVersion
		result.Updated = false
		return result
	}

	// Perform update
	result.NewVersion = latestVersion
	if err := performUpdate(cmdInfo, latestVersion, baseDir, filesystem); err != nil {
		result.Error = fmt.Errorf("failed to update command: %w", err)
		return result
	}

	result.Updated = true
	return result
}

func getLatestVersion(gitClient *git.Client, repoPath string) (string, error) {
	// Try to get latest tag first
	latestTag, err := gitClient.GetLatestTag(repoPath)
	if err == nil && latestTag != "" {
		return latestTag, nil
	}

	// Fall back to current commit
	commit, err := gitClient.GetCurrentCommit(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to determine version: %w", err)
	}

	// Use short commit hash
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return commit, nil
}

func versionNeedsUpdate(current, latest string) (bool, error) {
	// If versions are identical, no update needed
	if current == latest {
		return false, nil
	}

	// Try to parse as semantic versions
	currentSemver, currentErr := semver.NewVersion(current)
	latestSemver, latestErr := semver.NewVersion(latest)

	// If both parse successfully, compare them
	if currentErr == nil && latestErr == nil {
		return latestSemver.GreaterThan(currentSemver), nil
	}

	// If one is a semver and the other isn't, prefer the semver
	if currentErr == nil && latestErr != nil {
		// Current is semver, latest is not (probably a commit hash)
		// Generally, a tagged version is preferred over a commit
		return false, nil
	}
	if currentErr != nil && latestErr == nil {
		// Current is not semver, latest is
		// Update to the tagged version
		return true, nil
	}

	// Neither is a valid semver, return error to let caller decide
	return false, fmt.Errorf("cannot compare versions: %s vs %s", current, latest)
}

func performUpdate(cmdInfo *models.Command, newVersion, baseDir string, filesystem fs.FileSystem) error {
	// Use the install command with force flag to update
	installOpts := commands.InstallOptions{
		Repository: cmdInfo.Source,
		Version:    newVersion,
		Name:       cmdInfo.Name,
		Force:      true,
	}

	if err := commands.Install(installOpts); err != nil {
		return err
	}

	// Update the lock file with new update time
	lockManager := lock.NewManagerWithFS(baseDir, filesystem)
	if err := lockManager.Load(); err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Update the UpdatedAt timestamp
	if err := lockManager.UpdateCommand(cmdInfo.Name, func(cmd *models.Command) error {
		cmd.UpdatedAt = time.Now()
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update command in lock: %w", err)
	}

	// Save the lock file
	if err := lockManager.Save(); err != nil {
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	return nil
}

func displayResult(result Result) {
	if result.Updated {
		fmt.Println()
		output.PrintSuccessf("Successfully updated %s", result.Name)
		fmt.Printf("%s %s → %s\n",
			color.CyanString("Version:"),
			color.YellowString(result.CurrentVersion),
			color.GreenString(result.NewVersion))
	} else {
		output.PrintInfof("%s is already up to date", result.Name)
		fmt.Printf("%s %s\n",
			color.CyanString("Current version:"),
			color.GreenString(result.CurrentVersion))
	}
}

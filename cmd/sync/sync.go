package sync

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
	"github.com/gifflet/ccmd/pkg/project"
)

// Result contains the analysis results of what needs to be synced.
type Result struct {
	ToInstall []string
	ToRemove  []string
	Errors    []error
}

// NewCommand creates a new sync command.
func NewCommand() *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize installed commands with ccmd.yaml",
		Long: `Synchronize installed commands with ccmd.yaml configuration.

This command reads ccmd.yaml to understand declared dependencies, compares with
currently installed commands, installs missing commands, and removes commands
not declared in ccmd.yaml (with confirmation).

The sync operation will:
- Install all commands declared in ccmd.yaml that are not currently installed
- Remove installed commands that are not declared in ccmd.yaml (with confirmation)
- Update ccmd-lock.yaml with the current state after sync

Examples:
  # Sync commands with ccmd.yaml
  ccmd sync

  # Preview changes without applying them
  ccmd sync --dry-run

  # Skip confirmation prompts
  ccmd sync --force`,
		Args: cobra.NoArgs,
		RunE: errors.WrapCommand("sync", func(cmd *cobra.Command, args []string) error {
			return runSync(dryRun, force)
		}),
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")

	return cmd
}

func runSync(dryRun, force bool) error {
	log := logger.WithField("command", "sync")
	log.Debug("starting sync")

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

	// Get list of installed commands
	listOpts := commands.ListOptions{}
	installedList, err := commands.List(listOpts)
	if err != nil {
		return fmt.Errorf("failed to list installed commands: %w", err)
	}

	// Create map of installed commands for easy lookup
	installedMap := make(map[string]*commands.CommandDetail)
	for _, cmd := range installedList {
		installedMap[cmd.Name] = cmd
	}

	// Create map of config commands
	configCommands, err := config.GetCommands()
	if err != nil {
		return fmt.Errorf("failed to get commands from config: %w", err)
	}

	configMap := make(map[string]project.ConfigCommand)
	for _, cmd := range configCommands {
		_, repoName, err := cmd.ParseOwnerRepo()
		if err != nil {
			output.PrintErrorf("Invalid repository format in ccmd.yaml: %s", cmd.Repo)
			continue
		}
		configMap[repoName] = cmd
	}

	// Analyze what needs to be done
	result := analyzeSync(configMap, installedMap)

	// Display analysis
	output.PrintInfof("Sync Analysis:")
	output.Printf("")

	if len(result.ToInstall) == 0 && len(result.ToRemove) == 0 {
		output.PrintSuccessf("✓ All commands are in sync")
		return nil
	}

	if len(result.ToInstall) > 0 {
		output.PrintInfof("Commands to install:")
		for _, name := range result.ToInstall {
			if cmd, ok := configMap[name]; ok {
				if cmd.Version != "" {
					output.PrintInfof("  + %s@%s", cmd.Repo, cmd.Version)
				} else {
					output.PrintInfof("  + %s", cmd.Repo)
				}
			}
		}
		output.Printf("")
	}

	if len(result.ToRemove) > 0 {
		output.PrintInfof("Commands to remove (not in ccmd.yaml):")
		for _, name := range result.ToRemove {
			output.PrintInfof("  - %s", name)
		}
		output.Printf("")
	}

	// If dry run, stop here
	if dryRun {
		output.PrintInfof("Dry run mode - no changes were made")
		return nil
	}

	// Confirm removal if needed and not forced
	if len(result.ToRemove) > 0 && !force {
		output.PrintWarningf("The following commands will be removed:")
		for _, name := range result.ToRemove {
			output.PrintWarningf("  - %s", name)
		}
		output.Printf("")

		if !promptConfirmation("Do you want to proceed with removing these commands?") {
			output.PrintInfof("Sync canceled")
			return nil
		}
	}

	// Execute sync
	output.PrintInfof("Executing sync...")
	output.Printf("")

	// Install missing commands
	if len(result.ToInstall) > 0 {
		output.PrintInfof("Installing commands...")
		for _, name := range result.ToInstall {
			cmd := configMap[name]
			repository := fmt.Sprintf("https://github.com/%s.git", cmd.Repo)

			spinner := output.NewSpinner(fmt.Sprintf("Installing %s...", cmd.Repo))
			spinner.Start()

			installOpts := commands.InstallOptions{
				Repository: repository,
				Version:    cmd.Version,
				Name:       name,
				Force:      false,
			}

			if err := commands.Install(installOpts); err != nil {
				spinner.Stop()
				output.PrintErrorf("Failed to install %s: %v", cmd.Repo, err)
				result.Errors = append(result.Errors, fmt.Errorf("install %s: %w", cmd.Repo, err))
				continue
			}

			spinner.Stop()
			output.PrintSuccessf("✓ Installed %s", cmd.Repo)
		}
		output.Printf("")
	}

	// Remove extra commands
	if len(result.ToRemove) > 0 {
		output.PrintInfof("Removing commands...")
		for _, name := range result.ToRemove {
			spinner := output.NewSpinner(fmt.Sprintf("Removing %s...", name))
			spinner.Start()

			removeOpts := commands.RemoveOptions{
				Name: name,
			}

			if err := commands.Remove(removeOpts); err != nil {
				spinner.Stop()
				output.PrintErrorf("Failed to remove %s: %v", name, err)
				result.Errors = append(result.Errors, fmt.Errorf("remove %s: %w", name, err))
				continue
			}

			spinner.Stop()
			output.PrintSuccessf("✓ Removed %s", name)
		}
		output.Printf("")
	}

	// Update lock file
	if err := updateLockFile(pm); err != nil {
		log.WithError(err).Warn("failed to update ccmd-lock.yaml")
		output.PrintWarningf("Failed to update ccmd-lock.yaml: %v", err)
	} else {
		output.PrintInfof("Updated ccmd-lock.yaml")
	}

	// Summary
	output.Printf("")
	installedCount := len(result.ToInstall) - countErrors(result.Errors, "install")
	removedCount := len(result.ToRemove) - countErrors(result.Errors, "remove")

	if len(result.Errors) == 0 {
		output.PrintSuccessf("Sync completed successfully!")
		if installedCount > 0 {
			output.PrintSuccessf("  %d command(s) installed", installedCount)
		}
		if removedCount > 0 {
			output.PrintSuccessf("  %d command(s) removed", removedCount)
		}
	} else {
		output.PrintWarningf("Sync completed with %d error(s)", len(result.Errors))
		if installedCount > 0 {
			output.PrintInfof("  %d command(s) installed", installedCount)
		}
		if removedCount > 0 {
			output.PrintInfof("  %d command(s) removed", removedCount)
		}
		output.PrintErrorf("  %d operation(s) failed", len(result.Errors))
	}

	return nil
}

// analyzeSync compares config with installed commands and returns what needs to be done.
func analyzeSync(configCommands map[string]project.ConfigCommand,
	installedCommands map[string]*commands.CommandDetail) Result {
	result := Result{
		ToInstall: []string{},
		ToRemove:  []string{},
		Errors:    []error{},
	}

	// Find commands to install (in config but not installed)
	for name := range configCommands {
		if _, exists := installedCommands[name]; !exists {
			result.ToInstall = append(result.ToInstall, name)
		}
	}

	// Find commands to remove (installed but not in config)
	for name := range installedCommands {
		if _, exists := configCommands[name]; !exists {
			result.ToRemove = append(result.ToRemove, name)
		}
	}

	return result
}

// promptConfirmation prompts the user for yes/no confirmation.
func promptConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return isConfirmation(response)
}

// isConfirmation checks if the input is a confirmation (y/yes).
func isConfirmation(input string) bool {
	response := strings.TrimSpace(strings.ToLower(input))
	return response == "y" || response == "yes"
}

// countErrors counts errors containing a specific operation keyword.
func countErrors(errors []error, operation string) int {
	count := 0
	for _, err := range errors {
		if strings.Contains(err.Error(), operation) {
			count++
		}
	}
	return count
}

// updateLockFile regenerates the lock file based on currently installed commands.
func updateLockFile(pm *project.Manager) error {
	// Get current list of installed commands
	listOpts := commands.ListOptions{}
	installedList, err := commands.List(listOpts)
	if err != nil {
		return fmt.Errorf("failed to list installed commands: %w", err)
	}

	// Create new lock file
	lockFile := project.NewLockFile()

	// Add each installed command to lock file
	for _, cmd := range installedList {
		// Create command entry from installed command info
		projectCmd := &project.Command{
			Name:         cmd.Name,
			Source:       cmd.Source,
			Version:      cmd.Version,
			Resolved:     cmd.Source + "@" + cmd.Version,
			InstalledAt:  cmd.InstalledAt,
			UpdatedAt:    cmd.UpdatedAt,
			Dependencies: cmd.Dependencies,
			Metadata:     cmd.Metadata,
		}

		if err := lockFile.AddCommand(projectCmd); err != nil {
			return fmt.Errorf("failed to add command %s to lock file: %w", cmd.Name, err)
		}
	}

	// Save the lock file
	return pm.SaveLockFile(lockFile)
}

package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/gifflet/ccmd/internal/model"
)

// workspace implements the Workspace interface for managing multiple repositories
type workspace struct {
	mu           sync.RWMutex
	repositories map[string]Repository
	order        []string // maintains repository priority order
}

// NewWorkspace creates a new workspace instance
func NewWorkspace() Workspace {
	return &workspace{
		repositories: make(map[string]Repository),
		order:        make([]string, 0),
	}
}

// AddRepository adds a new repository to the workspace
func (w *workspace) AddRepository(repo Repository, name string) error {
	if repo == nil {
		return fmt.Errorf("repository cannot be nil")
	}

	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.repositories[name]; exists {
		return fmt.Errorf("repository %q already exists", name)
	}

	w.repositories[name] = repo
	w.order = append(w.order, name)

	return nil
}

// RemoveRepository removes a repository from the workspace
func (w *workspace) RemoveRepository(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.repositories[name]; !exists {
		return fmt.Errorf("repository %q not found", name)
	}

	delete(w.repositories, name)

	// Remove from order slice
	newOrder := make([]string, 0, len(w.order)-1)
	for _, n := range w.order {
		if n != name {
			newOrder = append(newOrder, n)
		}
	}
	w.order = newOrder

	return nil
}

// ListCommands returns all available commands across all repositories
func (w *workspace) ListCommands(ctx context.Context) ([]model.CommandInfo, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var allCommands []model.CommandInfo
	var errors []error

	// Use a map to track unique commands (by name)
	seen := make(map[string]bool)

	// Iterate in priority order
	for _, name := range w.order {
		repo := w.repositories[name]

		commands, err := repo.Discover(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("repository %q: %w", name, err))
			continue
		}

		for _, cmd := range commands {
			// Skip if we've already seen this command name
			// (higher priority repository wins)
			if seen[cmd.Name] {
				continue
			}

			seen[cmd.Name] = true
			allCommands = append(allCommands, cmd)
		}
	}

	// If all repositories failed, return error
	if len(errors) == len(w.repositories) && len(errors) > 0 {
		return nil, fmt.Errorf("all repositories failed: %v", errors)
	}

	return allCommands, nil
}

// ResolveCommand finds and loads a command by name across all repositories
func (w *workspace) ResolveCommand(ctx context.Context, name, version string) (*model.Command, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.repositories) == 0 {
		return nil, fmt.Errorf("no repositories configured")
	}

	var lastErr error

	// Search repositories in priority order
	for _, repoName := range w.order {
		repo := w.repositories[repoName]

		// First, check if the repository has the command
		commands, err := repo.Discover(ctx)
		if err != nil {
			lastErr = fmt.Errorf("repository %q discover: %w", repoName, err)
			continue
		}

		// Look for the command in discovered list
		var commandInfo *model.CommandInfo
		for _, cmd := range commands {
			if cmd.Name == name {
				// Check version if specified
				if version != "" && !isVersionMatch(cmd.Version, version) {
					continue
				}
				commandInfo = &cmd
				break
			}
		}

		if commandInfo == nil {
			continue
		}

		// Load the command
		command, err := repo.Load(ctx, commandInfo.ID)
		if err != nil {
			lastErr = fmt.Errorf("repository %q load: %w", repoName, err)
			continue
		}

		return command, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("command %q not found: %w", name, lastErr)
	}

	return nil, fmt.Errorf("command %q not found in any repository", name)
}

// isVersionMatch checks if a command version matches the requested version
// TODO: Implement proper version constraint matching (e.g., semver)
func isVersionMatch(cmdVersion, requestedVersion string) bool {
	// For now, just do exact match
	// In the future, this should support version constraints like "^1.0.0", "~2.1.0", etc.
	return cmdVersion == requestedVersion
}

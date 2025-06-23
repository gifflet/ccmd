// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package repository

import (
	"context"
	"io"

	"github.com/gifflet/ccmd/internal/model"
)

// Repository defines the interface for managing command repositories.
// It provides operations for discovering, loading, and accessing commands
// from various sources like local directories, Git repositories, and remote catalogs.
//
// Implementation consideration: In Docker-in-Docker environments, git operations
// may fail with authentication issues. Implementations should handle these gracefully
// and provide fallback options (e.g., using cached data or local repositories).
type Repository interface {
	// Initialize sets up the repository connection and prepares for operations.
	// It should validate connectivity, check authentication, and prepare any
	// necessary local storage or caches.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - error: Any error encountered during initialization
	Initialize(ctx context.Context) error

	// Discover searches for available commands in the repository.
	// It should recursively scan for ccmd.yaml files and build a catalog
	// of available commands with their metadata.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - []model.CommandInfo: List of discovered commands with metadata
	//   - error: Any error encountered during discovery
	//
	// Note: Implementations should handle large repositories efficiently,
	// potentially using pagination or streaming for better performance.
	Discover(ctx context.Context) ([]model.CommandInfo, error)

	// Load retrieves a specific command by its identifier.
	// The identifier format depends on the repository type:
	//   - Local: file path relative to repository root
	//   - Git: combination of repository URL and path
	//   - Remote: URL or catalog-specific identifier
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//   - id: Unique identifier for the command
	//
	// Returns:
	//   - *model.Command: The loaded command with full configuration
	//   - error: Any error encountered during loading
	Load(ctx context.Context, id string) (*model.Command, error)

	// GetResource retrieves a resource file associated with a command.
	// Resources can include templates, configuration files, or documentation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//   - commandID: Identifier of the command owning the resource
	//   - resourcePath: Path to the resource relative to the command
	//
	// Returns:
	//   - io.ReadCloser: Reader for the resource content (caller must close)
	//   - error: Any error encountered accessing the resource
	//
	// Note: Implementations should validate resource paths to prevent
	// directory traversal attacks.
	GetResource(ctx context.Context, commandID, resourcePath string) (io.ReadCloser, error)

	// Type returns the repository type identifier.
	// Common types include "local", "git", "http", "s3", etc.
	//
	// Returns:
	//   - string: Repository type identifier
	Type() string

	// Info returns metadata about the repository.
	// This includes connection details, capabilities, and current status.
	//
	// Returns:
	//   - model.RepositoryInfo: Repository metadata and status
	//
	// TODO: Define RepositoryInfo model structure including:
	//   - Connection status and health
	//   - Supported features and limitations
	//   - Cache status and statistics
	//   - Authentication status
	Info() model.RepositoryInfo
}

// Factory is a function that creates a new Repository instance.
// It receives configuration specific to the repository type.
//
// Parameters:
//   - config: Type-specific configuration map
//
// Returns:
//   - Repository: New repository instance
//   - error: Any error during creation
//
// Example configurations:
//   - Local: {"path": "/path/to/commands"}
//   - Git: {"url": "https://github.com/user/repo.git", "branch": "main"}
//   - HTTP: {"url": "https://catalog.example.com", "apiKey": "..."}
type Factory func(config map[string]interface{}) (Repository, error)

// Manager handles multiple repository instances and provides unified access.
// It maintains a registry of repository types and can create instances dynamically.
type Manager interface {
	// Register adds a new repository type with its factory function.
	//
	// Parameters:
	//   - repoType: Unique identifier for the repository type
	//   - factory: Function to create instances of this type
	Register(repoType string, factory Factory)

	// Create instantiates a new repository of the specified type.
	//
	// Parameters:
	//   - repoType: Type of repository to create
	//   - config: Configuration for the repository
	//
	// Returns:
	//   - Repository: New repository instance
	//   - error: Any error during creation
	Create(repoType string, config map[string]interface{}) (Repository, error)

	// List returns all registered repository types.
	//
	// Returns:
	//   - []string: List of registered type identifiers
	List() []string
}

// CacheableRepository extends Repository with caching capabilities.
// Implementations can use this to improve performance for remote repositories.
type CacheableRepository interface {
	Repository

	// RefreshCache updates the local cache with latest repository data.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - error: Any error during cache refresh
	RefreshCache(ctx context.Context) error

	// ClearCache removes all cached data for this repository.
	//
	// Returns:
	//   - error: Any error during cache clearing
	ClearCache() error

	// GetCacheInfo returns information about the current cache state.
	//
	// Returns:
	//   - model.CacheInfo: Cache statistics and metadata
	//
	// TODO: Define CacheInfo model structure
	GetCacheInfo() model.CacheInfo
}

// Workspace provides a unified interface for accessing multiple repositories.
// It manages repository lifecycle and provides command resolution across sources.
type Workspace interface {
	// AddRepository adds a new repository to the workspace.
	//
	// Parameters:
	//   - repo: Repository instance to add
	//   - name: Unique name for this repository in the workspace
	//
	// Returns:
	//   - error: Any error during addition
	AddRepository(repo Repository, name string) error

	// RemoveRepository removes a repository from the workspace.
	//
	// Parameters:
	//   - name: Name of the repository to remove
	//
	// Returns:
	//   - error: Any error during removal
	RemoveRepository(name string) error

	// ListCommands returns all available commands across all repositories.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - []model.CommandInfo: Aggregated list of all commands
	//   - error: Any error during listing
	ListCommands(ctx context.Context) ([]model.CommandInfo, error)

	// ResolveCommand finds and loads a command by name across all repositories.
	// It should handle version resolution and repository priority.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//   - name: Command name to resolve
	//   - version: Specific version or version constraint (optional)
	//
	// Returns:
	//   - *model.Command: Resolved and loaded command
	//   - error: Any error during resolution
	ResolveCommand(ctx context.Context, name, version string) (*model.Command, error)
}

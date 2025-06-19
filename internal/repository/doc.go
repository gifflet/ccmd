// Package repository provides interfaces and implementations for managing command repositories.
//
// The repository package defines a flexible system for discovering, loading, and managing
// commands from various sources including local directories, Git repositories, and remote
// catalogs. It provides a unified interface that allows ccmd to work with multiple
// repository types seamlessly.
//
// # Core Concepts
//
// Repository: A source of commands, which can be local, remote, or version-controlled.
// Each repository type implements the Repository interface to provide consistent access
// to commands regardless of the underlying storage mechanism.
//
// Manager: Handles the lifecycle of repository instances and maintains a registry of
// available repository types. New repository types can be registered at runtime.
//
// Workspace: Provides a unified view across multiple repositories, handling command
// resolution, version management, and repository prioritization.
//
// # Repository Types
//
// The package supports several repository types out of the box:
//
//   - Local: Commands stored in local directories
//   - Git: Commands in Git repositories (with branch/tag support)
//   - HTTP: Commands served from HTTP endpoints
//   - S3: Commands stored in S3-compatible object storage
//
// # Usage Example
//
//	// Create a repository manager
//	manager := repository.NewManager()
//
//	// Register repository types
//	manager.Register("local", repository.NewLocalFactory())
//	manager.Register("git", repository.NewGitFactory())
//
//	// Create a repository instance
//	repo, err := manager.Create("local", map[string]interface{}{
//	    "path": "/path/to/commands",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Initialize and discover commands
//	ctx := context.Background()
//	if err := repo.Initialize(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	commands, err := repo.Discover(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Extending with Custom Repository Types
//
// Custom repository types can be added by implementing the Repository interface
// and registering a factory function:
//
//	type MyCustomRepo struct {
//	    // implementation
//	}
//
//	func (r *MyCustomRepo) Initialize(ctx context.Context) error {
//	    // implementation
//	}
//
//	// ... implement other Repository methods
//
//	// Register the custom type
//	manager.Register("custom", func(config map[string]interface{}) (Repository, error) {
//	    return &MyCustomRepo{
//	        // initialize from config
//	    }, nil
//	})
//
// # Error Handling
//
// The package defines specific error types for common failure scenarios:
//
//   - ErrRepositoryNotFound: Repository doesn't exist or is inaccessible
//   - ErrCommandNotFound: Command doesn't exist in the repository
//   - ErrInvalidConfiguration: Repository configuration is invalid
//   - ErrAuthenticationFailed: Repository requires authentication
//
// # Performance Considerations
//
// For remote repositories, implementations should consider:
//
//   - Caching discovered commands to reduce network calls
//   - Lazy loading of command details and resources
//   - Parallel discovery for improved performance
//   - Progress reporting for long-running operations
//
// # Security Considerations
//
// Repository implementations must:
//
//   - Validate all file paths to prevent directory traversal
//   - Sanitize command metadata to prevent injection attacks
//   - Handle credentials securely (never log or expose them)
//   - Verify checksums/signatures for downloaded content
package repository

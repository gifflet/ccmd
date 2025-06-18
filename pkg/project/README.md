# Project Package

This package provides comprehensive functionality for managing ccmd project files (`ccmd.yaml` and `ccmd-lock.yaml`).

## Overview

The project package provides three main components:

1. **Configuration Management** (`schema.go`) - Handles `ccmd.yaml` files
2. **Lock File Management** (`lock.go`) - Handles `ccmd-lock.yaml` files  
3. **Project Manager** (`manager.go`) - High-level API for project operations

## Usage

### Using the Project Manager (Recommended)

```go
import "github.com/gifflet/ccmd/pkg/project"

// Create a manager for the current directory
manager := project.NewManager(".")

// Initialize a new project
err := manager.InitializeConfig()
if err != nil {
    log.Fatal(err)
}

// Add commands
err = manager.AddCommand("github/cli", "v2.0.0")
err = manager.AddCommand("junegunn/fzf", "latest")

// Update a command version
err = manager.UpdateCommand("github/cli", "v2.1.0")

// Remove a command
err = manager.RemoveCommand("junegunn/fzf")
```

### Direct Configuration File Operations

```go
// Load from file
config, err := project.LoadConfig("ccmd.yaml")
if err != nil {
    log.Fatal(err)
}

// Save to file
err = project.SaveConfig(config, "ccmd.yaml")
if err != nil {
    log.Fatal(err)
}

// Parse from reader
config, err := project.ParseConfig(reader)
if err != nil {
    log.Fatal(err)
}

// Write to writer
err = project.WriteConfig(config, writer)
if err != nil {
    log.Fatal(err)
}
```

### Lock File Operations

```go
// Create new lock file
lockFile := project.NewLockFile()

// Add a command
cmd := &project.Command{
    Name:         "gh",
    Repository:   "github/cli",  
    Version:      "v2.0.0",
    CommitHash:   "abc123...", // 40 char SHA
    InstalledAt:  time.Now(),
    UpdatedAt:    time.Now(),
    FileSize:     10485760,
    Checksum:     "sha256...", // 64 char SHA256
}
err := lockFile.AddCommand(cmd)

// Save to disk
err = lockFile.SaveToFile("ccmd-lock.yaml")

// Load from disk
lockFile, err = project.LoadFromFile("ccmd-lock.yaml")
```

## Schema Structure

The `ccmd.yaml` file has a simple structure:

```yaml
commands:
  - repo: owner/repository
    version: v1.0.0  # optional
```

### Fields

- **commands**: List of command declarations (required)
  - **repo**: Repository in format "owner/repository" (required)
  - **version**: Version specification (optional, defaults to "latest")

### Version Specification

The version field supports:
- Semantic versions: `v1.0.0`, `1.2.3`, `v2.0.0-beta.1`
- Branch names: `main`, `develop`, `feature/xyz`
- Tag names: `release-1.0`, `stable`
- Special value: `latest` (default)

### Validation Rules

1. At least one command must be defined
2. Repository must be in "owner/repo" format
3. Owner and repo names must contain only alphanumeric characters, hyphens, and underscores
4. Version format is validated for basic correctness

## Lock File Structure

The `ccmd-lock.yaml` file records installed command metadata:

```yaml
version: "1.0"
updated_at: 2024-01-01T00:00:00Z
commands:
  gh:
    name: gh
    repository: github/cli
    version: v2.0.0
    commit_hash: 1234567890abcdef1234567890abcdef12345678
    installed_at: 2024-01-01T00:00:00Z
    updated_at: 2024-01-01T00:00:00Z
    file_size: 10485760
    checksum: abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
    dependencies: []
    metadata:
      description: GitHub CLI
```

## Examples

### Complete ccmd.yaml Example

```yaml
commands:
  # Semantic version
  - repo: example/tool
    version: v1.2.3
  
  # Latest version (explicit)
  - repo: org/cli
    version: latest
  
  # Branch
  - repo: user/command
    version: develop
  
  # Default to latest (version omitted)
  - repo: tools/formatter
```

### Working with Commands

```go
// Parse owner and repo from command
cmd := config.Commands[0]
owner, repo, err := cmd.ParseOwnerRepo()
fmt.Printf("Owner: %s, Repo: %s\n", owner, repo)

// Check if using semantic version
if cmd.IsSemanticVersion() {
    fmt.Println("Using semantic version:", cmd.Version)
}

// Calculate file checksum
checksum, err := project.CalculateChecksum("/path/to/binary")
if err != nil {
    log.Fatal(err)
}
```

## Features

### Atomic File Operations

All file writes are performed atomically to prevent corruption:
- Write to temporary file
- Validate write succeeded  
- Rename temp file to target
- Clean up on failure

### Comprehensive Validation

- Repository format must be `owner/repo`
- Owner and repo names validated for allowed characters
- Version format validated (semantic version, branch, or tag)
- Lock file entries validated for completeness
- Checksums must be 64-char SHA256
- Commit hashes must be 40-char SHA

### Empty Configuration Support

Empty configuration files are valid, making it easy to start new projects:

```go
manager := project.NewManager(".")
err := manager.InitializeConfig() // Creates empty but valid ccmd.yaml
```
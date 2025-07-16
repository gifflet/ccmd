# ccmd Architecture

This document describes the internal architecture of ccmd, including design decisions, component interactions, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Design Principles](#design-principles)
- [System Architecture](#system-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [File System Layout](#file-system-layout)
- [Error Handling](#error-handling)
- [Security Considerations](#security-considerations)
- [Performance](#performance)
- [Future Considerations](#future-considerations)

## Overview

ccmd is designed as a lightweight, efficient command manager for Claude Code. It follows a simplified 2-layer architecture that separates concerns while avoiding unnecessary abstractions.

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│                      (cmd/*/...)                            │
│            Parsing, validation, user interaction            │
├─────────────────────────────────────────────────────────────┤
│                       Core Layer                            │
│                        (core/)                              │
│          Business logic, Git operations, metadata           │
├─────────────────────────────────────────────────────────────┤
│                    Support Packages                         │
│     ┌─────────────┬──────────────┬─────────────────┐      │
│     │   Output    │    Errors    │     Logger      │      │
│     │(pkg/output) │(pkg/errors)  │  (pkg/logger)   │      │
│     └─────────────┴──────────────┴─────────────────┘      │
│                    Test Utilities                           │
│                   (internal/fs)                             │
└─────────────────────────────────────────────────────────────┘
```

## Design Principles

### 1. Simplicity (KISS)
- **2-layer architecture**: CLI → Core
- Direct function calls over abstractions
- Clear, readable code

### 2. Minimalism
- One file per command in both layers
- Consolidated business logic in core/

### 3. Testability
- FileSystem interface for testing (internal/fs)
- Clear separation of concerns
- Mockable where necessary

### 4. Performance
- Direct Git CLI usage via exec.Command
- Minimal file I/O

### 5. User Experience
- Colored output with pkg/output
- Progress indicators and spinners
- Clear error messages via pkg/errors

## System Architecture

### 2-Layer Architecture

1. **CLI Layer** (`cmd/`)
   - Command parsing with Cobra
   - Flag validation
   - User interaction
   - Delegates to Core layer

2. **Core Layer** (`core/`)
   - All business logic
   - Git operations
   - File system operations
   - Metadata management

### Support Packages

- **pkg/errors**: Consistent error handling with sentinel errors
- **pkg/logger**: Convenient wrapper over slog
- **pkg/output**: Colored output, progress bars, spinners
- **internal/fs**: FileSystem interface for testing

### Key Components

#### CLI Entry Point (`cmd/ccmd/main.go`)

```go
// Each command follows a simple pattern
func main() {
    rootCmd := &cobra.Command{
        Use:   "ccmd",
        Short: "Claude Code Command Manager",
    }
    
    // Add commands
    rootCmd.AddCommand(
        install.NewCommand(),
        list.NewCommand(),
        remove.NewCommand(),
        // ...
    )
    
    rootCmd.Execute()
}
```

#### Command Implementation (`cmd/install/install.go`)

```go
// Minimal CLI layer - just parsing and delegation
func NewCommand() *cobra.Command {
    var version, name string
    var force bool
    
    cmd := &cobra.Command{
        Use:   "install [repository]",
        Short: "Install a command from a Git repository",
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 0 {
                return core.InstallFromConfig(force)
            }
            
            return core.Install(context.Background(), core.InstallOptions{
                Repository: args[0],
                Version:    version,
                Name:       name,
                Force:      force,
            })
        },
    }
    
    // Add flags
    return cmd
}
```

#### Core Logic (`core/install.go`)

```go
// All business logic in one place
type InstallOptions struct {
    Repository string
    Version    string
    Name       string
    Force      bool
}

func Install(ctx context.Context, opts InstallOptions) error {
    // 1. Validate input
    // 2. Clone repository
    // 3. Validate command structure
    // 4. Install files
    // 5. Update lock file
    
    return nil
}
```

## Component Details

### Git Operations (`core/git.go`)

Simple, direct Git operations using exec.Command:

```go
// Direct Git CLI usage - no abstractions
func gitClone(ctx context.Context, repo, dest, version string) error {
    args := []string{"clone", "--depth", "1"}
    if version != "" {
        args = append(args, "--branch", version)
    }
    args = append(args, repo, dest)
    
    cmd := exec.CommandContext(ctx, "git", args...)
    return cmd.Run()
}

func gitFetch(ctx context.Context, dir string) error {
    cmd := exec.CommandContext(ctx, "git", "fetch", "--tags")
    cmd.Dir = dir
    return cmd.Run()
}
```

### Metadata Management (`core/metadata.go`)

Handles project configuration and lock files:

```go
type ProjectConfig struct {
    Name        string   `yaml:"name,omitempty"`
    Version     string   `yaml:"version,omitempty"`
    Description string   `yaml:"description,omitempty"`
    Author      string   `yaml:"author,omitempty"`
    Repository  string   `yaml:"repository,omitempty"`
    Commands    []string `yaml:"commands,omitempty"`
}

type LockFile struct {
    Version  string                 `yaml:"version"`
    Commands map[string]CommandLock `yaml:"commands"`
}

// Direct functions - no unnecessary interfaces
func LoadProjectConfig(projectPath string) (*ProjectConfig, error)
func SaveProjectConfig(projectPath string, config *ProjectConfig) error
func LoadLockFile(projectPath string) (*LockFile, error)
func SaveLockFile(projectPath string, lock *LockFile) error
```

### Type Definitions (`core/types.go`)

Consolidated types used across the core:

```go
type CommandDetail struct {
    Name        string
    Version     string
    Repository  string
    InstalledAt string
    UpdatedAt   string
    Description string
}

type CommandMetadata struct {
    Name        string   `yaml:"name"`
    Version     string   `yaml:"version"`
    Description string   `yaml:"description"`
    Author      string   `yaml:"author"`
    Repository  string   `yaml:"repository"`
    Entry       string   `yaml:"entry"`
    Tags        []string `yaml:"tags"`
}
```

### File System Abstraction (`internal/fs/`)

Only abstraction kept for testing purposes:

```go
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Stat(path string) (os.FileInfo, error)
    MkdirAll(path string, perm os.FileMode) error
    RemoveAll(path string) error
    Exists(path string) (bool, error)
}

// Implementations
type OS struct{}     // Real filesystem
type MemFS struct{} // In-memory for tests
```

### Output Manager (`pkg/output/`)

Rich terminal output capabilities:

```go
// Colored output functions
func Print(format string, args ...interface{})
func PrintInfo(format string, args ...interface{})
func PrintSuccess(format string, args ...interface{})
func PrintWarning(format string, args ...interface{})
func PrintError(format string, args ...interface{})

// Progress indicators
func NewProgressBar(total int) *ProgressBar
func NewSpinner(message string) *Spinner

// Interactive prompts
func Prompt(message string) string
func Confirm(message string) bool
```

## Data Flow

### Install Command Flow

```
User Input → CLI Parser → core.Install()
                              ↓
                         Validate Input
                              ↓
                      Clone Repository (git)
                              ↓
                    Validate Command Structure
                              ↓
                       Install Files
                              ↓
                    Update Lock File
                              ↓
                      Success Output
```

### List Command Flow

```
User Input → CLI Parser → core.List()
                              ↓
                      Read Lock File
                              ↓
                 Read Command Metadata
                              ↓
                    Format and Display
```

## File System Layout

### Project Structure

```
my-project/
├── .claude/
│   ├── commands/              # Installed commands
│   │   ├── command1/         # Full command directory
│   │   │   ├── ccmd.yaml
│   │   │   ├── index.md
│   │   │   └── ...
│   │   └── command1.md       # Standalone markdown
├── ccmd.yaml                 # Project configuration
├── ccmd-lock.yaml           # Lock file
└── src/                     # Project files
```


## Error Handling

### Sentinel Errors (`pkg/errors`)

ccmd uses sentinel errors for consistent error handling:

```go
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInvalidInput  = errors.New("invalid input")
    ErrGitOperation  = errors.New("git operation failed")
    ErrFileOperation = errors.New("file operation failed")
)
```

### Error Creation Functions

Helper functions provide consistent error messages with context:

```go
// Create specific errors with context
errors.NotFound("command foo")           // "not found: command foo"
errors.AlreadyExists("command bar")      // "already exists: command bar"
errors.InvalidInput("invalid version")    // "invalid input: invalid version"
errors.GitError("clone", err)            // "git operation failed during clone: ..."
errors.FileError("read", path, err)      // "file operation failed: read on path: ..."
```

## Security Considerations

### Repository Validation

- Only clone from HTTPS/SSH URLs
- Validate repository structure before installation
- Check file permissions
- Prevent directory traversal

### Command Execution

- Commands are markdown files, not executables
- No automatic code execution
- User must explicitly run commands
- Sandboxed to project directory

## Performance

### Optimization Strategies

1. **Shallow Cloning**
   ```bash
   git clone --depth 1 --single-branch
   ```

2. **Efficient I/O**
   - Read files once
   - Batch operations where possible
   - Use buffered I/O

### Performance Targets

- Install: < 5 seconds
- List: < 100ms
- Remove: < 500ms
- Update: < 5 seconds per command

## Testing Strategy

### Unit Tests

- Test core logic in isolation
- Mock filesystem with internal/fs
- Focus on business logic

### Integration Tests

- Test end-to-end workflows
- Use real filesystem in temp directories
- Verify command interactions

### Example Test

```go
func TestInstall(t *testing.T) {
    // Use memory filesystem for testing
    fs := &memfs.MemFS{}
    
    // Test installation
    err := InstallWithFS(context.Background(), InstallOptions{
        Repository: "github.com/user/repo",
        Name:       "test-cmd",
    }, fs)
    
    assert.NoError(t, err)
    assert.True(t, fs.Exists(".claude/commands/test-cmd/index.md"))
}
```

## Conclusion

ccmd's architecture prioritizes simplicity, modularity, and extensibility. By following clear design principles and maintaining separation of concerns, the codebase remains maintainable and easy to extend with new features.
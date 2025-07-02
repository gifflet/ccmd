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

ccmd is designed as a lightweight, efficient command manager for Claude Code. It follows a modular architecture that separates concerns and allows for easy extension and maintenance.

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│                    (cmd/ccmd/main.go)                      │
├─────────────────────────────────────────────────────────────┤
│                     Command Layer                           │
│                   (cmd/*/?.go files)                       │
├─────────────────────────────────────────────────────────────┤
│                     Business Logic                          │
│                    (pkg/commands/)                          │
├─────────────────────────────────────────────────────────────┤
│                    Core Services                            │
│        ┌────────────┬──────────────┬────────────┐         │
│        │ Git Client │ Lock Manager │ Filesystem │         │
│        │  (pkg/git) │(internal/lock)│(internal/fs)│        │
│        └────────────┴──────────────┴────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                      Data Models                            │
│                   (internal/models/)                        │
└─────────────────────────────────────────────────────────────┘
```

## Design Principles

### 1. Simplicity (KISS)
- Minimal abstractions
- Clear, readable code
- Straightforward implementations

### 2. Modularity
- Separate concerns
- Reusable components
- Clear interfaces

### 3. Testability
- Dependency injection
- Interface-based design
- Mockable external dependencies

### 4. Performance
- Efficient Git operations
- Minimal file I/O
- Concurrent operations where beneficial

### 5. User Experience
- Clear error messages
- Progress feedback
- Intuitive commands

## System Architecture

### Layer Separation

1. **Presentation Layer** (`cmd/`)
   - CLI parsing and validation
   - User interaction
   - Output formatting

2. **Business Logic Layer** (`pkg/commands/`)
   - Command implementations
   - Business rules
   - Workflow orchestration

3. **Service Layer** (`pkg/`, `internal/`)
   - Git operations
   - File system operations
   - Lock management
   - Data persistence

4. **Data Layer** (`internal/models/`)
   - Data structures
   - Serialization/deserialization
   - Validation

### Key Components

#### CLI Entry Point (`cmd/ccmd/main.go`)

```go
// Simplified flow
func main() {
    app := &cli.App{
        Name:     "ccmd",
        Commands: []*cli.Command{
            installCmd,
            listCmd,
            removeCmd,
            searchCmd,
            updateCmd,
        },
    }
    app.Run(os.Args)
}
```

#### Command Implementation (`cmd/install/install.go`)

```go
// Each command follows this pattern
func Command() *cli.Command {
    return &cli.Command{
        Name:   "install",
        Action: runInstall,
        Flags:  installFlags(),
    }
}

func runInstall(c *cli.Context) error {
    // 1. Parse arguments
    // 2. Create command instance
    // 3. Execute business logic
    // 4. Handle output
}
```

#### Business Logic (`pkg/commands/install.go`)

```go
type InstallCommand struct {
    git    git.Client
    fs     fs.Interface
    lock   lock.Manager
    output output.Interface
}

func (c *InstallCommand) Execute(repo string, opts Options) error {
    // 1. Validate repository
    // 2. Clone repository
    // 3. Validate command structure
    // 4. Install files
    // 5. Update lock file
}
```

## Component Details

### Git Client (`pkg/git/`)

Handles all Git operations:

```go
type Client interface {
    Clone(url string, opts CloneOptions) error
    GetTags(url string) ([]string, error)
    GetLatestTag(url string) (string, error)
    GetCommitHash(path string) (string, error)
}
```

Features:
- Shallow cloning for efficiency
- Tag/version resolution
- Support for multiple Git providers
- Authentication handling

### Lock Manager (`internal/lock/`)

Manages the ccmd-lock.yaml file:

```go
type Manager interface {
    Read() (*LockFile, error)
    Write(lock *LockFile) error
    AddEntry(name string, entry LockEntry) error
    RemoveEntry(name string) error
    GetEntry(name string) (*LockEntry, bool)
}
```

Lock file format:
```json
{
  "version": "1.0",
  "commands": {
    "my-command": {
      "version": "1.2.3",
      "repository": "github.com/user/repo",
      "commit": "abc123...",
      "installed_at": "2024-01-01T00:00:00Z",
      "integrity": "sha256:..."
    }
  }
}
```

### File System Abstraction (`internal/fs/`)

Provides testable file operations:

```go
type Interface interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    Remove(path string) error
    RemoveAll(path string) error
    Exists(path string) bool
    Walk(root string, fn filepath.WalkFunc) error
}
```

Implementations:
- `RealFS`: Actual file system operations
- `MemFS`: In-memory for testing

### Output Manager (`internal/output/`)

Handles all user output:

```go
type Interface interface {
    Info(format string, args ...interface{})
    Success(format string, args ...interface{})
    Warning(format string, args ...interface{})
    Error(format string, args ...interface{})
    StartSpinner(message string) func()
    Progress(current, total int, message string)
}
```

Features:
- Colored output
- Progress indicators
- Spinner animations
- Table formatting

## Data Flow

### Install Command Flow

```
User Input → CLI Parser → Install Command → Validation
                                              ↓
                                         Git Clone
                                              ↓
                                    Validate Structure
                                              ↓
                                      Copy Files
                                              ↓
                                   Update Lock File
                                              ↓
                                    Success Output
```

### Update Command Flow

```
User Input → CLI Parser → Update Command → Read Lock File
                                              ↓
                                      Check for Updates
                                              ↓
                                   Clone New Version
                                              ↓
                                    Replace Files
                                              ↓
                                   Update Lock File
                                              ↓
                                    Success Output
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
│   │   └── command2/
│   └── config.yaml          # Optional config
├── ccmd.yaml                # Project configuration
├── ccmd-lock.yaml           # Lock file
└── src/                     # Project files
```

### Global Cache (Future)

```
~/.ccmd/
├── cache/                   # Downloaded repositories
│   ├── github.com/
│   │   └── user/
│   │       └── repo/
│   │           └── abc123/  # Commit hash
│   └── gitlab.com/
├── config.yaml             # Global config
└── registry.json           # Local registry cache
```

## Error Handling

### Sentinel Errors

ccmd uses sentinel errors for common error types:

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

### Error Flow

1. **Capture** - Errors are captured at origin using helper functions
2. **Wrap** - Add context with error creation functions
3. **Check** - Use `errors.Is()` to check error types
4. **Display** - Show appropriate message to user

Example:
```go
// Creating errors
if !exists {
    return errors.NotFound(fmt.Sprintf("command %s", name))
}

// Wrapping Git errors
if err := git.Clone(repo); err != nil {
    return errors.GitError("clone repository", err)
}

// Checking error types
if errors.Is(err, errors.ErrNotFound) {
    output.Warning("Command not found. Use 'ccmd search' to find available commands.")
    return nil
}
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

### Future Security Features

- Command signing with GPG
- Checksum verification
- Registry authentication
- Vulnerability scanning

## Performance

### Optimization Strategies

1. **Shallow Cloning**
   ```go
   git clone --depth 1 --single-branch
   ```

2. **Concurrent Operations**
   ```go
   // Update multiple commands concurrently
   var wg sync.WaitGroup
   for _, cmd := range commands {
       wg.Add(1)
       go func(c Command) {
           defer wg.Done()
           updateCommand(c)
       }(cmd)
   }
   wg.Wait()
   ```

3. **Caching**
   - Cache Git operations
   - Store registry data locally
   - Reuse cloned repositories

4. **Minimal File I/O**
   - Read files once
   - Batch write operations
   - Use buffered I/O

### Benchmarks

Key operations should complete within:
- Install: < 5 seconds
- List: < 100ms
- Remove: < 500ms
- Update: < 5 seconds per command

## Future Considerations

### Planned Features

1. **Global Command Registry**
   - Central repository of commands
   - Search and discovery
   - Ratings and reviews
   - Verified publishers

2. **Dependency Management**
   - Command dependencies
   - Version resolution
   - Dependency tree visualization

3. **Plugin System**
   - Extend ccmd functionality
   - Custom providers
   - Hook system

4. **Enhanced Security**
   - GPG signing
   - Integrity verification
   - Security advisories

### Architecture Evolution

1. **Service Registry Pattern**
   ```go
   type ServiceRegistry struct {
       git    git.Client
       fs     fs.Interface
       lock   lock.Manager
       cache  cache.Interface
   }
   ```

2. **Event System**
   ```go
   type EventBus interface {
       Subscribe(event string, handler func(data interface{}))
       Publish(event string, data interface{})
   }
   ```

3. **Provider Interface**
   ```go
   type Provider interface {
       Name() string
       Clone(url string) error
       GetMetadata(url string) (*Metadata, error)
   }
   ```

## Testing Strategy

### Unit Tests

- Test individual components in isolation
- Mock external dependencies
- Focus on business logic

### Integration Tests

- Test component interactions
- Use real file system in temp directories
- Verify end-to-end workflows

### Example Test Structure

```go
func TestInstallCommand(t *testing.T) {
    // Setup
    fs := memfs.New()
    git := mockgit.New()
    cmd := &InstallCommand{fs: fs, git: git}
    
    // Execute
    err := cmd.Execute("github.com/user/repo", Options{})
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, fs.Exists(".claude/commands/repo/index.md"))
}
```

## Conclusion

ccmd's architecture prioritizes simplicity, modularity, and extensibility. By following clear design principles and maintaining separation of concerns, the codebase remains maintainable and easy to extend with new features.
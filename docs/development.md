# Development Guide

This guide covers everything you need to know to develop and contribute to ccmd.

## Table of Contents

- [Environment Setup](#environment-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Building](#building)
- [Testing](#testing)
- [Debugging](#debugging)
- [Code Standards](#code-standards)
- [Adding New Commands](#adding-new-commands)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)

## Environment Setup

### Prerequisites

- **Go 1.23+**: [Download Go](https://go.dev/dl/)
- **Git**: [Download Git](https://git-scm.com/)
- **Make**: Usually pre-installed on Unix systems
- **golangci-lint**: For code linting

### Initial Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/gifflet/ccmd.git
   cd ccmd
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install development tools**
   ```bash
   # Install golangci-lint
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Install goimports
   go install golang.org/x/tools/cmd/goimports@latest
   
   # Install mockgen for generating mocks
   go install github.com/golang/mock/mockgen@latest
   ```

4. **Verify setup**
   ```bash
   make check-tools
   make test
   ```

### IDE Setup

#### VS Code

1. Install the Go extension
2. Add to `.vscode/settings.json`:
   ```json
   {
     "go.lintTool": "golangci-lint",
     "go.lintOnSave": "package",
     "go.formatTool": "goimports",
     "go.useLanguageServer": true,
     "gopls": {
       "staticcheck": true
     }
   }
   ```

#### GoLand/IntelliJ

1. Enable Go modules support
2. Configure golangci-lint as external linter
3. Set up file watchers for formatting

## Project Structure

```
ccmd/
├── cmd/                    # Command-line interfaces
│   ├── ccmd/              # Main CLI entry point
│   │   └── main.go
│   ├── install/           # Install command CLI
│   ├── list/              # List command CLI
│   ├── remove/            # Remove command CLI
│   ├── search/            # Search command CLI
│   ├── update/            # Update command CLI
│   ├── info/              # Info command CLI
│   ├── init/              # Init command CLI
│   └── sync/              # Sync command CLI
├── core/                  # Core business logic
│   ├── install.go         # Install logic
│   ├── list.go            # List logic
│   ├── remove.go          # Remove logic
│   ├── search.go          # Search logic
│   ├── update.go          # Update logic
│   ├── info.go            # Info logic
│   ├── init.go            # Init logic
│   ├── sync.go            # Sync logic
│   ├── git.go             # Git operations
│   ├── metadata.go        # Config/lock file management
│   └── types.go           # Shared types
├── pkg/                   # Public utilities
│   ├── errors/            # Error handling
│   ├── logger/            # Logging wrapper
│   └── output/            # Colored output, progress
├── internal/              # Private packages
│   └── fs/                # File system abstraction for tests
├── docs/                  # Documentation
├── examples/              # Example commands
├── scripts/               # Build and utility scripts
├── testdata/              # Test fixtures
├── tests/                 # Integration tests
├── .github/               # GitHub workflows
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── go.sum                 # Dependency checksums
```

### Package Responsibilities

- **cmd/**: CLI parsing with Cobra, flag handling, delegates to core
- **core/**: All business logic, Git operations, file management
- **pkg/errors**: Consistent error handling with sentinel errors
- **pkg/logger**: Convenient wrapper over slog
- **pkg/output**: Colored output, progress bars, spinners
- **internal/fs**: FileSystem interface for testing
- **scripts/**: Build automation, release scripts
- **testdata/**: Test fixtures and mock data

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-feature
```

### 2. Make Changes

Follow the TDD approach:

1. Write tests first
2. Implement functionality
3. Refactor if needed

### 3. Run Tests and Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run all checks
make check
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with conventional commit message
git commit -m "feat(install): add support for private repositories"
```

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/my-feature

# Create PR on GitHub
```

## Building

### Quick Build

```bash
# Build for current platform
make build

# Run the built binary
./bin/ccmd --help
```

### Cross-Platform Build

```bash
# Build for all platforms
make build-all

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
```

### Release Build

```bash
# Create release builds with version info
make release
```

### Build Flags

The build system sets these flags:

```go
-X main.version=$(VERSION)
-X main.commit=$(COMMIT)
-X main.date=$(DATE)
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with coverage
make test-coverage

# Run specific package tests
go test ./core/...

# Run specific test
go test -run TestInstall ./core/
```

### Writing Tests

#### Unit Test Example

```go
func TestParseRepository(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Repository
        wantErr bool
    }{
        {
            name:  "github https url",
            input: "https://github.com/user/repo.git",
            want: Repository{
                Host:  "github.com",
                Owner: "user",
                Name:  "repo",
            },
            wantErr: false,
        },
        {
            name:    "invalid url",
            input:   "not-a-url",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseRepository(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseRepository() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseRepository() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Integration Test Example

```go
func TestInstallCommandIntegration(t *testing.T) {
    // Skip in short mode
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Create temp directory
    tmpDir := t.TempDir()
    
    // Set up test environment
    fs := filesystem.NewReal()
    git := gitclient.New()
    
    // Create command
    cmd := &InstallCommand{
        fs:     fs,
        git:    git,
        config: Config{BaseDir: tmpDir},
    }
    
    // Execute
    err := cmd.Execute("github.com/test/repo", Options{})
    
    // Verify
    assert.NoError(t, err)
    assert.True(t, fs.Exists(filepath.Join(tmpDir, "commands", "repo")))
}
```

### Mock Generation

```bash
# Generate mocks for interfaces
mockgen -source=pkg/git/client.go -destination=pkg/git/mock_client.go -package=git

# Use in tests
func TestWithMockGit(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockGit := NewMockClient(ctrl)
    mockGit.EXPECT().Clone("repo-url", gomock.Any()).Return(nil)
    
    // Use mockGit in test
}
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

Aim for:
- 80%+ overall coverage
- 90%+ for critical paths
- 100% for error handling

## Debugging

### Debug Output

```bash
# Enable debug logging
CCMD_DEBUG=1 ccmd install github.com/user/repo

# Or use verbose flag
ccmd install github.com/user/repo -v
```

### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd/ccmd -- install github.com/user/repo

# Set breakpoints
(dlv) break main.main
(dlv) break core.Install
(dlv) continue
```

### VS Code Debugging

Add to `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug ccmd",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/ccmd",
            "args": ["install", "github.com/user/repo"],
            "env": {
                "CCMD_DEBUG": "1"
            }
        }
    ]
}
```

### Common Debugging Techniques

1. **Print debugging**
   ```go
   log.Printf("DEBUG: repo=%s, opts=%+v", repo, opts)
   ```

2. **Error wrapping**
   ```go
   if err != nil {
       return fmt.Errorf("at %s: %w", debug.Stack(), err)
   }
   ```

3. **Panic recovery**
   ```go
   defer func() {
       if r := recover(); r != nil {
           log.Printf("Panic: %v\n%s", r, debug.Stack())
       }
   }()
   ```

## Code Standards

### Go Code Style

Follow standard Go conventions:

1. **Naming**
   - Use camelCase for variables and functions
   - Use PascalCase for exported types
   - Use descriptive names

2. **Comments**
   - Add godoc comments to all exported items
   - Start with the name being declared
   - Use complete sentences

3. **Error Handling**
   ```go
   // Always check errors
   if err := doSomething(); err != nil {
       return fmt.Errorf("do something: %w", err)
   }
   
   // Custom errors
   var ErrNotFound = errors.New("command not found")
   
   // Error types
   type ValidationError struct {
       Field string
       Value string
   }
   ```

4. **Interfaces**
   ```go
   // Small, focused interfaces
   type Reader interface {
       Read([]byte) (int, error)
   }
   
   // Accept interfaces, return structs
   func NewService(r Reader) *Service {
       return &Service{reader: r}
   }
   ```

### Linting Rules

Our `.golangci.yml` configuration enforces:

- No unused variables
- No inefficient assignments
- Proper error checking
- Consistent formatting
- Cyclomatic complexity limits

Run linter before committing:
```bash
golangci-lint run
```

## Adding New Commands

### 1. Create Command Structure

```bash
# Create new command directory
mkdir cmd/newcmd

# Create command file
touch cmd/newcmd/newcmd.go
```

### 2. Implement Command Interface

```go
// cmd/newcmd/newcmd.go
package newcmd

import (
    "github.com/urfave/cli/v2"
    "github.com/gifflet/ccmd/core"
)

func Command() *cli.Command {
    return &cli.Command{
        Name:      "newcmd",
        Usage:     "Brief description",
        ArgsUsage: "[arguments]",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "option",
                Aliases: []string{"o"},
                Usage:   "Option description",
            },
        },
        Action: run,
    }
}

func run(c *cli.Context) error {
    cmd := commands.NewNewCmd()
    return cmd.Execute(c.Args().First(), commands.NewCmdOptions{
        Option: c.String("option"),
    })
}
```

### 3. Implement Business Logic

```go
// core/newcmd.go
package commands

type NewCmd struct {
    fs     fs.Interface
    output output.Interface
}

func NewNewCmd() *NewCmd {
    return &NewCmd{
        fs:     filesystem.NewReal(),
        output: output.NewConsole(),
    }
}

func (c *NewCmd) Execute(arg string, opts NewCmdOptions) error {
    // Implement command logic
    return nil
}
```

### 4. Add Tests

```go
// core/newcmd_test.go
package commands

import "testing"

func TestNewCmd(t *testing.T) {
    // Test implementation
}
```

### 5. Register Command

Add to `cmd/ccmd/main.go`:

```go
import "github.com/gifflet/ccmd/cmd/newcmd"

app.Commands = []*cli.Command{
    install.Command(),
    list.Command(),
    newcmd.Command(), // Add here
}
```

## Common Tasks

### Update Dependencies

```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get -u github.com/urfave/cli/v2
```

### Generate Mocks

```bash
# Generate all mocks
make generate

# Generate specific mock
mockgen -source=internal/fs/interface.go -destination=internal/fs/mock_fs.go -package=fs
```

### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkInstall ./core/

# With memory allocation stats
go test -bench=. -benchmem ./...
```

### Profile Performance

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./core/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./core/
go tool pprof mem.prof
```

## Troubleshooting

### Common Issues

1. **Module download failures**
   ```bash
   # Clear module cache
   go clean -modcache
   
   # Re-download
   go mod download
   ```

2. **Build failures**
   ```bash
   # Clean build cache
   go clean -cache
   
   # Rebuild
   make clean build
   ```

3. **Test failures**
   ```bash
   # Run tests verbosely
   go test -v ./...
   
   # Run with race detector
   go test -race ./...
   ```

4. **Linter errors**
   ```bash
   # See detailed errors
   golangci-lint run -v
   
   # Auto-fix some issues
   golangci-lint run --fix
   ```

### Debug Tips

1. **Check Go version**
   ```bash
   go version  # Should be 1.23+
   ```

2. **Verify module mode**
   ```bash
   go env GO111MODULE  # Should be "on" or ""
   ```

3. **Check for conflicts**
   ```bash
   go mod graph | grep conflict
   ```

4. **Inspect dependencies**
   ```bash
   go mod why github.com/some/package
   ```

### Getting Help

- Check existing issues on GitHub
- Ask in discussions
- Review test cases for examples
- Read Go documentation

## Performance Tips

1. **Avoid allocations in hot paths**
   ```go
   // Bad
   func process(items []string) []string {
       result := []string{}
       for _, item := range items {
           result = append(result, transform(item))
       }
       return result
   }
   
   // Good
   func process(items []string) []string {
       result := make([]string, 0, len(items))
       for _, item := range items {
           result = append(result, transform(item))
       }
       return result
   }
   ```

2. **Use sync.Pool for temporary objects**
   ```go
   var bufferPool = sync.Pool{
       New: func() interface{} {
           return new(bytes.Buffer)
       },
   }
   ```

3. **Profile before optimizing**
   - Use benchmarks to measure
   - Profile to find bottlenecks
   - Optimize only what matters

## Release Process

### Creating a Release

1. **Update version**
   ```bash
   # Update version in code
   # Create changelog entry
   ```

2. **Create tag**
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

3. **GitHub Actions handles**
   - Building for all platforms
   - Creating GitHub release
   - Uploading artifacts
   - Updating Homebrew formula

### Manual Release (if needed)

```bash
# Build release artifacts
make release

# Upload to GitHub
gh release create v1.2.3 ./dist/*
```
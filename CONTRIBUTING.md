# Contributing to ccmd

Thank you for your interest in contributing to ccmd! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- **Be respectful** - Treat everyone with respect and kindness
- **Be inclusive** - Welcome people of all backgrounds and identities
- **Be collaborative** - Work together to resolve conflicts and assume good intentions
- **Be professional** - Inappropriate behavior is not tolerated

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Git
- Make (optional but recommended)
- A GitHub account

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ccmd.git
   cd ccmd
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/gifflet/ccmd.git
   ```

### Development Setup

```bash
# Install dependencies
go mod download

# Run tests to ensure everything works
make test

# Build the project
make build
```

## Development Workflow

### 1. Create a Branch

Create a feature branch from `main`:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 2. Make Your Changes

- Write code following our [coding standards](#coding-standards)
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Commit Guidelines

We follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks

Examples:
```bash
git commit -m "feat(install): add support for GitLab repositories"
git commit -m "fix(lock): handle concurrent access properly"
git commit -m "docs(readme): update installation instructions"
```

### 4. Keep Your Branch Updated

```bash
git fetch upstream
git rebase upstream/main
```

### 5. Run Tests and Checks

Before submitting:

```bash
# Format code
make fmt

# Run linter
make lint

# Run all tests
make test

# Run all checks
make check
```

## Submitting Changes

### Pull Request Process

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a Pull Request on GitHub

3. Fill out the PR template:
   - Describe what the PR does
   - Reference any related issues
   - List any breaking changes
   - Include testing instructions

4. Wait for review:
   - All PRs require at least one review
   - Address feedback promptly
   - Keep the PR updated with main

### PR Requirements

- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] No merge conflicts with main

## Coding Standards

### Go Code Style

We follow standard Go conventions:

```go
// Package comments should be present
package commands

// ExportedFunction should have a comment starting with its name
func ExportedFunction() error {
    // Use meaningful variable names
    configPath := getConfigPath()
    
    // Handle errors explicitly
    if err := validatePath(configPath); err != nil {
        return errors.InvalidInput(fmt.Sprintf("config path %s", configPath))
    }
    
    return nil
}
```

### Error Handling

- Always check errors
- Use the `pkg/errors` package for consistent error handling
- Choose the appropriate error function based on the error type

```go
import "github.com/gifflet/ccmd/pkg/errors"

// For resource not found
if !exists {
    return errors.NotFound("command foo")
}

// For git operations
if err := git.Clone(repo); err != nil {
    return errors.GitError("clone", err)
}

// For file operations
if err := os.ReadFile(path); err != nil {
    return errors.FileError("read", path, err)
}
```

### Testing

- Write table-driven tests
- Use meaningful test names
- Mock external dependencies
- Aim for >80% code coverage

```go
func TestInstallCommand(t *testing.T) {
    tests := []struct {
        name    string
        repo    string
        wantErr bool
    }{
        {
            name:    "valid GitHub repository",
            repo:    "github.com/user/repo",
            wantErr: false,
        },
        {
            name:    "invalid repository URL",
            repo:    "not-a-url",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Install(tt.repo)
            if (err != nil) != tt.wantErr {
                t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Documentation

### Code Documentation

- Add godoc comments to all exported types and functions
- Include examples in doc comments when helpful
- Keep comments concise and clear

### User Documentation

When adding new features:

1. Update the README.md if needed
2. Add examples to the `examples/` directory
3. Update command help text
4. Add to the documentation in `docs/`

### Example Documentation

Create example files showing real-world usage:

```markdown
# examples/gitlab-install.md

# Installing Commands from GitLab

ccmd supports installing commands from GitLab repositories:

\```bash
# Public repository
ccmd install gitlab.com/user/my-command

# Private repository (requires authentication)
ccmd install gitlab.com/org/private-command
\```
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./core/...

# Run with race detector
go test -race ./...

# Run integration tests
make test-integration
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Mock external dependencies (filesystem, git, network)
- Test both success and error cases

### Test Coverage

We aim for high test coverage:
- New features should have >80% coverage
- Critical paths should have >90% coverage
- Use `make test-coverage` to check

## Project Structure

```
ccmd/
├── cmd/               # Command-line applications
│   ├── ccmd/         # Main CLI application
│   └── */            # Individual command implementations
├── pkg/              # Public packages
│   ├── commands/     # Command implementations
│   └── git/          # Git operations
├── internal/         # Private packages
│   ├── fs/          # Filesystem abstraction
│   ├── lock/        # Lock file management
│   ├── models/      # Data models
│   └── output/      # Output formatting
├── docs/            # Documentation
├── examples/        # Usage examples
└── scripts/         # Build and utility scripts
```

## Debugging

### Debug Output

Use the `-v` or `--verbose` flag for debug output:

```bash
ccmd install github.com/user/repo -v
```

### Common Issues

1. **Import cycles**: Keep dependencies unidirectional
2. **Race conditions**: Use proper locking for concurrent access
3. **File permissions**: Test with different permission scenarios

## Release Process

Releases are automated when tags are pushed:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

This triggers:
1. Build for all platforms
2. Run tests
3. Create GitHub release
4. Upload binaries
5. Update Homebrew formula

## Getting Help

- **Questions**: Use [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)
- **Bugs**: Open an [issue](https://github.com/gifflet/ccmd/issues)
- **Security**: Email security@ccmd.dev

## Community

- Follow our [Twitter](https://twitter.com/ccmd_dev)
- Join our [Discord](https://discord.gg/ccmd)
- Read our [blog](https://blog.ccmd.dev)

Thank you for contributing to ccmd! Your efforts help make Claude Code more powerful for everyone.
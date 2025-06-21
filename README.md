# ccmd - Claude Command Manager

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gifflet/ccmd)](https://goreportcard.com/report/github.com/gifflet/ccmd)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A simple and powerful command-line tool for managing custom commands in Claude Code. Install, update, and share commands from Git repositories with the ease of a package manager.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Creating Commands](#creating-commands)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Overview

ccmd (Claude Command Manager) brings package management capabilities to Claude Code, allowing you to:
- Easily install commands from Git repositories
- Keep commands updated with simple commands
- Share your own commands with the community
- Manage project-specific command configurations

## Features

- 📦 **Easy Installation** - Install commands from any Git repository with a single command
- 🔄 **Version Management** - Pin specific versions or update to the latest
- 🔍 **Command Discovery** - Search for commands in the community registry
- 🔒 **Lock File Support** - Reproducible installs across team members
- 🚀 **Fast & Lightweight** - Minimal dependencies, maximum performance
- 📝 **Project Scoped** - Commands are installed per project, not globally
- 🛡️ **Safe Updates** - Update with confidence using lock files

## Installation

### Using Homebrew (macOS/Linux)

```bash
brew tap gifflet/ccmd
brew install ccmd
```

### Using Go

```bash
go install github.com/gifflet/ccmd/cmd/ccmd@latest
```

### Download Binary

Download the latest binary for your platform from the [releases page](https://github.com/gifflet/ccmd/releases).

```bash
# Example for macOS (Apple Silicon)
curl -L https://github.com/gifflet/ccmd/releases/latest/download/ccmd-darwin-arm64.tar.gz | tar xz
sudo mv ccmd /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/gifflet/ccmd.git
cd ccmd
make build
sudo make install
```

## Quick Start

### 1. Initialize ccmd in your project

```bash
cd your-project
ccmd init
```

This interactive command guides you through creating a `ccmd.yaml` file for your Claude Code Command project. It will prompt you for metadata about your command and create the necessary `.claude/commands` directory structure.

### 2. Install a command

```bash
# Install from GitHub
ccmd install github.com/user/my-command

# Install specific version
ccmd install github.com/user/my-command@v1.0.0

# Install with custom name
ccmd install github.com/user/my-command --name mc
```

### 3. List installed commands

```bash
ccmd list
```

### 4. Update commands

```bash
# Update specific command
ccmd update my-command

# Update all commands
ccmd update --all
```

### 5. Remove a command

```bash
ccmd remove my-command
```

## Commands

### `ccmd init`
Initialize a new Claude Code Command project by creating the necessary configuration files and directory structure.

```bash
ccmd init
```

This interactive command guides you through setting up a new ccmd project:

- Prompts for command metadata (name, version, description, author, repository, etc.)
- Uses sensible defaults (current directory name for `name`, `1.0.0` for version, `index.md` for entry)
- Shows a preview of the generated `ccmd.yaml` before creating it
- Creates the `.claude/commands` directory structure
- All fields except `name` and `version` are optional

Example session:
```
$ ccmd init
This utility will walk you through creating a ccmd.yaml file.
Press ^C at any time to quit.

name: (my-project) my-command
version: (1.0.0) 
description: A command to automate tasks
author: John Doe
repository: https://github.com/johndoe/my-command
entry: (index.md) 
tags (comma-separated): automation, productivity

About to write to /path/to/ccmd.yaml:

name: my-command
version: 1.0.0
description: A command to automate tasks
author: John Doe
repository: https://github.com/johndoe/my-command
entry: index.md
tags:
  - automation
  - productivity

Is this OK? (yes) 

✓ Created .claude/commands directory
✓ Created ccmd.yaml

Your ccmd project has been initialized!
```

### `ccmd install`
Install a command from a Git repository.

```bash
# Basic usage
ccmd install <repository>

# With version
ccmd install <repository>@<version>

# Options
ccmd install github.com/user/cmd --name custom-name
ccmd install github.com/user/cmd --force  # Reinstall
```

### `ccmd list`
List all installed commands.

```bash
# Basic list
ccmd list

# Detailed output
ccmd list --long

# JSON output
ccmd list --json
```

### `ccmd update`
Update installed commands to their latest versions.

```bash
# Update specific command
ccmd update <command-name>

# Update all commands
ccmd update --all

# Update to specific version
ccmd update <command-name>@v2.0.0
```

### `ccmd remove`
Remove an installed command.

```bash
ccmd remove <command-name>

# Remove multiple
ccmd remove cmd1 cmd2 cmd3
```

### `ccmd search`
Search for commands in the registry.

```bash
# Search by keyword
ccmd search automation

# Search with filters
ccmd search --tags testing,ci
ccmd search --author gifflet
```

### `ccmd info`
Display detailed information about a command.

```bash
# Show info for installed command
ccmd info <command-name>

# Show info from repository
ccmd info github.com/user/command
```

### `ccmd run`
Run an installed command (useful for testing).

```bash
ccmd run <command-name> [args...]
```

## Creating Commands

Creating a command for ccmd is simple. Start by running `ccmd init` in your command directory to interactively create the configuration file. Your repository needs:

1. **ccmd.yaml** - Command metadata (created by `ccmd init`)
2. **index.md** - Command instructions for Claude
3. **README.md** - Documentation for users

### Example Command Structure

```
my-awesome-command/
├── ccmd.yaml          # Command metadata (required)
├── index.md           # Command for Claude (required)
├── README.md          # User documentation
├── examples/          # Usage examples
│   └── example.md
└── LICENSE           # License file
```

### ccmd.yaml Format

```yaml
name: my-awesome-command
version: 1.0.0
description: Automates awesome tasks in Claude Code
author: Your Name
repository: https://github.com/username/my-awesome-command
entry: index.md  # Optional, defaults to index.md
tags:
  - automation
  - productivity
  - testing
commands:  # Optional
  - other-command@^1.0.0
```

### index.md Example

```markdown
# My Awesome Command

You are an AI assistant helping with task automation. When the user invokes this command, you should:

## Instructions

1. Analyze the current project structure
2. Generate appropriate configuration files
3. Set up the automation pipeline
4. Provide clear feedback to the user

## Parameters

- `--type`: Type of automation (test, build, deploy)
- `--config`: Path to configuration file

## Examples

When user says "setup automation", you should...
```

See our [Command Creation Guide](docs/command-structure.md) for detailed instructions.

## Configuration

ccmd stores all data in your project's `.claude` directory:

```
your-project/
├── .claude/
│   ├── commands/          # Installed command files
│   ├── {command}.md       # Command metadata
├── src/
└── ...
```

## Development

### Prerequisites

- Go 1.23 or higher
- Git
- Make (optional but recommended)

### Setup

```bash
# Clone repository
git clone https://github.com/gifflet/ccmd.git
cd ccmd

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally
./bin/ccmd --help
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./pkg/commands/...
```

### Code Style

We use standard Go formatting and linting:

```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- Code of Conduct
- Development workflow
- Submitting pull requests
- Reporting issues

## Roadmap

- [x] Core command management (install, update, remove)
- [x] Git repository support
- [x] Version management
- [x] Lock file support
- [x] Interactive project initialization (`ccmd init`)
- [ ] Official command registry
- [ ] Command dependencies
- [ ] Command signing and verification
- [ ] Plugin system for extending ccmd
- [ ] Global command installation option

## Community

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issues**: [GitHub Issues](https://github.com/gifflet/ccmd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)
- **Discord**: [Join our Discord](https://discord.gg/ccmd)

## License

ccmd is released under the MIT License. See [LICENSE](LICENSE) for details.
# ccmd - Claude Command Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A simple command-line utility for managing custom commands in Claude Code, inspired by npm's package management approach.

## Overview

ccmd (Claude Command Manager) allows you to install, update, and manage custom commands from Git repositories, making it easy to extend Claude Code's functionality with community-created tools.

## Features

- 📦 **Install** commands from Git repository
- 🔄 **Update** commands to their latest versions
- 🗑️ **Remove** commands when no longer needed
- 📋 **List** all installed commands
- 🔍 **Search** for commands in the registry
- ⚡ **Simple** and fast, following the KISS principle

## Installation

### Using Go

```bash
go install github.com/gifflet/ccmd@latest
```

### From Source

```bash
git clone https://github.com/gifflet/ccmd.git
cd ccmd
go build -o ccmd cmd/ccmd/main.go
sudo mv ccmd /usr/local/bin/
```

### Using Homebrew (macOS)

```bash
brew tap gifflet/ccmd
brew install ccmd
```

## Quick Start

### Install a command

```bash
# Install from a Git repository
ccmd install github.com/user/command-repo

# Install with a specific version/tag
ccmd install github.com/user/command-repo@v1.2.0

# Install from any git source
ccmd install https://gitlab.com/user/repo
ccmd install git@github.com:user/repo.git
```

### Update commands

```bash
# Update a specific command
ccmd update command-name

# Update all commands
ccmd update --all
```

### Remove a command

```bash
ccmd remove command-name
```

### List installed commands

```bash
ccmd list
```

### Search for commands

```bash
ccmd search keyword
```

## Command Structure

Commands are markdown files that provide instructions for Claude Code. Each command repository should follow this structure:

```
command-repo/
├── ccmd.yaml          # Command metadata
├── index.md           # The actual command (markdown file)
├── README.md          # Command documentation
└── LICENSE            # License file
```

### ccmd.yaml Example

```yaml
name: my-command
version: 1.0.0
description: A useful command for Claude Code
author: Your Name
repository: https://github.com/user/command-repo
entry: index.md  # Optional, defaults to index.md
tags:
  - automation
  - testing
  - development
```

### index.md Example

```markdown
# My Command

This command helps you automate X task in Claude Code.

## Usage

1. First, do this...
2. Then, do that...
3. Finally, complete with...

## Examples

Example of using this command...
```

## Configuration

ccmd installs commands locally to your project:

```
my-project/
├── .claude/
│   ├── commands/       # Installed command files
│   └── commands.lock   # Lock file tracking installed commands
└── src/
```

No global configuration is needed - ccmd works entirely at the project level.

## Creating Commands

To create a command compatible with ccmd:

1. Create a new Git repository
2. Add a `ccmd.yaml` file with metadata
3. Write your command instructions in `index.md` (or custom name)
4. Add documentation in `README.md`
5. Push to GitHub/GitLab/etc
6. Share with the community!

See our [Command Creation Guide](docs/CREATING_COMMANDS.md) for detailed instructions.

## Architecture

ccmd follows a simple architecture:

```
ccmd (CLI)
  ├── Git Client (fetch from repositories)
  ├── Command Manager (install/update/remove markdown files)
  └── Project Storage (.claude/)
    ├── commands/        # Installed command files
    └── commands.lock    # Lock file for tracking
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/gifflet/ccmd.git
cd ccmd

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o ccmd cmd/ccmd/main.go
```

## Roadmap

- [x] Basic install/update/remove functionality
- [x] Git repository support
- [ ] Command registry/discovery
- [ ] Command dependencies
- [ ] Command verification/signing
- [ ] Auto-update functionality
- [ ] Command templates

## Community

- **Issues**: [GitHub Issues](https://github.com/gifflet/ccmd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built for the [Claude Code](https://claude.ai/code) community

---
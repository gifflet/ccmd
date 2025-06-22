# ccmd - Claude Command Manager

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gifflet/ccmd)](https://goreportcard.com/report/github.com/gifflet/ccmd)

Simple command-line tool for managing custom commands in Claude Code. Install and share commands from Git repositories with the ease of a package manager.

## Why ccmd?

Managing custom Claude Code commands across multiple projects can be challenging. ccmd solves this by treating commands as versioned, reusable packages:

- **Keep commands out of your codebase**: Store command definitions (.md files and AI context) in separate repositories, keeping your project repositories clean
- **Version control**: Each command has its own version, allowing you to use different versions in different projects
- **Reusability**: Install the same command in multiple projects without duplication
- **Easy sharing**: Share commands with your team or the community through Git repositories
- **Simple management**: Install, update, and remove commands with familiar package manager semantics

Think of ccmd as "npm for Claude Code commands" - centralize your AI tooling configurations and use them anywhere.

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

> For other installation methods, see [Installation Guide](docs/installation.md)

## Quick Start

### 1. Initialize your project
```bash
cd your-project
ccmd init
```

### 2. Install a demo command
```bash
ccmd install https://github.com/gifflet/hello-world
```

### 3. Use it in Claude Code
```
/hello-world
```

That's it! You've just installed and used your first ccmd command.

## Commands

| Command | Description |
|---------|-------------|
| `ccmd init` | Initialize a new command project |
| `ccmd install <repo>` | Install a command from a Git repository |
| `ccmd install` | Install all commands from ccmd.yaml |
| `ccmd list` | List installed commands |
| `ccmd update <command>` | Update a specific command |
| `ccmd remove <command>` | Remove an installed command |
| `ccmd search <keyword>` | Search for commands in the registry |
| `ccmd info <command>` | Show detailed command information |

> For detailed usage and options, see [Command Reference](docs/commands.md)

## Creating Your Own Commands

Creating a command for ccmd is simple. Your repository needs:

1. **ccmd.yaml** - Command metadata (created by `ccmd init`)
2. **index.md** - Command instructions for Claude

### Quick Start

```bash
mkdir my-command && cd my-command
ccmd init  # Creates ccmd.yaml interactively
```

### Example Structure

```
my-command/
├── ccmd.yaml          # Command metadata (required)
└── index.md           # Command for Claude (required)
```

### Example ccmd.yaml

```yaml
name: my-command
version: 1.0.0
description: Automates tasks in Claude Code
author: Your Name
repository: https://github.com/username/my-command
entry: index.md  # Optional, defaults to index.md
```

> For complete guide with examples, see [Creating Commands](docs/creating-commands.md)

## Example Commands

Here are some commands you can install and try:

- **hello-world**: Simple demo command
  ```bash
  ccmd install https://github.com/gifflet/hello-world
  ```

## Documentation

- **[Full Documentation](docs/)** - Complete guides and references
- **[Command Creation Guide](docs/creating-commands.md)** - Create your own commands

## Community

- **Issues**: [GitHub Issues](https://github.com/gifflet/ccmd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT License - see [LICENSE](LICENSE) for details
---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 10
type: docs
description: >
  Get up and running with ccmd in minutes. Learn how to install, configure, and use your first Claude Code commands.
keywords: ["installation", "setup", "Claude Code setup", "quickstart", "slash commands tutorial"]
---

## Prerequisites

Before you begin, ensure you have:

- **Node.js** (v16 or higher) or **Go** (v1.23 or higher)
- **Git** installed and configured
- **Claude Code** installed on your system
- A GitHub account (for creating and sharing commands)

## Installation

Choose your preferred installation method:

### Via NPM (Recommended)

```bash
npm install -g @gifflet/ccmd
```

### Via Go

```bash
go install github.com/gifflet/ccmd/cmd/ccmd@latest
```

### Verify Installation

After installation, verify ccmd is working:

```bash
ccmd --version
```

## Quick Start

### 1. Initialize Your Project

Navigate to your project directory and initialize ccmd:

```bash
cd your-project
ccmd init
```

This creates the necessary directory structure for managing Claude Code commands.

### 2. Install a Demo Command

Let's install a simple hello-world command to test:

```bash
ccmd install gifflet/hello-world
```

### 3. Use the Command in Claude Code

Open Claude Code in your project and type:

```
/hello-world
```

The command will execute and provide its functionality!

## Understanding ccmd

### Project Structure

After initialization, your project will have:

```
your-project/
├── .claude/
│   └── commands/       # Installed commands live here
├── ccmd.yaml          # Project configuration
└── ccmd-lock.yaml     # Lock file (like package-lock.json)
```

### Configuration Files

**ccmd.yaml** - Defines your project's command dependencies:
```yaml
name: my-project
version: 1.0.0
description: My awesome project
commands:
  - github.com/user/command1
  - github.com/user/command2@v1.2.0
```

**ccmd-lock.yaml** - Locks specific versions for reproducible installs:
```yaml
version: 1
commands:
  command1:
    version: v1.0.0
    repository: https://github.com/user/command1
    installed_at: "2024-01-15T10:30:00Z"
```

## Common Workflows

### Installing Commands

Install from various sources:

```bash
# Install latest version
ccmd install github.com/user/repo

# Install specific version
ccmd install github.com/user/repo@v1.0.0

# Install with custom name
ccmd install github.com/user/repo --name mycommand

# Install all from ccmd.yaml
ccmd install
```

### Managing Commands

```bash
# List installed commands
ccmd list

# Update a command
ccmd update command-name

# Remove a command
ccmd remove command-name

# Search installed commands
ccmd search keyword
```

### Sharing Your Setup

To share your command configuration with your team:

1. Commit `ccmd.yaml` and `ccmd-lock.yaml` to version control
2. Team members clone the repository
3. They run `ccmd install` to get all commands

## Best Practices

1. **Version Control**: Always commit `ccmd.yaml` and `ccmd-lock.yaml`
2. **Semantic Versioning**: Use specific versions for production projects
3. **Documentation**: Document which commands your project uses
4. **Regular Updates**: Keep commands updated with `ccmd update`

## Troubleshooting

### Command not found after installation

- Run `ccmd list` to verify installation
- Check `.claude/commands/` directory
- Ensure Claude Code is restarted after installation

### Installation fails

- Verify you have access to the repository
- Check your internet connection
- Ensure Git is properly configured

### Commands not working in Claude Code

- Verify the command is properly installed with `ccmd info command-name`
- Check the command's documentation for usage instructions
- Ensure you're using the correct syntax (e.g., `/command-name`)

## Next Steps

Now that you have ccmd installed and working:

- [Explore available commands](/usage/commands/) in the command reference
- [Create your own commands](/creating-commands/) to share with others
- [Browse examples](/examples/) to see real-world use cases
- [Join the community](https://github.com/gifflet/ccmd/discussions) to share and discover commands
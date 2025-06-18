# Install Command Examples

This document demonstrates how to use the `ccmd install` command.

## Basic Installation

Install a command from a Git repository:

```bash
# Install latest version
ccmd install github.com/user/my-command

# Install from HTTPS URL
ccmd install https://github.com/user/my-command.git

# Install from SSH URL
ccmd install git@github.com:user/my-command.git
```

## Version Specification

Install a specific version or tag:

```bash
# Install specific version tag
ccmd install github.com/user/my-command@v1.0.0

# Install specific commit
ccmd install github.com/user/my-command@abc123

# Using --version flag
ccmd install github.com/user/my-command --version v2.1.0
```

## Custom Command Name

Override the default command name:

```bash
# Install with custom name
ccmd install github.com/user/my-command --name mycommand

# Useful for installing multiple versions
ccmd install github.com/user/tool@v1.0.0 --name tool-v1
ccmd install github.com/user/tool@v2.0.0 --name tool-v2
```

## Force Reinstall

Force reinstall an existing command:

```bash
# Force reinstall (overwrites existing)
ccmd install github.com/user/my-command --force

# Update to latest version
ccmd install github.com/user/my-command@latest --force
```

## Command Structure Requirements

For a repository to be installable, it must have:

1. **ccmd.yaml** - Command metadata file
2. **index.md** - Command documentation

Example `ccmd.yaml`:

```yaml
name: my-command
version: 1.0.0
description: A useful command for doing things
author: Your Name
repository: https://github.com/user/my-command
license: MIT
tags:
  - utility
  - productivity
```

## Installation Process

When you install a command, ccmd:

1. Validates the repository is accessible
2. Clones the repository (shallow clone for efficiency)
3. Validates the command structure (ccmd.yaml and index.md)
4. Copies files to `~/.config/ccmd/commands/<command-name>/`
5. Creates a standalone `.md` file for quick access
6. Updates the lock file with installation details

## Error Handling

Common errors and solutions:

```bash
# Repository not found
ccmd install github.com/invalid/repo
# Error: repository not accessible

# Missing ccmd.yaml
ccmd install github.com/regular/repo
# Error: ccmd.yaml not found

# Command already exists
ccmd install github.com/user/existing-command
# Error: command 'existing-command' already exists (use --force to reinstall)
```

## After Installation

Once installed, you can:

```bash
# Use the command
ccmd my-command

# List installed commands
ccmd list

# Get command info
ccmd info my-command

# Remove the command
ccmd remove my-command
```
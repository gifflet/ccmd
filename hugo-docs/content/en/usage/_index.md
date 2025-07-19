---
title: "Command Reference"
linkTitle: "Commands"
weight: 20
type: docs
description: >
  Complete reference for all ccmd commands. Learn how to install, manage, and use Claude Code slash commands effectively.
keywords: ["ccmd commands", "command reference", "Claude Code CLI", "slash commands reference", "AI development tools"]
---

This document provides detailed information about all ccmd commands, their options, and usage examples.

## Table of Contents

- [Overview](#overview)
- [Global Options](#global-options)
- Commands:
  - [ccmd init](#ccmd-init)
  - [ccmd install](#ccmd-install)
  - [ccmd list](#ccmd-list)
  - [ccmd update](#ccmd-update)
  - [ccmd remove](#ccmd-remove)
  - [ccmd search](#ccmd-search)
  - [ccmd info](#ccmd-info)
  - [ccmd sync](#ccmd-sync)

## Overview

ccmd is a command-line tool for managing Claude Code commands. It provides a package manager-like experience for installing, updating, and managing custom commands from Git repositories.

### Command Structure

All ccmd commands follow this general pattern:

```bash
ccmd <command> [arguments] [flags]
```

### Getting Help

To see available commands:
```bash
ccmd --help
```

To see help for a specific command:
```bash
ccmd <command> --help
```

## Global Options

These options are available for all ccmd commands:

- `--version` - Display ccmd version information
- `--help` - Display help information

## ccmd init

Initialize a new Claude Code Command project by creating the necessary configuration files and directory structure.

### Usage

```bash
ccmd init
```

### Description

This interactive command guides you through setting up a new ccmd project. It prompts for essential metadata about your command and generates:
- `ccmd.yaml` - Command configuration file
- `.claude/commands/` - Directory structure for commands

### Options

This command has no additional flags. It runs interactively.

### Interactive Prompts

- **name**: Command name (defaults to current directory name)
- **version**: Semantic version (defaults to "1.0.0")
- **description**: Brief description of what your command does
- **author**: Your name or organization
- **repository**: Git repository URL
- **entry**: Entry point file (defaults to "index.md")
- **tags**: Comma-separated list of tags

### Examples

```bash
# Initialize a new command project
cd my-command
ccmd init

# Example interaction:
# name: (my-command) 
# version: (1.0.0) 
# description: Automates common development tasks
# author: Jane Doe
# repository: https://github.com/janedoe/my-command
# entry: (index.md) 
# tags (comma-separated): automation, dev-tools
```

### Notes

- If a `ccmd.yaml` file already exists, it will load existing values as defaults
- The command creates the `.claude/commands` directory structure automatically
- After initialization, create your `index.md` file with command instructions

## ccmd install

Install a command from a Git repository or install all commands from ccmd.yaml.

### Usage

```bash
ccmd install [repository] [flags]
```

### Description

When no repository is provided, installs all commands defined in the project's ccmd.yaml file. When a repository is provided, installs the command and adds it to ccmd.yaml and ccmd-lock.yaml.

### Options

- `-v, --version <version>` - Version/tag to install (defaults to latest)
- `-n, --name <name>` - Override command name
- `-f, --force` - Force reinstall if already exists

### Examples

```bash
# Install all commands from ccmd.yaml
ccmd install

# Install latest version of a command
ccmd install github.com/user/repo

# Install specific version
ccmd install github.com/user/repo@v1.0.0
ccmd install github.com/user/repo --version v1.0.0

# Install with custom name
ccmd install github.com/user/repo --name mycommand

# Force reinstall
ccmd install github.com/user/repo --force
```

### Supported Repository Formats

- `github.com/user/repo`
- `https://github.com/user/repo`
- `git@github.com:user/repo`
- `https://github.com/user/repo.git`
- `git@github.com:user/repo.git`
- `user/repo` (assumes GitHub)

## ccmd list

List all commands managed by ccmd with their versions, sources, and metadata.

### Usage

```bash
ccmd list [flags]
```

### Description

Shows only commands that are tracked in the ccmd-lock.yaml file and have entries in the .claude/commands/ directory.

### Options

- `-l, --long` - Show detailed output including metadata

### Examples

```bash
# List commands in table format
ccmd list

# Show detailed information
ccmd list --long
```

### Output Format

**Simple format** shows:
- NAME - Command name
- VERSION - Installed version
- DESCRIPTION - Brief description
- UPDATED - Last update time

**Long format** includes:
- All simple format fields
- Author information
- Tags
- License
- Homepage
- Installation timestamps
- Structure verification status

### Notes

- Commands with broken structure are marked with ⚠ 
- Use `--long` flag to see details about structure issues

## ccmd update

Update installed commands to their latest versions.

### Usage

```bash
ccmd update [command] [flags]
```

### Description

Updates a specific command or all commands to their latest versions from their source repositories.

### Options

- `-a, --all` - Update all installed commands
- `-c, --check` - Only check for updates without installing
- `-f, --force` - Force update even if version appears current

### Examples

```bash
# Update a specific command
ccmd update my-command

# Update all commands
ccmd update --all

# Check for updates without installing
ccmd update --check
ccmd update my-command --check

# Force update
ccmd update my-command --force
```

### Notes

- Without `--all` flag, you must specify a command name
- The `--check` flag shows available updates without making changes
- Updates preserve any local configuration in ccmd.yaml

## ccmd remove

Remove an installed command and clean up all associated files.

### Usage

```bash
ccmd remove <command-name> [flags]
```

### Description

Removes a command from the .claude/commands directory and optionally updates configuration files.

### Options

- `-f, --force` - Force removal without confirmation
- `-s, --save` - Update ccmd.yaml and ccmd-lock.yaml files

### Examples

```bash
# Remove with confirmation prompt
ccmd remove my-command

# Force removal without confirmation
ccmd remove my-command --force

# Remove and update config files
ccmd remove my-command --save
```

### Confirmation

Unless `--force` is used, the command will display:
- Command name and version
- Description (if available)
- Confirmation prompt

## ccmd search

Search for installed commands by keyword, tags, or author.

### Usage

```bash
ccmd search [keyword] [flags]
```

### Description

Searches through locally installed commands. This command searches metadata including names, descriptions, tags, and authors.

### Options

- `-t, --tags <tags>` - Filter by tags (comma-separated)
- `-a, --author <author>` - Filter by author
- `--all` - Show all commands (ignore keyword)

### Examples

```bash
# Search by keyword
ccmd search review

# Search by tags
ccmd search --tags code-review,quality

# Search by author
ccmd search --author "John Doe"

# List all installed commands
ccmd search --all

# Combine filters
ccmd search review --tags automation --author "Jane Doe"
```

### Output

Each result shows:
- Command name and version
- Description
- Author (if available)
- Tags (if available)
- Repository URL

## ccmd info

Display detailed information about an installed command.

### Usage

```bash
ccmd info <command-name> [flags]
```

### Description

Shows comprehensive information about a specific installed command, including metadata and structure verification.

### Options

- `--json` - Output in JSON format

### Examples

```bash
# Show command information
ccmd info my-command

# Output as JSON
ccmd info my-command --json
```

### Information Displayed

**Command Information:**
- Name, version, author
- Description
- Repository URL
- Homepage (if available)
- License (if specified)
- Tags
- Entry point file

**Installation Details:**
- Source repository
- Installation timestamp
- Last update timestamp

**Structure Verification:**
- Command directory status
- Standalone .md file status
- ccmd.yaml presence
- index.md presence
- Any structure issues

**Content Preview:**
- First 10 lines of the command's index.md file

## ccmd sync

Synchronize installed commands with ccmd.yaml configuration.

### Usage

```bash
ccmd sync [flags]
```

### Description

Analyzes the difference between ccmd.yaml and installed commands, then:
- Installs commands listed in ccmd.yaml but not installed
- Removes commands installed but not in ccmd.yaml
- Updates ccmd-lock.yaml to reflect current state

### Options

- `-n, --dry-run` - Show what would be done without making changes
- `-f, --force` - Force sync without confirmation

### Examples

```bash
# Analyze and sync commands
ccmd sync

# Preview changes without executing
ccmd sync --dry-run

# Force sync without confirmation
ccmd sync --force
```

### Sync Analysis Output

Shows:
- Commands to install (marked with +)
- Commands to remove (marked with -)
- Summary of operations to be performed

### Notes

- Useful after cloning a project with existing ccmd.yaml
- Helps maintain consistency between configuration and installed commands
- The `--dry-run` flag is recommended to preview changes first

## Common Workflows

### Setting Up a New Project

```bash
# 1. Initialize ccmd in your project
cd my-project
ccmd init

# 2. Install some commands
ccmd install github.com/user/command1
ccmd install github.com/user/command2

# 3. Commit the configuration
git add ccmd.yaml ccmd-lock.yaml
git commit -m "Add ccmd commands"
```

### Cloning an Existing Project

```bash
# 1. Clone the repository
git clone https://github.com/user/project
cd project

# 2. Install all commands
ccmd install
# or
ccmd sync
```

### Keeping Commands Updated

```bash
# Check for updates
ccmd update --all --check

# Update all commands
ccmd update --all

# Update specific command
ccmd update my-command
```

## See Also

- [Creating Commands](/creating-commands/) - Guide for creating your own commands
- [Examples](/examples/) - Real-world use cases and patterns
- [FAQ](/faq/) - Common questions and troubleshooting
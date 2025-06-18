# Remove Command Examples

This document shows various ways to use the `ccmd remove` command.

## Basic Remove

Remove a single installed command:

```bash
# Remove command
ccmd remove my-command

# Output:
# ? Are you sure you want to remove 'my-command'? (y/N) y
# ✓ Removed my-command successfully
```

## Force Remove (Skip Confirmation)

Remove without confirmation prompt:

```bash
# Force remove
ccmd remove my-command --force

# Or use -f shorthand
ccmd remove my-command -f

# Output:
# ✓ Removed my-command successfully
```

## Remove Multiple Commands

Remove several commands at once:

```bash
# Remove multiple commands
ccmd remove cmd1 cmd2 cmd3

# Output:
# ? Are you sure you want to remove 3 commands? (y/N) y
# ✓ Removed cmd1 successfully
# ✓ Removed cmd2 successfully
# ✓ Removed cmd3 successfully
# 
# Removed 3 commands
```

## Remove with Pattern

Remove commands matching a pattern:

```bash
# Remove all test commands
ccmd remove "test-*"

# Remove commands with specific prefix
ccmd remove "old-*" --force

# Output:
# Found 3 commands matching pattern "old-*":
# - old-tool-v1
# - old-helper
# - old-generator
# 
# ✓ Removed 3 commands
```

## Remove All Commands

Remove all installed commands (use with caution):

```bash
# Remove everything
ccmd remove --all

# Output:
# ⚠️  WARNING: This will remove all 12 installed commands
# ? Are you absolutely sure? Type 'yes' to confirm: yes
# 
# ✓ Removed all commands
# ✓ Cleaned up .claude directory
```

## Remove with Backup

Create a backup before removing:

```bash
# Remove but keep backup
ccmd remove my-command --backup

# Output:
# ✓ Creating backup of my-command...
# ✓ Backup saved to .claude/backups/my-command-backup-20240115
# ✓ Removed my-command successfully
# 
# To restore: ccmd restore my-command-backup-20240115
```

## Dry Run

See what would be removed without actually removing:

```bash
# Dry run
ccmd remove my-command --dry-run

# Output:
# Would remove:
# - Command: my-command v1.2.0
# - Location: .claude/commands/my-command/
# - Installed: 2024-01-10
# - Size: 2.3 MB
# 
# No changes made (dry run)
```

## Remove and Clean Cache

Remove command and clean up any cached data:

```bash
# Remove with cache cleanup
ccmd remove my-command --clean-cache

# Output:
# ✓ Removed my-command successfully
# ✓ Cleaned cache entries (freed 5.2 MB)
```

## Interactive Remove

Select commands to remove interactively:

```bash
# Interactive mode
ccmd remove --interactive

# Output:
# Select commands to remove:
# [ ] my-command (v1.2.0, 2.3 MB)
# [x] old-tool (v0.5.0, 1.1 MB)
# [x] unused-helper (v1.0.0, 500 KB)
# [ ] active-tool (v2.0.0, 3.2 MB)
# 
# Press space to select, enter to confirm
# 
# ✓ Removed 2 commands (freed 1.6 MB)
```

## Remove by Criteria

Remove commands based on various criteria:

```bash
# Remove commands not used in 30 days
ccmd remove --unused-days 30

# Remove commands from specific author
ccmd remove --author old-author --force

# Remove commands with specific tag
ccmd remove --tag deprecated

# Remove commands installed before date
ccmd remove --before 2023-01-01
```

## Remove Broken Commands

Remove commands that are corrupted or incomplete:

```bash
# Find and remove broken commands
ccmd remove --broken

# Output:
# Scanning for broken commands...
# Found 2 broken commands:
# - corrupt-cmd (missing index.md)
# - incomplete-tool (invalid ccmd.yaml)
# 
# ✓ Removed 2 broken commands
```

## Remove with Dependencies

Handle commands that have dependencies:

```bash
# Remove command and its dependencies
ccmd remove my-command --include-deps

# Output:
# my-command has 2 dependencies:
# - helper-lib (used only by my-command)
# - shared-tool (used by 3 other commands)
# 
# Will remove:
# - my-command
# - helper-lib (no other dependents)
# 
# Will keep:
# - shared-tool (still needed)
# 
# ? Proceed? (y/N) y
# ✓ Removed 2 commands
```

## Undo Remove

Restore recently removed commands:

```bash
# Show recently removed
ccmd restore --list

# Output:
# Recently removed commands:
# 1. my-command (removed 5 minutes ago)
# 2. old-tool (removed 1 hour ago)
# 3. test-cmd (removed 2 days ago)

# Restore specific command
ccmd restore my-command

# Restore by number
ccmd restore 1
```

## Remove Command Groups

Remove related commands together:

```bash
# Remove a command group
ccmd remove --group testing-tools

# Remove all commands from a repository
ccmd remove --source github.com/old-org/*
```

## Safe Remove

Extra safety checks before removing:

```bash
# Safe remove with verification
ccmd remove my-command --safe

# Output:
# Performing safety checks for 'my-command':
# ✓ No other commands depend on this
# ✓ No active processes using this command
# ✓ Backup available from 2 days ago
# 
# ? Proceed with removal? (y/N)
```

## Remove and Report

Get a detailed report of what was removed:

```bash
# Remove with detailed report
ccmd remove old-* --report

# Output saved to: .claude/removal-report-20240115.txt
# 
# Summary:
# - Removed 5 commands
# - Freed 15.3 MB disk space
# - Removed 127 files
# - Cleaned 5 lock entries
```

## Handle Remove Errors

Common errors when removing commands:

### Permission Denied

```bash
ccmd remove protected-command
# Error: Permission denied
# 
# Cannot remove 'protected-command': insufficient permissions
# 
# Try:
# - Run with sudo: sudo ccmd remove protected-command
# - Check file permissions: ls -la .claude/commands/protected-command
```

### Command Not Found

```bash
ccmd remove nonexistent
# Error: Command not found
# 
# No command named 'nonexistent' is installed
# 
# To see installed commands: ccmd list
```

### Command In Use

```bash
ccmd remove active-command
# Error: Command in use
# 
# Cannot remove 'active-command': currently being used by another process
# 
# Try:
# - Wait and retry
# - Force remove: ccmd remove active-command --force
```

## Bulk Operations

Efficient removal of many commands:

```bash
# Remove commands from list file
ccmd remove --from-file remove-list.txt

# Where remove-list.txt contains:
# old-command1
# deprecated-tool
# unused-helper

# Remove based on size
ccmd remove --larger-than 10MB

# Remove oldest commands
ccmd remove --oldest 5
```

## Post-Remove Cleanup

Clean up after removing commands:

```bash
# Full cleanup after removals
ccmd cleanup

# Output:
# ✓ Removed empty directories
# ✓ Updated lock file
# ✓ Cleaned orphaned cache entries
# ✓ Compacted command database
# 
# Freed 25.3 MB total disk space
```

## Remove Configuration

Configure remove behavior in `.claude/config.yaml`:

```yaml
remove:
  # Always create backups
  auto_backup: true
  
  # Skip confirmation for small commands
  skip_confirm_under: 1MB
  
  # Keep removal history
  keep_history: true
  history_days: 30
  
  # Safety settings
  prevent_remove_if_dependent: true
  warn_large_removals: true
  large_threshold: 10
```
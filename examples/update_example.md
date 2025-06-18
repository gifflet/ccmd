# Update Command Examples

This document demonstrates various use cases for the `ccmd update` command.

## Basic Update

Update a single command to its latest version:

```bash
# Update specific command
ccmd update my-command

# Output:
# ✓ Checking for updates...
# ✓ Found update: v1.0.0 → v1.2.0
# ✓ Downloading new version...
# ✓ Updated my-command to v1.2.0
```

## Update All Commands

Update all installed commands at once:

```bash
# Update everything
ccmd update --all

# Output:
# ✓ Checking 5 commands for updates...
# ✓ Updated my-command: v1.0.0 → v1.2.0
# ✓ Updated helper-tool: v2.1.0 → v2.3.1
# ✓ code-generator is already up to date (v3.0.0)
# ✓ test-runner is already up to date (v1.5.0)
# ✓ doc-builder is already up to date (v2.0.0)
# 
# Updated 2 of 5 commands
```

## Update to Specific Version

Update a command to a specific version or tag:

```bash
# Update to specific version
ccmd update my-command@v1.1.0

# Update to latest beta
ccmd update my-command@beta

# Update to specific commit
ccmd update my-command@abc123def
```

## Check for Updates Without Installing

See what updates are available without applying them:

```bash
# Check single command
ccmd update my-command --check

# Check all commands
ccmd update --all --check

# Output:
# Available updates:
# - my-command: v1.0.0 → v1.2.0 (2 versions behind)
#   Changes: bug fixes, new features
# - helper-tool: v2.1.0 → v2.3.1 (patch update available)
#   Changes: performance improvements
```

## Update with Backup

Create a backup before updating (useful for critical commands):

```bash
# Update with automatic backup
ccmd update my-command --backup

# Output:
# ✓ Creating backup of my-command v1.0.0...
# ✓ Backup saved to .claude/backups/my-command-1.0.0-20240115T120000
# ✓ Updating to v1.2.0...
# ✓ Update complete. Run 'ccmd restore my-command' to rollback if needed.
```

## Force Update

Force update even if already on latest version:

```bash
# Force reinstall latest version
ccmd update my-command --force

# Useful when:
# - Files might be corrupted
# - Want to ensure clean installation
# - Testing update process
```

## Update from Different Source

Update from a different repository or fork:

```bash
# Switch to a fork
ccmd update my-command --source github.com/newfork/my-command

# Update from specific branch
ccmd update my-command --source github.com/user/repo@develop
```

## Batch Updates with Filter

Update commands matching certain criteria:

```bash
# Update all commands from specific author
ccmd update --all --author yourname

# Update commands with specific tag
ccmd update --all --tag automation

# Update commands installed before date
ccmd update --all --before 2024-01-01
```

## Interactive Update

Choose which commands to update interactively:

```bash
# Interactive mode
ccmd update --interactive

# Output:
# Select commands to update:
# [ ] my-command (v1.0.0 → v1.2.0)
# [x] helper-tool (v2.1.0 → v2.3.1)
# [ ] code-generator (up to date)
# 
# Press space to select, enter to confirm
```

## Update with Changelog

View changelog before updating:

```bash
# Show changelog
ccmd update my-command --changelog

# Output:
# my-command changelog (v1.0.0 → v1.2.0):
# 
# v1.2.0 (2024-01-15)
# - Added new feature X
# - Fixed bug with Y
# - Improved performance by 50%
# 
# v1.1.0 (2024-01-01)
# - Added support for Z
# - Updated documentation
# 
# Proceed with update? (y/N)
```

## Rollback After Update

If something goes wrong, rollback to previous version:

```bash
# Rollback to previous version
ccmd rollback my-command

# Rollback to specific version
ccmd rollback my-command@v1.0.0

# List available versions to rollback to
ccmd rollback my-command --list
```

## Update Configuration

Configure update behavior in `.claude/config.yaml`:

```yaml
# Auto-check for updates
update:
  auto_check: true
  check_interval: 24h
  notify_only: false  # If true, only notify, don't auto-update
  
# Update preferences
preferences:
  backup_before_update: true
  show_changelog: true
  interactive_by_default: false
```

## Handling Update Errors

Common errors and solutions:

### Repository Not Found

```bash
ccmd update deleted-command
# Error: Repository no longer exists
# 
# The repository for 'deleted-command' could not be found.
# It may have been deleted or moved.
# 
# Options:
# - Remove the command: ccmd remove deleted-command
# - Update source: ccmd update deleted-command --source new/location
```

### Network Issues

```bash
ccmd update my-command
# Error: Network timeout
# 
# Failed to connect to github.com
# 
# Troubleshooting:
# - Check your internet connection
# - Try again with: ccmd update my-command --retry 3
# - Use a proxy: HTTPS_PROXY=http://proxy:8080 ccmd update my-command
```

### Version Conflicts

```bash
ccmd update my-command
# Error: Version conflict
# 
# Cannot update my-command from v2.0.0 to v1.5.0 (downgrade)
# 
# Options:
# - Force downgrade: ccmd update my-command@v1.5.0 --force
# - Stay on current: ccmd update --skip my-command
```

## Update Strategies

### Conservative Updates

Only update patch versions:

```bash
# Only patch updates (1.0.0 → 1.0.1)
ccmd update --all --patch-only

# Only minor updates (1.0.0 → 1.1.0)
ccmd update --all --minor-only
```

### Staged Updates

Update in stages for safety:

```bash
# 1. Update dev commands first
ccmd update --all --tag development

# 2. Test everything works
ccmd test --all

# 3. Update production commands
ccmd update --all --tag production
```

### Scheduled Updates

Set up automatic updates (requires cron or similar):

```bash
# Add to crontab
0 0 * * 0 cd /project && ccmd update --all --auto

# Or use ccmd daemon (future feature)
ccmd daemon start --update-interval 7d
```

## Performance Tips

### Parallel Updates

Update multiple commands in parallel:

```bash
# Enable parallel updates (faster)
ccmd update --all --parallel

# Limit concurrent updates
ccmd update --all --parallel --jobs 3
```

### Shallow Updates

For large repositories, use shallow clones:

```bash
# Shallow update (faster, less bandwidth)
ccmd update my-command --shallow

# Full update (slower, complete history)
ccmd update my-command --full
```

## Update Notifications

Get notified about available updates:

```bash
# Check and notify only
ccmd update --all --notify

# Output:
# 📢 Update notifications:
# 
# 2 updates available:
# - my-command: v1.0.0 → v1.2.0
# - helper-tool: v2.1.0 → v2.3.1
# 
# Run 'ccmd update --all' to install updates
```

## Advanced Update Scenarios

### Update with Dependencies

When commands have dependencies:

```bash
# Update including dependencies
ccmd update my-command --with-dependencies

# Skip dependency updates
ccmd update my-command --skip-dependencies
```

### Conditional Updates

Update based on conditions:

```bash
# Update only if tests pass
ccmd update my-command --if-tests-pass

# Update only if no breaking changes
ccmd update my-command --no-breaking
```

### Update Verification

Verify updates after installation:

```bash
# Update with verification
ccmd update my-command --verify

# Output:
# ✓ Updated to v1.2.0
# ✓ Verifying installation...
# ✓ Structure check passed
# ✓ Command loads successfully
# ✓ Update verified
```
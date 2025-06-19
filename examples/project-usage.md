# Project-Level Dependency Management

The enhanced `ccmd install` command now supports project-level dependency management through `ccmd.yaml` and `ccmd-lock.yaml` files.

## Usage Examples

### Installing Commands from ccmd.yaml

When you have a `ccmd.yaml` file in your project directory, you can install all listed commands at once:

```bash
# Install all commands defined in ccmd.yaml
ccmd install
```

Example `ccmd.yaml`:
```yaml
commands:
  - repo: user/awesome-cli
    version: v1.2.3
  - repo: org/helper-tool
    version: v2.0.0
  - repo: dev/test-cmd
    # version is optional, defaults to latest
```

### Installing Individual Commands

When you install a command with arguments, it will be automatically added to your project's `ccmd.yaml` and `ccmd-lock.yaml`:

```bash
# Install a specific command
ccmd install github.com/user/mycmd@v1.0.0

# This will:
# 1. Install the command
# 2. Add it to ccmd.yaml
# 3. Update ccmd-lock.yaml with exact version info
```

### Project Files

#### ccmd.yaml
Contains the desired state of your project's command dependencies:
```yaml
commands:
  - repo: user/cmd1
    version: v1.0.0
  - repo: org/cmd2
    version: v2.1.0
```

#### ccmd-lock.yaml
Contains the exact state of installed commands, including commit hashes and checksums:
```yaml
version: "1.0"
updated_at: 2024-01-15T10:30:00Z
commands:
  cmd1:
    name: cmd1
    repository: https://github.com/user/cmd1.git
    version: v1.0.0
    commit_hash: abc123...
    installed_at: 2024-01-15T10:30:00Z
    # ... other metadata
```

## Workflow

1. Create a `ccmd.yaml` file listing your project's command dependencies
2. Run `ccmd install` to install all commands
3. Commit both `ccmd.yaml` and `ccmd-lock.yaml` to version control
4. Team members can run `ccmd install` to get the same commands

## Benefits

- **Reproducible Builds**: Lock file ensures everyone gets the same versions
- **Easy Onboarding**: New team members just run `ccmd install`
- **Version Control**: Track command dependencies alongside your code
- **Automatic Updates**: When installing new commands, files are updated automatically
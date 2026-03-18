# Plugin Installation Examples

This document shows how to install Claude Code plugins using ccmd.

## Basic Plugin Installation

ccmd automatically detects whether a repository is a plugin or a command by reading the `type` field in the repository's `ccmd.yaml`. No special flags are needed.

```bash
# Install a plugin (auto-detected via type: plugin in the repo's ccmd.yaml)
ccmd install gifflet/review-plugin

# Install a specific version
ccmd install gifflet/review-plugin@1.0.0

# Install using full URL
ccmd install https://github.com/gifflet/review-plugin
```

## Install All Plugins and Commands from ccmd.yaml

```bash
# Installs all entries in the plugins: and commands: sections
ccmd install
```

## Force Plugin Installation

The `--plugin` flag is optional and only needed to force installation as a plugin when the repository does not have `type: plugin` in its `ccmd.yaml`.

```bash
# Force installation as plugin (use only when auto-detection is not available)
ccmd install user/some-repo --plugin
```

## Project ccmd.yaml with Plugins

When you install a plugin, ccmd updates your project's `ccmd.yaml` automatically:

```yaml
name: my-project
version: 1.0.0
commands:
  - gifflet/hello-world@1.0.0
plugins:
  - gifflet/review-plugin@1.0.0
```

## What Happens During Plugin Installation

1. Plugin repository is cloned to `.claude/plugins/{name}/`
2. Plugin is registered in `.claude/settings.json` under `enabledPlugins`
3. A marketplace entry is created/updated at `.claude/plugins/.claude-plugin/marketplace.json`
4. Claude Code discovers the plugin automatically on next startup

## After Installation

Plugins appear in `ccmd list` with `plugin` in the Type column:

```
NAME            VERSION  TYPE     DESCRIPTION
hello-world     1.0.0    command  Simple demo command
review-plugin   1.0.0    plugin   AI-powered code review
```

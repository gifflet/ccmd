---
title: "Plugin Support"
linkTitle: "Plugins"
weight: 35
type: docs
description: >
  Manage Claude Code plugins with ccmd. Install, create, and share plugins that extend Claude Code's capabilities.
keywords: ["ccmd plugins", "Claude Code plugins", "plugin manager", "Claude plugin", "AI plugins"]
---

ccmd supports not only slash commands but also Claude Code plugins — packages that extend Claude Code's own capabilities through its plugin system.

## Commands vs Plugins

| | Commands | Plugins |
|-|----------|---------|
| **Purpose** | Define reusable slash commands (`/name`) | Extend Claude Code itself |
| **Installation dir** | `.claude/commands/{name}` | `.claude/plugins/{name}` |
| **`ccmd.yaml` type** | (omitted, default) | `type: plugin` |
| **Entry field** | Required | Not required |
| **Registration** | Lock file only | Lock file + `settings.json` |
| **Marketplace** | No | Yes |

Use **commands** to define AI instructions invoked via `/slash-command`.

Use **plugins** to add tools, integrations, or context sources that Claude Code loads at startup.

## Installing a Plugin

ccmd automatically detects whether a repository is a plugin or a command by reading the `type` field in its `ccmd.yaml`. No special flags required.

```bash
# Install a plugin (auto-detected)
ccmd install gifflet/review-plugin

# Install a specific version
ccmd install gifflet/review-plugin@1.0.0
```

After installation, the plugin appears in `ccmd list` with type `plugin`:

```
NAME            VERSION  TYPE     DESCRIPTION
hello-world     1.0.0    command  Simple demo command
review-plugin   1.0.0    plugin   AI-powered code review
```

### What Happens During Installation

1. Plugin repository is cloned to `.claude/plugins/{name}/`
2. Plugin is registered in `.claude/settings.json`:
   ```json
   {
     "enabledPlugins": {
       "review-plugin@ccmd": true
     },
     "extraKnownMarketplaces": {
       "ccmd": { ... }
     }
   }
   ```
3. Marketplace is updated at `.claude/plugins/.claude-plugin/marketplace.json`
4. Claude Code discovers the plugin automatically on next startup

## Installing All Plugins from ccmd.yaml

When your project declares plugins in `ccmd.yaml`, run:

```bash
ccmd install
```

This installs all `commands` and `plugins` entries simultaneously.

### Project ccmd.yaml Example

```yaml
name: my-project
version: 1.0.0
commands:
  - gifflet/hello-world@1.0.0
plugins:
  - gifflet/review-plugin@1.0.0
```

## Removing a Plugin

```bash
ccmd remove review-plugin
```

This removes the plugin directory, unregisters it from `settings.json`, and updates the marketplace.

## Creating a Plugin

### 1. Initialize

```bash
mkdir my-plugin && cd my-plugin
ccmd init --plugin
# or shorthand
ccmd init -p
```

### 2. Resulting Structure

```
my-plugin/
├── ccmd.yaml                    # Plugin metadata with type: plugin
├── .claude-plugin/
│   └── plugin.json              # Claude Code plugin manifest
└── README.md
```

### 3. ccmd.yaml

```yaml
type: plugin
name: my-plugin
version: 1.0.0
description: My Claude Code plugin
author: Your Name
repository: https://github.com/username/my-plugin
tags:
  - productivity
license: MIT
```

### 4. Plugin Manifest (.claude-plugin/plugin.json)

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My Claude Code plugin",
  "author": {
    "name": "Your Name"
  },
  "repository": "https://github.com/username/my-plugin",
  "license": "MIT"
}
```

### 5. Publishing

```bash
git init && git add . && git commit -m "feat: initial plugin"
git remote add origin https://github.com/username/my-plugin
git push -u origin main
git tag v1.0.0 && git push --tags
```

Users install it with:

```bash
ccmd install username/my-plugin
```

## Example Plugin

[**review-plugin**](https://github.com/gifflet/review-plugin) — AI-powered code review plugin for Claude Code.

```bash
ccmd install gifflet/review-plugin
```

## See Also

- [Command Reference](/usage/) - All ccmd commands including plugin flags
- [Creating Commands](/creating-commands/) - Guide for creating slash commands
- [Examples](/examples/) - More usage examples

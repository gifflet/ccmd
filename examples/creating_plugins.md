# Creating Claude Code Plugins

This guide walks through creating a ccmd-compatible Claude Code plugin, using [review-plugin](https://github.com/gifflet/review-plugin) as a reference example.

## Plugins vs Commands

| | Commands | Plugins |
|-|----------|---------|
| Installation dir | `.claude/commands/{name}` | `.claude/plugins/{name}` |
| `ccmd.yaml` type | (omitted) | `type: plugin` |
| Entry field | Required | Not required |
| Integration | Slash commands (`/name`) | Claude Code plugin system |
| Marketplace | No | Yes |

Use a **command** when you want to define a reusable slash command (`/my-command`) that Claude follows.

Use a **plugin** when you want to extend Claude Code itself — adding tools, context sources, or integrations.

## Quick Start

```bash
mkdir my-plugin && cd my-plugin
ccmd init --plugin
```

This creates the following structure:

```
my-plugin/
├── ccmd.yaml                    # Plugin metadata with type: plugin
├── .claude-plugin/
│   └── plugin.json              # Claude Code plugin manifest
└── README.md                    # Plugin documentation
```

## ccmd.yaml for a Plugin

The key difference from a command is `type: plugin`. The `entry` field is not required.

```yaml
type: plugin
name: review-plugin
version: 1.0.0
description: AI-powered code review plugin for Claude Code
author: Your Name
repository: https://github.com/username/review-plugin
tags:
  - code-review
  - quality
license: MIT
```

## Plugin Manifest (.claude-plugin/plugin.json)

This file is read by Claude Code to register the plugin:

```json
{
  "name": "review-plugin",
  "version": "1.0.0",
  "description": "AI-powered code review plugin for Claude Code",
  "author": {
    "name": "Your Name"
  },
  "repository": "https://github.com/username/review-plugin",
  "license": "MIT"
}
```

## Example: review-plugin Structure

The [gifflet/review-plugin](https://github.com/gifflet/review-plugin) follows this structure:

```
review-plugin/
├── ccmd.yaml                    # type: plugin
├── .claude-plugin/
│   └── plugin.json              # Plugin manifest
└── README.md
```

To install and try it:

```bash
ccmd install gifflet/review-plugin
```

## Publishing Your Plugin

1. Push your repository to GitHub
2. Create a release tag: `git tag v1.0.0 && git push --tags`
3. Users install it with: `ccmd install username/my-plugin`

ccmd resolves the tag automatically and records it in `ccmd-lock.yaml`.

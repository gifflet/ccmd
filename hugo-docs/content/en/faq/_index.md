---
title: "FAQ"
linkTitle: "FAQ"
weight: 50
type: docs
description: >
  Frequently asked questions about ccmd and Claude Code slash commands.
keywords: ["FAQ", "troubleshooting", "Claude Code help", "slash commands FAQ", "common issues"]
---

## General Questions

### What is ccmd?

ccmd (Claude Command Manager) is a package manager for Claude Code slash commands. It allows you to install, update, and manage custom commands from Git repositories, similar to how npm manages JavaScript packages.

### Why should I use ccmd?

- **Reusability**: Use the same commands across multiple projects
- **Version Control**: Track specific versions of commands
- **Easy Sharing**: Share commands with your team via Git
- **Clean Codebase**: Keep AI configurations separate from project code
- **Simple Management**: Familiar package manager semantics

### How is ccmd different from just copying command files?

ccmd provides:
- Automated installation and updates
- Version management and locking
- Dependency tracking
- Command discovery and search
- Consistent project structure

## Installation & Setup

### What are the system requirements?

- **Node.js** v16+ or **Go** v1.23+
- **Git** installed and configured
- **Claude Code** installed
- macOS, Linux, or Windows

### How do I install ccmd?

Via NPM (recommended):
```bash
npm install -g @gifflet/ccmd
```

Via Go:
```bash
go install github.com/gifflet/ccmd/cmd/ccmd@latest
```

### Where does ccmd install commands?

Commands are installed in `.claude/commands/` within your project directory. This keeps commands project-specific and version-controlled.

## Using Commands

### How do I find available commands?

Currently, you can:
1. Search GitHub for repositories with `ccmd` topic
2. Check the [ccmd discussions](https://github.com/gifflet/ccmd/discussions)
3. Ask in the Claude Code community

A central registry is planned for the future.

### Can I use private repositories?

Yes! ccmd supports private repositories. Ensure you have Git configured with appropriate credentials (SSH keys or tokens).

### How do I use a specific version of a command?

```bash
# Install specific version
ccmd install github.com/user/command@v1.2.0

# Or in ccmd.yaml
commands:
  - github.com/user/command@v1.2.0
```

### What happens if a command updates?

Commands are locked to specific versions in `ccmd-lock.yaml`. Updates only happen when you explicitly run `ccmd update`.

## Creating Commands

### What makes a valid ccmd command?

At minimum, a command needs:
1. `ccmd.yaml` - Command metadata
2. `index.md` - Claude instructions

### What version numbering should I use?

Follow [semantic versioning](https://semver.org/):
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

## Troubleshooting

### "Command not found" after installation

1. Run `ccmd list` to verify installation
2. Check `.claude/commands/` directory
3. Restart Claude Code
4. Ensure command name matches what you're typing

### Installation fails with "repository not found"

- Verify the repository URL is correct
- Check you have access (for private repos)
- Ensure Git is properly configured
- Try the full URL format: `https://github.com/user/repo`

### "Permission denied" errors

- Check Git SSH keys are configured
- For HTTPS, ensure credentials are saved
- Verify repository permissions

### Commands not working in Claude Code

1. Verify with `ccmd info command-name`
2. Check the command's `index.md` exists
3. Review command syntax in documentation
4. Ensure you're using the correct format: `/command-name`

### How do I debug ccmd issues?

1. Use verbose mode: `ccmd install -v repo`
2. Check Git connectivity: `git ls-remote <repo>`
3. Verify file permissions
4. Check ccmd version: `ccmd --version`

## Best Practices

### Should I commit ccmd files?

**Yes, commit:**
- `ccmd.yaml` - Your project's command dependencies
- `ccmd-lock.yaml` - Ensures reproducible installs

**No, don't commit:**
- `.claude/commands/` - These are installed from sources

### How often should I update commands?

- **Development**: Update freely to get new features
- **Production**: Update cautiously, test thoroughly
- **Always**: Read changelogs before updating

### Can I modify installed commands?

Not recommended. Instead:
1. Fork the command repository
2. Make your changes
3. Install your fork
4. Optionally, submit a PR to the original

## Advanced Usage

### Can I use ccmd in CI/CD?

Yes! In your CI pipeline:
```bash
npm install -g @gifflet/ccmd
ccmd install
# Commands are now available
```

### How do I use ccmd with Docker?

Add to your Dockerfile:
```dockerfile
RUN npm install -g @gifflet/ccmd
COPY ccmd.yaml ccmd-lock.yaml ./
RUN ccmd install
```

## Contributing

### How can I contribute to ccmd?

- Report bugs via [GitHub Issues](https://github.com/gifflet/ccmd/issues)
- Submit PRs for features or fixes
- Improve documentation
- Create and share commands
- Help others in discussions

### Where can I get help?

- [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)
- [GitHub Issues](https://github.com/gifflet/ccmd/issues) for bugs
- Twitter: [@gifflet_](https://twitter.com/gifflet_)

### How can I stay updated?

- Watch the [GitHub repository](https://github.com/gifflet/ccmd)
- Follow [@gifflet_](https://twitter.com/gifflet_) on Twitter
- Join the discussions

Have a question not answered here? [Ask in discussions](https://github.com/gifflet/ccmd/discussions/new?category=q-a)!
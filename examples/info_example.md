# Info Command Examples

This document demonstrates how to use the `ccmd info` command to get detailed information about commands.

## Basic Info

Get information about an installed command:

```bash
# Show info for installed command
ccmd info auto-deploy

# Output:
# Command: auto-deploy
# Version: 2.1.0
# Author: devops-tools
# Description: Automated deployment pipeline for multiple platforms
# 
# Repository: https://github.com/devops-tools/auto-deploy
# License: MIT
# Installed: 2024-01-10 10:30 AM
# Size: 2.3 MB
# 
# Tags: deployment, automation, ci-cd, devops
# 
# Entry Point: index.md
# Documentation: https://auto-deploy.dev/docs
```

## Info from Repository

Get info about a command without installing it:

```bash
# Info from repository URL
ccmd info github.com/user/new-command

# Output:
# Command: new-command (not installed)
# Version: 1.0.0
# Author: user
# Description: A new command for testing
# 
# Repository: https://github.com/user/new-command
# License: Apache-2.0
# Size: ~1.5 MB (estimated)
# 
# Tags: testing, development
# 
# To install: ccmd install github.com/user/new-command
```

## Detailed Info

Show all available information:

```bash
# Detailed view
ccmd info auto-deploy --detailed

# Output:
# ═══════════════════════════════════════════════════════════════
# Command: auto-deploy
# ═══════════════════════════════════════════════════════════════
# 
# Basic Information:
#   Name: auto-deploy
#   Version: 2.1.0
#   Author: devops-tools
#   Email: team@devops-tools.com
#   License: MIT
# 
# Description:
#   Automated deployment pipeline for multiple platforms with
#   support for Docker, Kubernetes, and traditional servers.
# 
# Installation:
#   Date: 2024-01-10 10:30:00
#   Source: github.com/devops-tools/auto-deploy@v2.1.0
#   Method: git clone
#   Size: 2.3 MB (2,411,520 bytes)
#   Files: 15
# 
# Repository:
#   URL: https://github.com/devops-tools/auto-deploy
#   Stars: 1,234
#   Issues: 23 open, 156 closed
#   Last Updated: 2024-01-08
# 
# Configuration:
#   Entry: index.md
#   Config File: config.yaml (optional)
#   Dependencies: docker-cli@^20.0.0, kubectl@^1.25.0
# 
# Tags: deployment, automation, ci-cd, devops, docker, kubernetes
# 
# Documentation:
#   README: https://github.com/devops-tools/auto-deploy#readme
#   Docs: https://auto-deploy.dev/docs
#   Examples: https://auto-deploy.dev/examples
# 
# Changelog (latest):
#   v2.1.0 - Added Kubernetes support
#   v2.0.0 - Major refactor, breaking changes
#   v1.9.0 - Added rollback functionality
```

## Show Files

List files included in the command:

```bash
# Show file list
ccmd info auto-deploy --files

# Output:
# Files in auto-deploy:
# 
# Path                          Size      Modified
# ccmd.yaml                     521 B     2024-01-08
# index.md                      3.2 KB    2024-01-08
# README.md                     8.7 KB    2024-01-08
# LICENSE                       1.1 KB    2023-12-01
# config/
#   default.yaml               2.3 KB    2024-01-05
#   kubernetes.yaml            4.1 KB    2024-01-08
# templates/
#   docker-compose.yml         1.8 KB    2024-01-03
#   deployment.yaml            3.5 KB    2024-01-08
# scripts/
#   deploy.sh                  5.2 KB    2024-01-07
#   rollback.sh                2.9 KB    2024-01-07
# examples/
#   basic-deploy.md            2.1 KB    2024-01-02
#   advanced-deploy.md         4.3 KB    2024-01-05
#   troubleshooting.md         3.8 KB    2024-01-06
# 
# Total: 15 files (2.3 MB)
```

## Show Dependencies

Display command dependencies:

```bash
# Show dependencies
ccmd info auto-deploy --deps

# Output:
# Dependencies for auto-deploy:
# 
# Direct Dependencies:
#   - docker-cli (^20.0.0) - Required
#     Status: System dependency (not managed by ccmd)
#   
#   - kubectl (^1.25.0) - Required  
#     Status: System dependency (not managed by ccmd)
#   
#   - config-validator (^1.0.0) - Optional
#     Status: Not installed
#     Install: ccmd install github.com/devops-tools/config-validator
# 
# This command depends on: 3 packages
# This command is depended on by: 0 commands
```

## Version History

Show available versions:

```bash
# Show version history
ccmd info auto-deploy --versions

# Output:
# Version History for auto-deploy:
# 
# VERSION   RELEASE DATE   STATUS          NOTES
# v2.2.0    2024-01-15    Available       Added AWS ECS support
# v2.1.0    2024-01-08    Installed  ✓    Added Kubernetes support
# v2.0.0    2023-12-20    Available       Major refactor
# v1.9.0    2023-12-01    Available       Added rollback
# v1.8.2    2023-11-15    Available       Bug fixes
# v1.8.1    2023-11-10    Available       Performance improvements
# v1.8.0    2023-11-01    Available       Added monitoring
# 
# Current: v2.1.0 (1 version behind latest)
# To update: ccmd update auto-deploy
```

## JSON Output

Get info in JSON format:

```bash
# JSON output
ccmd info auto-deploy --json

# Pretty JSON
ccmd info auto-deploy --json --pretty
```

## Check for Updates

Check if updates are available:

```bash
# Check for updates
ccmd info auto-deploy --check-update

# Output:
# Command: auto-deploy
# Current Version: 2.1.0
# Latest Version: 2.2.0
# 
# Update Available:
#   Version: 2.2.0
#   Released: 2024-01-15 (3 days ago)
#   Changes:
#     - Added AWS ECS support
#     - Improved error handling
#     - Fixed rollback issue #142
# 
# To update: ccmd update auto-deploy
```

## Show README

Display the command's README:

```bash
# Show README
ccmd info auto-deploy --readme

# Output:
# # Auto Deploy
# 
# Automated deployment pipeline for multiple platforms.
# 
# ## Features
# - Docker deployment
# - Kubernetes deployment
# - Traditional server deployment
# - Rollback support
# - Health checks
# 
# ## Installation
# ```bash
# ccmd install github.com/devops-tools/auto-deploy
# ```
# 
# [... rest of README ...]
```

## Show Configuration

Display configuration options:

```bash
# Show config options
ccmd info auto-deploy --config

# Output:
# Configuration for auto-deploy:
# 
# Available Options:
#   deployment.target (string)
#     - Default: "docker"
#     - Options: docker, kubernetes, server
#     - Description: Deployment target platform
#   
#   deployment.namespace (string)
#     - Default: "default"
#     - Description: Kubernetes namespace (k8s only)
#   
#   deployment.timeout (integer)
#     - Default: 300
#     - Description: Deployment timeout in seconds
#   
#   rollback.enabled (boolean)
#     - Default: true
#     - Description: Enable automatic rollback on failure
# 
# Example config (.claude/commands/auto-deploy/config.yaml):
# ```yaml
# deployment:
#   target: kubernetes
#   namespace: production
#   timeout: 600
# rollback:
#   enabled: true
# ```
```

## Compare Versions

Compare two versions of a command:

```bash
# Compare versions
ccmd info auto-deploy --compare v2.0.0,v2.1.0

# Output:
# Comparing auto-deploy versions:
# 
# v2.0.0 → v2.1.0
# 
# Added:
#   + Kubernetes deployment support
#   + kubectl dependency
#   + templates/deployment.yaml
#   + New configuration options
# 
# Changed:
#   ~ Refactored deployment logic
#   ~ Updated documentation
#   ~ Improved error messages
# 
# Removed:
#   - Legacy deployment method
#   - Deprecated config options
# 
# Size: 2.1 MB → 2.3 MB (+200 KB)
```

## Security Information

Show security-related information:

```bash
# Security info
ccmd info auto-deploy --security

# Output:
# Security Information for auto-deploy:
# 
# License: MIT (Open Source)
# 
# Permissions Required:
#   - File system read/write
#   - Network access (for deployments)
#   - Process execution (docker, kubectl)
# 
# Security Checks:
#   ✓ No known vulnerabilities
#   ✓ Dependencies up to date
#   ✓ Code signed by author
#   ✓ Regular security updates
# 
# Last Security Update: 2024-01-05
# Security Contact: security@devops-tools.com
```

## Usage Examples

Show usage examples from the command:

```bash
# Show examples
ccmd info auto-deploy --examples

# Output:
# Examples for auto-deploy:
# 
# Basic Docker Deployment:
# ```
# # Deploy to Docker
# Use auto-deploy with target: docker
# ```
# 
# Kubernetes Deployment:
# ```
# # Deploy to Kubernetes cluster
# Use auto-deploy with target: kubernetes, namespace: prod
# ```
# 
# Rollback Deployment:
# ```
# # Rollback to previous version
# Use auto-deploy rollback
# ```
# 
# View all examples: ccmd run auto-deploy --help
```

## Check Integrity

Verify command integrity:

```bash
# Check integrity
ccmd info auto-deploy --verify

# Output:
# Verifying auto-deploy integrity:
# 
# ✓ Structure check passed
# ✓ Required files present
# ✓ Metadata valid
# ✓ No corrupted files
# ✓ Permissions correct
# 
# Integrity Status: VALID
```

## Show Related Commands

Find related or similar commands:

```bash
# Show related
ccmd info auto-deploy --related

# Output:
# Commands related to auto-deploy:
# 
# Similar Commands:
#   - k8s-deploy (v3.0.0) - Kubernetes-specific deployment tool
#   - docker-deploy (v1.5.0) - Docker-focused deployment
#   - ci-deploy (v2.2.0) - CI/CD integrated deployment
# 
# Complementary Commands:
#   - health-checker (v1.0.0) - Monitor deployed services
#   - log-aggregator (v2.1.0) - Collect deployment logs
#   - rollback-manager (v1.2.0) - Advanced rollback features
# 
# From Same Author (devops-tools):
#   - config-validator (v1.0.0) - Validate deployment configs
#   - secret-manager (v1.5.0) - Manage deployment secrets
```

## Export Info

Export command information:

```bash
# Export to file
ccmd info auto-deploy --export auto-deploy-info.txt

# Export as JSON
ccmd info auto-deploy --json --export auto-deploy.json

# Export for sharing
ccmd info auto-deploy --share

# Output:
# Shareable link created:
# https://ccmd.dev/commands/auto-deploy/v2.1.0
# 
# This link includes:
# - Command information
# - Installation instructions
# - Documentation
```
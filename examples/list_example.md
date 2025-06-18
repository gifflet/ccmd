# List Command Examples

This document shows various ways to use the `ccmd list` command to view installed commands.

## Basic List

Show all installed commands:

```bash
# Simple list
ccmd list

# Output:
# Installed commands:
# 
# auto-deploy      v2.1.0    Automated deployment pipeline
# test-runner      v1.5.0    Run tests with advanced features  
# doc-generator    v3.2.1    Generate documentation from code
# api-mocker       v1.0.3    Mock API endpoints for testing
# code-formatter   v2.0.0    Format code in multiple languages
# 
# Total: 5 commands
```

## Detailed List

Show more information about each command:

```bash
# Long format
ccmd list --long

# Or use -l shorthand
ccmd list -l

# Output:
# Installed commands:
# 
# NAME             VERSION   AUTHOR          INSTALLED        SIZE     DESCRIPTION
# auto-deploy      v2.1.0    devops-tools    2024-01-10      2.3 MB   Automated deployment pipeline
# test-runner      v1.5.0    testing-pro     2024-01-08      1.8 MB   Run tests with advanced features
# doc-generator    v3.2.1    doc-tools       2024-01-05      3.1 MB   Generate documentation from code
# api-mocker       v1.0.3    api-tools       2024-01-03      856 KB   Mock API endpoints for testing
# code-formatter   v2.0.0    format-master   2024-01-01      1.2 MB   Format code in multiple languages
# 
# Total: 5 commands (9.3 MB)
```

## JSON Output

Get output in JSON format for scripting:

```bash
# JSON format
ccmd list --json

# Pretty JSON
ccmd list --json --pretty

# Output:
# {
#   "commands": [
#     {
#       "name": "auto-deploy",
#       "version": "2.1.0",
#       "author": "devops-tools",
#       "description": "Automated deployment pipeline",
#       "installed_at": "2024-01-10T10:30:00Z",
#       "size": 2411520,
#       "repository": "github.com/devops-tools/auto-deploy"
#     },
#     ...
#   ],
#   "total": 5,
#   "total_size": 9748480
# }
```

## Filter by Pattern

List commands matching a pattern:

```bash
# List commands starting with "test"
ccmd list "test*"

# List commands containing "api"
ccmd list "*api*"

# Output:
# Commands matching "*api*":
# 
# api-mocker       v1.0.3    Mock API endpoints for testing
# api-generator    v2.0.0    Generate REST APIs from schemas
# 
# Total: 2 commands (matching pattern)
```

## Sort Options

Sort the list by different criteria:

```bash
# Sort by name (default)
ccmd list --sort name

# Sort by version
ccmd list --sort version

# Sort by install date (newest first)
ccmd list --sort date

# Sort by size (largest first)
ccmd list --sort size

# Sort by author
ccmd list --sort author

# Reverse sort order
ccmd list --sort size --reverse
```

## Filter by Author

Show commands from specific authors:

```bash
# Single author
ccmd list --author devops-tools

# Multiple authors
ccmd list --author "devops-tools,testing-pro"
```

## Filter by Tag

List commands with specific tags:

```bash
# Single tag
ccmd list --tag automation

# Multiple tags (commands with ALL tags)
ccmd list --tags "testing,ci"

# Multiple tags (commands with ANY tag)
ccmd list --tags "testing|ci"
```

## Show Updates Available

Check if updates are available:

```bash
# Show update status
ccmd list --check-updates

# Output:
# Installed commands:
# 
# NAME             INSTALLED   LATEST    STATUS
# auto-deploy      v2.1.0      v2.2.0    Update available
# test-runner      v1.5.0      v1.5.0    Up to date
# doc-generator    v3.2.1      v4.0.0    Major update available
# api-mocker       v1.0.3      v1.0.3    Up to date
# code-formatter   v2.0.0      v2.0.1    Patch available
# 
# Updates available: 3 commands
```

## Tree View

Show commands in a tree structure (useful for dependencies):

```bash
# Tree view
ccmd list --tree

# Output:
# Installed commands:
# 
# ├── auto-deploy (v2.1.0)
# │   ├── config-validator (v1.0.0) [dependency]
# │   └── yaml-parser (v2.3.0) [dependency]
# ├── test-runner (v1.5.0)
# │   └── assertion-lib (v3.0.0) [dependency]
# ├── doc-generator (v3.2.1)
# ├── api-mocker (v1.0.3)
# └── code-formatter (v2.0.0)
# 
# Total: 5 commands (3 dependencies)
```

## Show Only Names

List just the command names:

```bash
# Names only
ccmd list --names-only

# Output:
# auto-deploy
# test-runner
# doc-generator
# api-mocker
# code-formatter
```

## Export List

Export the list to a file:

```bash
# Export to text file
ccmd list --export commands.txt

# Export as JSON
ccmd list --json --export commands.json

# Export as CSV
ccmd list --csv --export commands.csv

# Export for backup
ccmd list --backup-format --export backup.ccmd
```

## Show Command Paths

Display where commands are installed:

```bash
# Show paths
ccmd list --show-paths

# Output:
# Installed commands:
# 
# auto-deploy      v2.1.0    ~/.claude/commands/auto-deploy/
# test-runner      v1.5.0    ~/.claude/commands/test-runner/
# doc-generator    v3.2.1    ~/.claude/commands/doc-generator/
```

## Group by Category

Group commands by their categories:

```bash
# Group by category
ccmd list --group-by category

# Output:
# Development Tools:
#   - code-formatter (v2.0.0)
#   - doc-generator (v3.2.1)
# 
# Testing:
#   - test-runner (v1.5.0)
#   - api-mocker (v1.0.3)
# 
# DevOps:
#   - auto-deploy (v2.1.0)
```

## Show Broken Commands

Identify commands with issues:

```bash
# Check for broken commands
ccmd list --health-check

# Output:
# Installed commands:
# 
# ✓ auto-deploy      v2.1.0    OK
# ✓ test-runner      v1.5.0    OK
# ⚠ doc-generator    v3.2.1    Warning: Missing optional file (examples/)
# ✗ api-mocker       v1.0.3    Error: Corrupted index.md
# ✓ code-formatter   v2.0.0    OK
# 
# Total: 5 commands (1 error, 1 warning)
```

## Filter by Date

List commands by installation date:

```bash
# Installed in last 7 days
ccmd list --installed-within 7d

# Installed before specific date
ccmd list --installed-before 2024-01-01

# Installed in date range
ccmd list --installed-between 2024-01-01,2024-01-31
```

## Show Statistics

Display usage statistics:

```bash
# Show stats
ccmd list --stats

# Output:
# Command Statistics:
# 
# Total commands: 5
# Total size: 9.3 MB
# Average size: 1.86 MB
# 
# By author:
#   - devops-tools: 1 command
#   - testing-pro: 1 command
#   - doc-tools: 1 command
#   - api-tools: 1 command
#   - format-master: 1 command
# 
# By category:
#   - Development: 2 commands
#   - Testing: 2 commands
#   - DevOps: 1 command
# 
# Installation timeline:
#   - Last 7 days: 2 commands
#   - Last 30 days: 5 commands
#   - Oldest: code-formatter (30 days ago)
#   - Newest: auto-deploy (5 days ago)
```

## Interactive Mode

Browse commands interactively:

```bash
# Interactive list
ccmd list --interactive

# Features:
# - Navigate with arrow keys
# - View command details with Enter
# - Update with 'u'
# - Remove with 'd'
# - Search with '/'
```

## Compare with Remote

Compare local commands with registry:

```bash
# Compare versions
ccmd list --compare-remote

# Output:
# Installed commands vs Registry:
# 
# NAME             LOCAL      REMOTE     STATUS
# auto-deploy      v2.1.0     v2.2.0     Behind by 1 version
# test-runner      v1.5.0     v1.5.0     In sync
# doc-generator    v3.2.1     v4.0.0     Major version behind
# old-tool         v0.9.0     -          Not in registry
# private-cmd      v1.0.0     N/A        Private repository
```

## List Configuration

Show commands with their configuration:

```bash
# Show with config
ccmd list --with-config

# Output:
# auto-deploy (v2.1.0)
#   Config: custom deployment targets defined
#   
# test-runner (v1.5.0)
#   Config: default settings
#   
# doc-generator (v3.2.1)
#   Config: custom templates enabled
```

## Performance Information

Show performance metrics:

```bash
# Show performance info
ccmd list --performance

# Output:
# Command Performance Metrics:
# 
# NAME             LOAD TIME   LAST USED   USAGE COUNT
# auto-deploy      120ms       2 hrs ago   45 times
# test-runner      95ms        1 day ago   128 times
# doc-generator    230ms       3 days ago  67 times
# api-mocker       80ms        1 week ago  23 times
# code-formatter   110ms       Today       203 times
```

## Custom Columns

Select specific columns to display:

```bash
# Custom columns
ccmd list --columns name,version,size

# All available columns
ccmd list --columns name,version,author,size,date,repo,description,tags
```

## List Aliases

If commands have aliases:

```bash
# Show with aliases
ccmd list --show-aliases

# Output:
# Installed commands:
# 
# auto-deploy (v2.1.0)
#   Aliases: deploy, ad
#   
# test-runner (v1.5.0)
#   Aliases: test, tr
```
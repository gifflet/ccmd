# Search Command Examples

This document demonstrates how to use the `ccmd search` command to find commands in the registry.

## Basic Search

Search for commands by keyword:

```bash
# Simple search
ccmd search automation

# Output:
# Found 12 commands matching "automation":
# 
# 1. auto-deploy (v2.1.0) by devops-tools
#    Automated deployment pipeline for multiple platforms
#    ⭐ 4.8 (245 reviews) | 10.2k installs
# 
# 2. test-automation (v3.5.2) by testing-pro
#    Comprehensive test automation framework
#    ⭐ 4.6 (189 reviews) | 8.7k installs
# 
# 3. ci-automation (v1.8.0) by ci-experts
#    CI/CD automation tools and templates
#    ⭐ 4.5 (156 reviews) | 6.3k installs
```

## Search with Multiple Keywords

Search using multiple terms:

```bash
# Multiple keywords (AND search)
ccmd search docker kubernetes

# Multiple keywords (OR search)  
ccmd search "docker OR kubernetes"

# Phrase search
ccmd search "continuous integration"
```

## Filter by Tags

Search commands with specific tags:

```bash
# Single tag
ccmd search --tag testing

# Multiple tags
ccmd search --tags testing,automation,ci

# Exclude tags
ccmd search automation --exclude-tags deprecated,legacy
```

## Filter by Author

Find commands from specific authors:

```bash
# Commands by author
ccmd search --author gifflet

# Multiple authors
ccmd search --authors gifflet,devtools,automation-pro

# Search within author's commands
ccmd search testing --author gifflet
```

## Sort Results

Control how results are sorted:

```bash
# Sort by popularity (default)
ccmd search testing --sort popularity

# Sort by rating
ccmd search testing --sort rating

# Sort by recent updates
ccmd search testing --sort updated

# Sort by name
ccmd search testing --sort name

# Reverse sort order
ccmd search testing --sort rating --reverse
```

## Limit Results

Control number of results:

```bash
# Show only top 5 results
ccmd search automation --limit 5

# Show all results (no limit)
ccmd search automation --all

# Paginate results
ccmd search automation --page 2 --per-page 10
```

## Detailed Output

Get more information about each result:

```bash
# Detailed view
ccmd search automation --detailed

# Output includes:
# - Full description
# - Installation command
# - Dependencies
# - Recent updates
# - Author information
# - License
```

## Output Formats

Different output formats for various uses:

```bash
# JSON output (for scripts)
ccmd search automation --json

# CSV output
ccmd search automation --csv

# Minimal output (names only)
ccmd search automation --names-only

# Full details
ccmd search automation --full
```

## Advanced Filters

Complex filtering options:

```bash
# By version requirements
ccmd search --min-version 2.0.0

# By license
ccmd search --license MIT

# By size
ccmd search --max-size 5MB

# By rating
ccmd search --min-rating 4.0

# By install count
ccmd search --min-installs 1000

# By last update
ccmd search --updated-within 30d
```

## Search in Descriptions

Search in different fields:

```bash
# Search in descriptions only
ccmd search "error handling" --in description

# Search in names only  
ccmd search test --in name

# Search everywhere (default)
ccmd search docker --in all
```

## Boolean Search

Use boolean operators for complex searches:

```bash
# AND operator
ccmd search "docker AND deployment"

# OR operator
ccmd search "jest OR mocha OR vitest"

# NOT operator
ccmd search "testing NOT unit"

# Combined
ccmd search "(docker OR kubernetes) AND deployment"
```

## Search Similar Commands

Find commands similar to one you know:

```bash
# Find similar to installed command
ccmd search --similar-to my-deploy-tool

# Find alternatives
ccmd search --alternative-to old-tool
```

## Search with Categories

Browse commands by category:

```bash
# List all categories
ccmd search --list-categories

# Search within category
ccmd search --category development

# Multiple categories
ccmd search --categories "development,automation"
```

## Interactive Search

Use interactive mode for exploration:

```bash
# Interactive search
ccmd search --interactive

# Features:
# - Real-time filtering
# - Preview command details
# - Direct installation
# - Save searches
```

## Search History

Work with previous searches:

```bash
# Show search history
ccmd search --history

# Repeat last search
ccmd search --last

# Repeat specific search
ccmd search --repeat 3
```

## Save Search Results

Save results for later use:

```bash
# Save to file
ccmd search automation --save results.txt

# Save as JSON
ccmd search automation --json --save results.json

# Save as installable list
ccmd search automation --save-list install-list.txt
```

## Search Examples by Use Case

### Find Testing Tools

```bash
ccmd search --tags testing,unit-test,integration-test \
            --sort rating \
            --min-rating 4.0 \
            --limit 10
```

### Find Docker-related Commands

```bash
ccmd search docker \
            --tags container,deployment \
            --exclude-tags deprecated \
            --sort popularity
```

### Find Recent Commands

```bash
ccmd search --updated-within 7d \
            --sort updated \
            --limit 20
```

### Find Lightweight Commands

```bash
ccmd search --max-size 1MB \
            --tags utility,tool \
            --sort size
```

## Search Tips

### Use Quotes for Exact Phrases

```bash
# Without quotes - searches for "error" OR "handling"
ccmd search error handling

# With quotes - searches for exact phrase
ccmd search "error handling"
```

### Combine Filters for Precision

```bash
# Very specific search
ccmd search "api testing" \
            --author trusted-tools \
            --tag automation \
            --min-rating 4.5 \
            --license MIT
```

### Use Wildcards

```bash
# Wildcard search
ccmd search "test*"  # matches test, testing, tester, etc.

# Multiple wildcards
ccmd search "*deploy*"  # matches auto-deploy, deployment, etc.
```

## Search Configuration

Configure search behavior in `.claude/config.yaml`:

```yaml
search:
  # Default registry
  registry: https://registry.ccmd.dev
  
  # Default filters
  defaults:
    min_rating: 3.0
    exclude_tags: [deprecated, legacy]
    sort: popularity
    limit: 20
  
  # Search behavior
  behavior:
    case_sensitive: false
    fuzzy_matching: true
    include_beta: false
    
  # Caching
  cache:
    enabled: true
    ttl: 1h
```

## Offline Search

Search in locally cached data:

```bash
# Search offline (cached data only)
ccmd search automation --offline

# Update cache then search
ccmd search --update-cache automation

# Force online search
ccmd search automation --online
```

## Search Operators Reference

| Operator | Description | Example |
|----------|-------------|---------|
| AND | Both terms must match | `docker AND deploy` |
| OR | Either term matches | `jest OR mocha` |
| NOT | Exclude term | `testing NOT e2e` |
| "" | Exact phrase | `"error handling"` |
| * | Wildcard | `test*` |
| () | Grouping | `(docker OR k8s) AND deploy` |

## Troubleshooting Search

### No Results Found

```bash
ccmd search "very specific term"
# No commands found matching "very specific term"
# 
# Try:
# - Using fewer or different keywords
# - Checking spelling
# - Using wildcards: ccmd search "very*"
# - Browsing categories: ccmd search --list-categories
```

### Registry Unavailable

```bash
ccmd search testing
# Error: Could not connect to registry
# 
# Falling back to cached results (last updated: 2 hours ago)
# 
# To force offline mode: ccmd search testing --offline
# To retry: ccmd search testing --retry 3
```
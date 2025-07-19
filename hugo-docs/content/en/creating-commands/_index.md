---
title: "Creating Commands"
linkTitle: "Creating Commands"
weight: 30
type: docs
description: >
  Step-by-step guide to creating, publishing, and sharing your own Claude Code slash commands.
keywords: ["create Claude commands", "slash command development", "AI command creation", "Claude Code development", "command tutorial"]
---

This guide walks you through creating, publishing, and sharing a ccmd command from scratch. We'll create a real example command called "code-reviewer" that helps with code review tasks.

## Prerequisites

- ccmd installed on your system
- Git installed
- GitHub account
- Basic familiarity with Markdown

## Overview

We'll create a command that:
- Reviews code for common issues
- Suggests improvements
- Can be installed by anyone using ccmd

## Step 1: Initialize Your Command Project

First, create a directory for your command and initialize it:

```bash
# Create and enter the command directory
mkdir code-reviewer
cd code-reviewer

# Initialize the ccmd project
ccmd init
```

When prompted, enter these values:

```
This utility will walk you through creating a ccmd.yaml file.
Press ^C at any time to quit.

name: (code-reviewer) code-reviewer
version: (1.0.0) 1.0.0
description: AI-powered code review assistant for Claude Code
author: John Doe
repository: https://github.com/johndoe/code-reviewer
entry: (index.md) index.md
tags (comma-separated): code-review, quality, automation

About to write to /path/to/code-reviewer/ccmd.yaml:

name: code-reviewer
version: 1.0.0
description: AI-powered code review assistant for Claude Code
author: John Doe
repository: https://github.com/johndoe/code-reviewer
entry: index.md
tags:
  - code-review
  - quality
  - automation

Is this OK? (yes) yes
```

This creates:
- `ccmd.yaml` - Your command's metadata
- `.claude/commands/` directory structure (if needed)

## Step 2: Create the Command Instructions

Create the `index.md` file with instructions for Claude:

```bash
touch index.md
```

Add the following content to `index.md`:

```markdown
   # Code Reviewer Command

   You are an AI assistant specialized in code review. When this command is invoked, you help users review their code for quality, best practices, and potential issues.

   ## Instructions

   When the user invokes this command, you should:

   1. **Analyze the code structure**
      - Identify the programming language
      - Understand the project context
      - Note architectural patterns

   2. **Review for common issues**
      - Code style and formatting
      - Potential bugs or errors
      - Security vulnerabilities
      - Performance concerns

   3. **Suggest improvements**
      - Better algorithms or data structures
      - Cleaner code patterns
      - Missing error handling
      - Documentation gaps

   4. **Provide actionable feedback**
      - Be specific about line numbers
      - Explain why something is an issue
      - Offer concrete solutions

   ## Parameters

   - `--file <path>`: Review a specific file
   - `--severity <level>`: Filter by severity (error, warning, info)
   - `--focus <area>`: Focus on specific areas (security, performance, style)

   ## Examples

   ### Basic Usage
   User: "/code-reviewer --file src/api.js"

   You should:
   1. Read src/api.js
   2. Analyze the code
   3. Provide a structured review with:
      - Summary of findings
      - Detailed issues with line numbers
      - Suggested fixes

   ### Focused Review
   User: "/code-reviewer --focus security"

   You should:
   1. Scan all relevant files
   2. Focus specifically on security issues
   3. Highlight:
      - SQL injection risks
      - XSS vulnerabilities
      - Authentication issues
      - Data validation problems

   ## Response Format

   Structure your responses as:

   ### Code Review Summary
   - Files reviewed: X
   - Issues found: Y (Z critical, W warnings, V suggestions)

   ### Critical Issues
   1. **[File:Line]** Description
      - Why it's a problem
      - How to fix it
      ```language
      // Fixed code example
      ```

   ### Warnings
   [Similar format]

   ### Suggestions
   [Similar format]

   ### Overall Assessment
   Brief summary of code quality and next steps.
```

## Step 3: Create User Documentation

Create a README.md for users:

```bash
touch README.md
```

Add the following content:

```markdown
      # Code Reviewer Command for ccmd

      AI-powered code review assistant that helps you catch bugs, improve code quality, and follow best practices.

      ## Installation

      ```bash
      ccmd install github.com/johndoe/code-reviewer
      ```

      ## Usage

      Basic code review:
      ```bash
      /code-reviewer --file src/main.js
      ```

      Security-focused review:
      ```bash
      /code-reviewer --focus security
      ```

      Review all files:
      ```bash
      /code-reviewer
      ```

      ## Features

      - 🐛 Bug detection
      - 🔒 Security vulnerability scanning
      - ⚡ Performance optimization suggestions
      - 📝 Documentation improvements
      - 🎨 Code style recommendations

      ## Options

      - `--file <path>`: Review specific file
      - `--severity <level>`: Filter by severity (error|warning|info)
      - `--focus <area>`: Focus area (security|performance|style|all)

      ## Examples

      ### Review a Python file
      ```
      /code-reviewer --file app.py
      ```

      ### Security audit
      ```
      /code-reviewer --focus security --severity error
      ```

      ## Requirements

      - Works with any programming language
      - Best results with common languages (JavaScript, Python, Go, etc.)

      ## Author

      John Doe - [@johndoe](https://github.com/johndoe)

      ## License

      MIT
```

## Step 4: Version Control with Git

Initialize Git and create your first commit:

```bash
# Initialize Git repository
git init

# Add all files
git add .

# Create initial commit
git commit -m "feat: initial code-reviewer command implementation"

# Create version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
```

## Step 5: Create GitHub Repository

1. Go to [GitHub](https://github.com/new)
2. Create a new repository named `code-reviewer`
3. Don't initialize with README (we already have one)
4. Create the repository

Then connect your local repository:

```bash
# Add GitHub remote (replace with your username)
git remote add origin https://github.com/johndoe/code-reviewer.git

# Push code and tags
git push -u origin main
git push origin --tags
```

## Step 6: Install and Test Your Command

Now anyone can install your command:

```bash
# Install from GitHub
ccmd instal https://github.com/johndoe/code-reviewer

# Verify installation
ccmd list
```

Output should show:
```
NAME                  VERSION     DESCRIPTION                               UPDATED
--------------------  ----------  ----------------------------------------  --------------------
code-reviewer         1.0.0       AI-powered code review assistant for      1 minute ago
                                  Claude Code
```

Test in Claude Code:
```
/code-reviewer --file <path-to-file>
```

## Step 7: Updating Your Command

When you need to update your command:

### 1. Make Changes

Edit your files as needed:
```bash
# Edit index.md to add new features
# Update README.md with new documentation
```

### 2. Update Version

Edit `ccmd.yaml`:
```yaml
version: 1.1.0  # Bump version according to semver
```

### 3. Commit and Tag

```bash
# Commit changes
git add .
git commit -m "feat: add support for TypeScript type checking"

# Create new version tag
git tag -a v1.1.0 -m "Release version 1.1.0

Features:
- TypeScript type checking
- Improved security scanning
- Better error messages"

# Push updates
git push origin main
git push origin --tags
```

### 4. Users Update

Users can update to the latest version:
```bash
ccmd update code-reviewer
```

## Version Guidelines

Follow semantic versioning (semver):

- **MAJOR** (1.0.0 → 2.0.0): Breaking changes
- **MINOR** (1.0.0 → 1.1.0): New features, backwards compatible
- **PATCH** (1.0.0 → 1.0.1): Bug fixes

Examples:
- Adding a new parameter: Minor version bump
- Changing command behavior: Major version bump
- Fixing a typo: Patch version bump

## Best Practices

### 1. Clear Instructions
Write instructions that are specific and actionable. Claude needs to understand exactly what to do.

### 2. Useful Examples
Include real-world examples in your index.md that demonstrate common use cases.

### 3. Error Handling
Tell Claude how to handle common error scenarios:
```markdown
If the file doesn't exist, respond with:
"Error: File 'filename' not found. Please check the file path."
```

### 4. Consistent Updates
When updating:
- Document changes in commit messages
- Update README.md with new features
- Use meaningful version tags

## Troubleshooting

### Command not found after installation
- Check `ccmd list` to verify installation
- Ensure ccmd.yaml has correct name field
- Verify you have access to the repository

### Changes not reflecting after update
- Check you pushed tags: `git push origin --tags`
- Users may need to run: `ccmd update code-reviewer`
- Verify version was bumped in ccmd.yaml

### Installation fails
- Ensure repository exists and is accessible
- Check ccmd.yaml is valid YAML
- Verify index.md exists in repository root

## Complete Example Repository Structure

Your final repository structure should look like:

```
code-reviewer/
├── ccmd.yaml          # Command metadata
├── index.md           # Claude instructions
└── README.md          # User documentation
```

## Next Steps

1. **Enhance your command**: Add more features based on user feedback
2. **Create more commands**: Build a suite of useful tools
3. **Share with community**: Announce in ccmd discussions
4. **Collaborate**: Accept PRs and issues from users

## Resources

- [ccmd Documentation](https://github.com/gifflet/ccmd)
- [API Reference](/technical/api/) - File formats and specifications
- [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)

---

Congratulations! You've created and published your first ccmd command. Users around the world can now install and use your code-reviewer command to improve their code quality.
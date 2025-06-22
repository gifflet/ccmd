# Command Structure Guide

This guide explains how to create commands for ccmd, covering the required structure, best practices, and advanced features.

## Table of Contents

- [Overview](#overview)
- [Project vs Command Configuration](#project-vs-command-configuration)
- [Required Files](#required-files)
- [File Structure](#file-structure)
- [Command ccmd.yaml Reference](#command-ccmdyaml-reference)
- [Project ccmd.yaml Reference](#project-ccmdyaml-reference)
- [ccmd-lock.yaml Reference](#ccmd-lockyaml-reference)
- [Writing Command Instructions](#writing-command-instructions)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Testing Your Command](#testing-your-command)

## Overview

A ccmd command is a Git repository containing instructions for Claude Code. Commands help automate tasks, provide specialized knowledge, or enhance Claude's capabilities in specific domains.

## Project vs Command Configuration

It's important to understand that ccmd uses two different types of `ccmd.yaml` files:

1. **Project ccmd.yaml** - Located in your project root, lists commands to install
2. **Command ccmd.yaml** - Located in each command's repository, defines command metadata

These files have completely different structures and purposes.

## Required Files

Every command MUST have these two files:

1. **ccmd.yaml** - Metadata about your command
2. **index.md** - Instructions for Claude (can be named differently if specified in ccmd.yaml)

Additional recommended files:
- **README.md** - Documentation for users
- **LICENSE** - License for your command
- **examples/** - Usage examples

## File Structure

### Basic Command Structure

```
my-command/
├── ccmd.yaml          # Required: Command metadata
├── index.md           # Required: Claude instructions
└── README.md          # Recommended: User documentation
```

### Installed Command Structure

When installed, commands are stored in your project:

```
your-project/
├── ccmd.yaml          # Project configuration
├── ccmd-lock.yaml     # Lock file (auto-generated)
└── .claude/
    └── commands/
        ├── my-command.md      # Standalone file (copy of index.md)
        └── my-command/        # Full command directory
            ├── ccmd.yaml
            ├── index.md
            └── README.md
```

## Command ccmd.yaml Reference

The `ccmd.yaml` file in a command repository defines the command's metadata:

```yaml
# Required fields
name: my-awesome-command          # Command name (lowercase, hyphens allowed)
version: 1.0.0                    # Semantic version (major.minor.patch)
entry: index.md                   # Entry file (default: index.md)

# Optional fields
description: Short description            # One-line description
author: Your Name                         # Command author
repository: https://github.com/user/repo  # Source repository
tags:                                     # Tags for discovery
  - automation
  - testing
  - development
```

All fields except `tags` are required for a valid command.

## Project ccmd.yaml Reference

The `ccmd.yaml` file in your project root lists commands to install:

```yaml
commands:
  - owner/repo             # Install latest version
  - owner/repo@1.0.0       # Install specific version
  - owner/repo@branch      # Install from branch
```

This is a simple list format - no other fields are used in the project's ccmd.yaml.

## ccmd-lock.yaml Reference

The `ccmd-lock.yaml` file tracks installed command versions:

```yaml
version: "1.0"
lockfileVersion: 1
commands:
  command-name:
    name: command-name
    version: 1.0.0
    source: https://github.com/owner/repo.git
    resolved: https://github.com/owner/repo.git@1.0.0
    commit: abc123def456...
    installed_at: 2025-06-22T01:07:51.524358-03:00
    updated_at: 2025-06-22T01:07:51.524358-03:00
```

This file is automatically managed by ccmd and should not be edited manually.

## Writing Command Instructions

The `index.md` file contains instructions for Claude. Write clear, specific instructions that help Claude understand what to do.

### Basic Structure

```markdown
# Command Name

Brief description of what this command does.

## Purpose

Explain the command's purpose and when to use it.

## Instructions

1. Step-by-step instructions
2. Be specific and clear
3. Include error handling

## Parameters

- `--option`: Description of option
- `--flag`: What this flag does

## Examples

### Example 1: Basic Usage
When the user says "..." you should...

### Example 2: Advanced Usage
For complex scenarios...

## Notes

- Important considerations
- Limitations
- Best practices
```

### Advanced Instructions

```markdown
# Advanced Command

You are an AI assistant specialized in [domain]. When this command is invoked, follow these guidelines:

## Core Responsibilities

1. **Analysis Phase**
   - Examine the project structure
   - Identify relevant files
   - Understand the context

2. **Planning Phase**
   - Create a plan of action
   - Consider edge cases
   - Validate assumptions

3. **Execution Phase**
   - Implement the solution
   - Provide clear feedback
   - Handle errors gracefully

## Context Understanding

You have access to:
- File system operations
- Code analysis capabilities
- Pattern matching

## Decision Framework

When deciding how to proceed:

```flowchart
Start -> Analyze Request -> Is it valid?
  |                           |
  Yes                         No -> Request clarification
  |
  V
Plan approach -> Execute -> Verify -> Complete
```

## Error Handling

Common errors and how to handle them:

1. **Missing Dependencies**
   - Check for required tools
   - Suggest installation commands
   - Provide alternatives

2. **Invalid Input**
   - Validate parameters
   - Show usage examples
   - Explain what went wrong

## Code Generation Guidelines

When generating code:
- Follow language best practices
- Include error handling
- Add meaningful comments
- Consider performance

## Response Format

Structure your responses as:

1. **Acknowledgment** - Confirm understanding
2. **Plan** - Outline the approach
3. **Implementation** - Execute the plan
4. **Summary** - Recap what was done
5. **Next Steps** - Suggest follow-up actions
```

## Best Practices

### 1. Clear Instructions

❌ **Bad:**
```markdown
Help with testing stuff.
```

✅ **Good:**
```markdown
You are an AI assistant specialized in creating comprehensive test suites. When invoked:

1. Analyze the codebase to identify testable components
2. Generate unit tests with >80% coverage
3. Include edge cases and error scenarios
4. Use the project's existing test framework
```

### 2. Specific Examples

❌ **Bad:**
```markdown
Handle various cases appropriately.
```

✅ **Good:**
```markdown
## Examples

### Creating a REST API Test
When user says: "test my API endpoint"
1. Identify the endpoint (e.g., POST /api/users)
2. Generate test cases:
   - Valid input: `{ "name": "John", "email": "john@example.com" }`
   - Missing fields: `{ "name": "John" }`
   - Invalid email: `{ "name": "John", "email": "invalid" }`
3. Create the test file with proper assertions
```

### 3. Parameter Documentation

❌ **Bad:**
```markdown
Supports various options.
```

✅ **Good:**
```markdown
## Parameters

- `--language <lang>`: Target programming language (js, python, go, rust)
  - Default: Detected from project
  - Example: `--language python`
  
- `--style <style>`: Code style to follow
  - Options: standard, google, airbnb
  - Default: standard
  - Example: `--style google`

- `--output <path>`: Where to save generated files
  - Default: Current directory
  - Example: `--output ./tests/`
```

### 4. Context Awareness

Include information about what Claude should look for:

```markdown
## Project Analysis

Before proceeding, analyze:

1. **Technology Stack**
   - Check package.json, requirements.txt, go.mod
   - Identify frameworks (React, Django, etc.)
   - Note build tools and configurations

2. **Project Structure**
   - Locate source directories
   - Find test directories
   - Identify configuration files

3. **Coding Standards**
   - Detect linting configurations
   - Observe existing code style
   - Check for formatting rules
```

### 5. Error Messages

Provide helpful error messages:

```markdown
## Error Handling

If the project type cannot be determined:
"I couldn't automatically detect your project type. Please specify:
- For Node.js: Ensure package.json exists
- For Python: Ensure requirements.txt or setup.py exists
- For Go: Ensure go.mod exists
Or use the --language flag to specify manually."
```

## Examples

### Simple Command Example

**Command's ccmd.yaml:**
```yaml
name: format-json
version: 1.0.0
description: Format and validate JSON files
author: Jane Doe
repository: https://github.com/janedoe/format-json
entry: index.md
tags:
  - json
  - formatting
  - validation
```

```markdown
# index.md
# Format JSON Command

This command formats and validates JSON files in your project.

## Instructions

When invoked, you should:

1. Find all JSON files in the current directory and subdirectories
2. Validate each file for proper JSON syntax
3. Format with 2-space indentation
4. Report any errors found
5. Optionally fix the errors if requested

## Parameters

- `--fix`: Automatically fix formatting issues
- `--indent <n>`: Use n spaces for indentation (default: 2)
- `--sort-keys`: Sort object keys alphabetically

## Example Usage

User: "format all json files"
Response: 
- Search for .json files
- Validate and format each file
- Report: "Formatted 5 JSON files. Found and fixed 2 syntax errors."
```

### Project Configuration Example

**Project's ccmd.yaml:**
```yaml
commands:
  - janedoe/format-json
  - apitools/api-generator@2.1.0
  - myorg/internal-tool@main
```

### Complete Command Example

**Command's ccmd.yaml:**
```yaml
name: api-generator
version: 2.1.0
description: Generate REST API boilerplate with tests and documentation
author: API Tools Team
repository: https://github.com/apitools/api-generator
entry: index.md
tags:
  - api
  - rest
  - boilerplate
  - testing
```

```markdown
# index.md
# API Generator Command

You are an AI assistant specialized in generating REST API boilerplate code with comprehensive tests and documentation.

## Core Capabilities

1. Generate REST API endpoints with CRUD operations
2. Create corresponding test suites
3. Generate OpenAPI/Swagger documentation
4. Set up authentication and validation

## Workflow

### 1. Analysis Phase
- Detect project type and framework
- Identify existing patterns
- Check for configuration files

### 2. Generation Phase

Based on user input, generate:

#### API Endpoints
```javascript
// Example for "generate user API"
router.get('/users', async (req, res) => {
  const users = await User.findAll();
  res.json(users);
});

router.post('/users', validateUser, async (req, res) => {
  const user = await User.create(req.body);
  res.status(201).json(user);
});
```

#### Tests
```javascript
describe('User API', () => {
  it('should return all users', async () => {
    const response = await request(app).get('/users');
    expect(response.status).toBe(200);
    expect(response.body).toBeInstanceOf(Array);
  });
});
```

#### Documentation
```yaml
paths:
  /users:
    get:
      summary: Get all users
      responses:
        200:
          description: List of users
```

## Parameters

- `--framework <name>`: Target framework (express, fastify, koa)
- `--database <type>`: Database type (postgres, mongodb, mysql)
- `--auth <method>`: Authentication method (jwt, oauth, basic)
- `--no-tests`: Skip test generation
- `--no-docs`: Skip documentation generation

## Examples

### Basic Usage
User: "generate a product API"
Actions:
1. Create routes/products.js with CRUD endpoints
2. Create tests/products.test.js with test suite
3. Update OpenAPI spec with product endpoints
4. Create models/Product.js with schema

### Advanced Usage
User: "generate API for blog with posts and comments using postgres"
Actions:
1. Set up Sequelize with PostgreSQL
2. Create models: Post, Comment with associations
3. Generate nested routes: /posts/:id/comments
4. Include pagination and filtering
5. Add comprehensive tests
6. Generate full OpenAPI documentation
```

## Testing Your Command

Before publishing, test your command thoroughly:

### 1. Local Testing

```bash
# Clone your command repository
git clone https://github.com/you/your-command
cd your-command

# Validate structure
ls ccmd.yaml index.md  # Should exist
```

### 2. Validation Checklist

- [ ] ccmd.yaml is valid YAML
- [ ] All required fields are present (name, version, description, author, repository, entry)
- [ ] Version follows semantic versioning (e.g., 1.0.0)
- [ ] index.md exists and is readable
- [ ] Instructions are clear and specific
- [ ] Examples work as documented
- [ ] Error cases are handled

### 3. Integration Testing

Test your command in different scenarios:

```bash
# Test in empty directory
mkdir test-empty && cd test-empty
/your-command

# Test in existing project
cd ~/my-project
/your-command --resource users

# Test with parameters
/your-command --resource products --no-tests
```

## Publishing Your Command

### 1. Prepare Repository

```bash
# Ensure all files are committed
git add .
git commit -m "feat: initial command implementation"

# Create a version tag
git tag -a v1.0.0 -m "Initial release"
git push origin main --tags
```

### 2. Share Your Command

Once published, others can install your command:

```bash
ccmd install github.com/username/my-command
```

### 3. Documentation

Create a clear README.md with:
- Installation instructions
- Usage examples
- Parameter documentation
- Common use cases

## How Commands Are Installed

When you run `ccmd install`, the following happens:

1. **Repository is cloned** to a temporary directory
2. **Validation** ensures ccmd.yaml and index.md exist
3. **Files are copied** to `.claude/commands/[command-name]/`
4. **Standalone file** is created at `.claude/commands/[command-name].md`
5. **Lock file** is updated with version information

The standalone `.md` file is what Claude Code uses when you invoke the command.

## Troubleshooting

### Common Issues

1. **Command not found after installation**
   - Ensure ccmd.yaml has correct name field
   - Check file permissions

2. **Instructions not working as expected**
   - Test with different phrasings
   - Add more specific examples
   - Include edge cases

3. **Version conflicts**
   - Use proper semantic versioning
   - Document breaking changes
   - Consider backwards compatibility

## Support

- Read the [ccmd documentation](https://github.com/gifflet/ccmd)
- Ask in [GitHub Discussions](https://github.com/gifflet/ccmd/discussions)
- Report issues on [GitHub](https://github.com/gifflet/ccmd/issues)
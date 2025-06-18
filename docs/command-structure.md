# Command Structure Guide

This guide explains how to create commands for ccmd, covering the required structure, best practices, and advanced features.

## Table of Contents

- [Overview](#overview)
- [Required Files](#required-files)
- [File Structure](#file-structure)
- [ccmd.yaml Reference](#ccmdyaml-reference)
- [Writing Command Instructions](#writing-command-instructions)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Testing Your Command](#testing-your-command)
- [Publishing](#publishing)

## Overview

A ccmd command is a Git repository containing instructions for Claude Code. Commands help automate tasks, provide specialized knowledge, or enhance Claude's capabilities in specific domains.

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
├── README.md          # Recommended: User documentation
└── LICENSE            # Recommended: License file
```

### Advanced Command Structure

```
my-advanced-command/
├── ccmd.yaml          # Metadata
├── index.md           # Main instructions
├── README.md          # User docs
├── LICENSE            # License
├── examples/          # Usage examples
│   ├── basic.md
│   ├── advanced.md
│   └── troubleshooting.md
├── prompts/           # Additional prompts
│   ├── setup.md
│   └── cleanup.md
└── templates/         # File templates
    ├── config.yaml
    └── Dockerfile
```

## ccmd.yaml Reference

The `ccmd.yaml` file defines your command's metadata:

```yaml
# Required fields
name: my-awesome-command          # Command name (lowercase, hyphens allowed)
version: 1.0.0                    # Semantic version (major.minor.patch)
description: Short description    # One-line description (max 100 chars)

# Recommended fields
author: Your Name                 # Command author
email: your.email@example.com    # Contact email
repository: https://github.com/user/repo  # Source repository
license: MIT                      # License type

# Optional fields
entry: index.md                   # Entry file (default: index.md)
homepage: https://example.com     # Project homepage
documentation: https://docs.example.com  # Documentation URL
issues: https://github.com/user/repo/issues  # Issue tracker

# Command categorization
tags:                            # Tags for discovery
  - automation
  - testing
  - development
  
categories:                      # Main categories
  - development
  - productivity

# Advanced features
dependencies:                    # Other commands this depends on
  - other-command@^1.0.0
  - helper-command@~2.1.0

requirements:                    # System requirements
  - nodejs>=18.0.0
  - python>=3.8

config:                         # Configuration options
  default_timeout: 30
  allow_network: true
  
# Experimental features  
includes:                       # Additional files to include
  - prompts/*.md
  - templates/*
```

### Version Constraints

Dependencies use semantic versioning constraints:
- `^1.0.0` - Compatible with 1.x.x (>=1.0.0 <2.0.0)
- `~1.2.0` - Approximately 1.2.x (>=1.2.0 <1.3.0)
- `1.2.3` - Exact version
- `>=1.0.0` - Minimum version
- `*` or `latest` - Any version

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

### Simple Command

```yaml
# ccmd.yaml
name: format-json
version: 1.0.0
description: Format and validate JSON files
author: Jane Doe
repository: https://github.com/janedoe/format-json
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

### Advanced Command

```yaml
# ccmd.yaml
name: api-generator
version: 2.1.0
description: Generate REST API boilerplate with tests and documentation
author: API Tools Team
email: team@apitools.dev
repository: https://github.com/apitools/api-generator
homepage: https://apitools.dev
license: MIT

tags:
  - api
  - rest
  - boilerplate
  - testing

categories:
  - development
  - automation

dependencies:
  - openapi-validator@^1.0.0

requirements:
  - nodejs>=16.0.0

config:
  default_framework: express
  include_tests: true
  include_docs: true
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

# Test with ccmd
ccmd install file:///path/to/your-command
ccmd run your-command
```

### 2. Validation Checklist

- [ ] ccmd.yaml is valid YAML
- [ ] All required fields are present
- [ ] Version follows semver
- [ ] index.md exists and is readable
- [ ] Instructions are clear and specific
- [ ] Examples work as documented
- [ ] Error cases are handled

### 3. Integration Testing

Test your command in different scenarios:

```bash
# Test in empty directory
mkdir test-empty && cd test-empty
ccmd run your-command

# Test in existing project
cd ~/my-project
ccmd run your-command

# Test with parameters
ccmd run your-command --option value
```

## Publishing

### 1. Prepare Repository

```bash
# Ensure all files are committed
git add .
git commit -m "feat: initial command implementation"

# Create a version tag
git tag -a v1.0.0 -m "Initial release"
git push origin main --tags
```

### 2. Documentation

Create a comprehensive README.md:

```markdown
# My Command

## Installation

\```bash
ccmd install github.com/username/my-command
\```

## Usage

\```bash
ccmd run my-command [options]
\```

## Options

- `--option`: Description

## Examples

[Include practical examples]

## License

MIT
```

### 3. Submit to Registry (Coming Soon)

Once the official registry is available:

```bash
ccmd publish
```

## Advanced Features

### Multiple Entry Points

```yaml
# ccmd.yaml
entry: main.md

commands:
  setup:
    entry: commands/setup.md
    description: Initial setup
  
  cleanup:
    entry: commands/cleanup.md
    description: Clean up resources
```

### Conditional Logic

```markdown
## Platform-Specific Instructions

Check the operating system and adjust:

- **macOS**: Use `brew install`
- **Linux**: Use `apt-get` or `yum`
- **Windows**: Use `choco` or download manually
```

### Templates and Snippets

Include reusable templates:

```
templates/
├── docker/
│   ├── Dockerfile.node
│   └── Dockerfile.python
└── ci/
    ├── github-actions.yml
    └── gitlab-ci.yml
```

Reference in instructions:

```markdown
Use the appropriate Dockerfile template:
- For Node.js: Use templates/docker/Dockerfile.node
- For Python: Use templates/docker/Dockerfile.python
```

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
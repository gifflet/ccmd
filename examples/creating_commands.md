# Creating Commands - Complete Example

This guide walks through creating a complete ccmd command from scratch.

## Example: Creating a "Code Review Assistant" Command

Let's create a command that helps with code reviews in Claude Code.

### Step 1: Set Up Repository

```bash
# Create repository
mkdir code-review-assistant
cd code-review-assistant

# Initialize git
git init
```

### Step 2: Create ccmd.yaml

```yaml
# ccmd.yaml
name: code-review-assistant
version: 1.0.0
description: AI-powered code review assistant for multiple languages
author: Jane Developer
email: jane@example.com
repository: https://github.com/jane/code-review-assistant
homepage: https://code-review-assistant.dev
license: MIT

# Entry point (optional, defaults to index.md)
entry: index.md

# Categorization
tags:
  - code-review
  - quality
  - automation
  - development
  
categories:
  - development
  - productivity

# Optional metadata
keywords:
  - review
  - code quality
  - pull request
  - lint
  - best practices

# Requirements (optional)
requirements:
  - git>=2.0.0

# Configuration schema (optional)
config:
  languages:
    type: array
    default: ["javascript", "python", "go"]
    description: Languages to review
  
  strictness:
    type: string
    default: "medium"
    enum: ["low", "medium", "high"]
    description: Review strictness level
```

### Step 3: Create index.md

```markdown
# Code Review Assistant

You are an AI code review assistant. When invoked, you help developers review code changes, identify issues, and suggest improvements.

## Core Capabilities

1. **Code Analysis**
   - Identify bugs and potential issues
   - Check for security vulnerabilities
   - Suggest performance improvements
   - Ensure code style consistency

2. **Best Practices**
   - Verify naming conventions
   - Check for proper error handling
   - Ensure adequate commenting
   - Validate test coverage

3. **Architecture Review**
   - Assess design patterns
   - Check for code duplication
   - Evaluate modularity
   - Review dependencies

## Instructions

When the user invokes this command, follow these steps:

### 1. Analyze Context

First, understand what needs to be reviewed:

- Check for uncommitted changes: `git status`
- Look for staged changes: `git diff --cached`
- Review recent commits: `git log --oneline -5`
- Identify changed files: `git diff --name-only`

### 2. Perform Review

Based on the context, perform a comprehensive review:

#### For Each Changed File:

1. **Syntax and Style**
   - Check for syntax errors
   - Verify consistent formatting
   - Ensure proper indentation
   - Look for unused variables/imports

2. **Logic and Bugs**
   - Identify potential null/undefined issues
   - Check for off-by-one errors
   - Verify error handling
   - Look for race conditions

3. **Security**
   - Check for hardcoded credentials
   - Identify injection vulnerabilities
   - Review authentication/authorization
   - Check for sensitive data exposure

4. **Performance**
   - Identify N+1 queries
   - Check for unnecessary loops
   - Look for memory leaks
   - Suggest caching opportunities

5. **Maintainability**
   - Assess code complexity
   - Check for proper abstractions
   - Verify meaningful names
   - Ensure adequate documentation

### 3. Provide Feedback

Structure your review feedback as follows:

```
## Code Review Summary

**Files Reviewed**: X files
**Issues Found**: Y issues (Z critical, W warnings, V suggestions)

### Critical Issues 🔴

1. **[File:Line]** Description of critical issue
   ```language
   // Current code
   ```
   
   **Suggestion**:
   ```language
   // Improved code
   ```
   
   **Reason**: Explanation of why this is critical

### Warnings 🟡

1. **[File:Line]** Description of warning...

### Suggestions 🟢

1. **[File:Line]** Description of suggestion...

### Positive Findings ✅

- Good practice noticed in...
- Well-structured code in...

## Overall Assessment

[Provide overall feedback on code quality, architecture, and suggestions for improvement]
```

## Parameters

The command accepts these parameters:

- `--files <pattern>`: Review only files matching pattern
- `--commit <hash>`: Review specific commit
- `--branch <name>`: Review changes in branch
- `--language <lang>`: Focus on specific language
- `--strictness <level>`: Set review strictness (low/medium/high)
- `--focus <area>`: Focus on specific area (security/performance/style)

## Examples

### Example 1: Review Uncommitted Changes

User: "review my changes"

Action:
1. Run `git status` to see changed files
2. Run `git diff` to see changes
3. Analyze each change
4. Provide structured feedback

### Example 2: Review Specific Commit

User: "review commit abc123"

Action:
1. Run `git show abc123`
2. Analyze the commit changes
3. Provide detailed review

### Example 3: Security-Focused Review

User: "review for security issues"

Action:
1. Focus on security aspects
2. Check for vulnerabilities
3. Suggest security improvements

## Language-Specific Checks

### JavaScript/TypeScript
- Check for `console.log` statements
- Verify Promise handling
- Look for missing `await` keywords
- Check for proper TypeScript types

### Python
- Verify PEP 8 compliance
- Check for proper exception handling
- Look for mutable default arguments
- Verify type hints (Python 3.5+)

### Go
- Check for proper error handling
- Verify goroutine safety
- Look for defer usage
- Check for proper interface usage

### Java
- Verify null checks
- Check for resource leaks
- Look for proper exception handling
- Verify thread safety

## Integration with Git Workflow

When reviewing pull requests:

1. Fetch the PR branch
2. Compare with base branch
3. Review only changed files
4. Format feedback for PR comments

## Configuration

Users can configure behavior via `.claude/commands/code-review-assistant/config.yaml`:

```yaml
languages: ["javascript", "python", "go", "java"]
strictness: high
focus_areas:
  - security
  - performance
  - maintainability
ignore_patterns:
  - "*.test.js"
  - "*.spec.ts"
  - "__tests__/*"
```

## Best Practices for Reviews

1. **Be Constructive**: Always suggest improvements
2. **Prioritize Issues**: Critical > Warnings > Suggestions
3. **Provide Examples**: Show better alternatives
4. **Acknowledge Good Code**: Highlight positive aspects
5. **Be Specific**: Include file names and line numbers
6. **Explain Why**: Provide reasoning for suggestions

## Error Handling

Handle these common scenarios:

- No changes to review: "No changes found to review"
- Binary files: "Skipping binary file: [filename]"
- Large files: "File too large, showing summary"
- Merge conflicts: "Merge conflicts detected, resolve first"
```

### Step 4: Create README.md

```markdown
# Code Review Assistant

An AI-powered code review assistant for Claude Code that helps maintain code quality and catch issues early.

## Features

- 🔍 **Comprehensive Analysis**: Reviews code for bugs, security issues, and performance problems
- 🌐 **Multi-Language Support**: Works with JavaScript, Python, Go, Java, and more
- 🎯 **Customizable Focus**: Configure to focus on security, performance, or style
- 📊 **Detailed Reports**: Provides structured feedback with severity levels
- 🔄 **Git Integration**: Works seamlessly with your Git workflow

## Installation

```bash
ccmd install github.com/jane/code-review-assistant
```

## Usage

### Basic Review

Review all uncommitted changes:

```bash
ccmd run code-review-assistant
```

### Review Specific Files

```bash
ccmd run code-review-assistant --files "src/*.js"
```

### Review a Commit

```bash
ccmd run code-review-assistant --commit abc123
```

### Security-Focused Review

```bash
ccmd run code-review-assistant --focus security
```

## Configuration

Create `.claude/commands/code-review-assistant/config.yaml`:

```yaml
# Languages to review
languages:
  - javascript
  - python
  - go

# Review strictness (low/medium/high)
strictness: medium

# Areas to focus on
focus_areas:
  - security
  - performance
  - maintainability

# Files to ignore
ignore_patterns:
  - "*.test.js"
  - "*.spec.ts"
  - "vendor/*"
```

## Examples

### Example 1: JavaScript Code Review

```javascript
// Before review
function processUser(user) {
  console.log(user);
  fetch('/api/user/' + user.id)
    .then(res => res.json())
    .then(data => {
      document.getElementById('user').innerHTML = data.name;
    });
}

// After review suggestions
async function processUser(user) {
  if (!user?.id) {
    throw new Error('Invalid user object');
  }
  
  try {
    const response = await fetch(`/api/user/${encodeURIComponent(user.id)}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    const userElement = document.getElementById('user');
    if (userElement) {
      userElement.textContent = data.name; // Prevent XSS
    }
  } catch (error) {
    console.error('Failed to process user:', error);
    // Handle error appropriately
  }
}
```

### Example 2: Python Code Review

```python
# Before review
def calculate_average(numbers):
  total = 0
  for n in numbers:
    total += n
  return total / len(numbers)

# After review suggestions
def calculate_average(numbers: list[float]) -> float:
    """
    Calculate the arithmetic mean of a list of numbers.
    
    Args:
        numbers: List of numeric values
        
    Returns:
        The arithmetic mean
        
    Raises:
        ValueError: If the list is empty
        TypeError: If list contains non-numeric values
    """
    if not numbers:
        raise ValueError("Cannot calculate average of empty list")
    
    try:
        return sum(numbers) / len(numbers)
    except TypeError as e:
        raise TypeError(f"List contains non-numeric values: {e}")
```

## Review Output Example

```
## Code Review Summary

**Files Reviewed**: 3 files
**Issues Found**: 7 issues (2 critical, 3 warnings, 2 suggestions)

### Critical Issues 🔴

1. **[auth.js:45]** SQL Injection vulnerability
   ```javascript
   const query = `SELECT * FROM users WHERE id = ${userId}`;
   ```
   
   **Suggestion**:
   ```javascript
   const query = 'SELECT * FROM users WHERE id = ?';
   db.query(query, [userId]);
   ```
   
   **Reason**: Direct string interpolation in SQL queries can lead to SQL injection attacks.

2. **[api.py:23]** Missing authentication check
   ```python
   @app.route('/api/admin/users')
   def get_users():
       return User.query.all()
   ```
   
   **Suggestion**:
   ```python
   @app.route('/api/admin/users')
   @require_auth
   @require_role('admin')
   def get_users():
       return User.query.all()
   ```

### Warnings 🟡

1. **[utils.js:12]** Potential null reference
2. **[data.py:34]** Mutable default argument
3. **[server.go:56]** Unchecked error

### Suggestions 🟢

1. **[index.js:5]** Consider using const instead of let
2. **[helpers.py:15]** Add type hints for better code clarity

### Positive Findings ✅

- Excellent error handling in payment.js
- Well-structured module organization
- Comprehensive test coverage in utils module
```

## Supported Languages

- JavaScript/TypeScript
- Python
- Go
- Java
- Ruby
- PHP
- C#
- Rust
- Swift
- Kotlin

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- Issues: [GitHub Issues](https://github.com/jane/code-review-assistant/issues)
- Discussions: [GitHub Discussions](https://github.com/jane/code-review-assistant/discussions)
- Email: support@code-review-assistant.dev
```

### Step 5: Add Examples Directory

Create `examples/security-review.md`:

```markdown
# Security-Focused Code Review Example

This example shows how to use the code review assistant with a security focus.

## Command

```bash
ccmd run code-review-assistant --focus security
```

## Sample Output

When reviewing code with potential security issues:

```
## Security-Focused Code Review

**Security Issues Found**: 5 (3 high, 2 medium)

### High Severity 🔴

1. **[login.js:23]** Hardcoded credentials detected
   ```javascript
   const API_KEY = "sk-1234567890abcdef";
   ```
   
   **Fix**: Store sensitive data in environment variables
   ```javascript
   const API_KEY = process.env.API_KEY;
   ```

2. **[user.py:45]** SQL Injection vulnerability
   ```python
   query = f"SELECT * FROM users WHERE email = '{email}'"
   ```
   
   **Fix**: Use parameterized queries
   ```python
   query = "SELECT * FROM users WHERE email = %s"
   cursor.execute(query, (email,))
   ```

3. **[api.go:67]** Missing input validation
   ```go
   userId := r.URL.Query().Get("id")
   user := GetUser(userId) // No validation
   ```
   
   **Fix**: Validate and sanitize input
   ```go
   userId := r.URL.Query().Get("id")
   if !isValidUUID(userId) {
       http.Error(w, "Invalid user ID", http.StatusBadRequest)
       return
   }
   user := GetUser(userId)
   ```

### Medium Severity 🟡

1. **[config.js:12]** Insecure random number generation
2. **[auth.py:34]** Weak password requirements

## Security Best Practices Applied

✅ Checked for injection vulnerabilities
✅ Validated authentication/authorization
✅ Reviewed encryption usage
✅ Checked for hardcoded secrets
✅ Validated input sanitization
✅ Reviewed error handling for info leakage
```
```

### Step 6: Create LICENSE

```
MIT License

Copyright (c) 2024 Jane Developer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Step 7: Test Your Command Locally

```bash
# Test structure
ls -la
# Should show: ccmd.yaml, index.md, README.md, LICENSE, examples/

# Test with ccmd (local install)
ccmd install file:///path/to/code-review-assistant

# Run the command
ccmd run code-review-assistant
```

### Step 8: Publish to GitHub

```bash
# Add all files
git add .

# Commit
git commit -m "Initial implementation of code-review-assistant"

# Add remote
git remote add origin https://github.com/jane/code-review-assistant.git

# Push
git push -u origin main

# Create version tag
git tag -a v1.0.0 -m "Initial release"
git push origin v1.0.0
```

### Step 9: Share Your Command

Now others can install your command:

```bash
ccmd install github.com/jane/code-review-assistant
```

## Best Practices Demonstrated

1. **Clear Metadata**: Comprehensive ccmd.yaml with all relevant fields
2. **Detailed Instructions**: Step-by-step guide for Claude in index.md
3. **User Documentation**: Complete README with examples
4. **Examples**: Practical examples showing different use cases
5. **Error Handling**: Clear instructions for handling edge cases
6. **Configuration**: Flexible configuration options
7. **Multi-Language**: Support for multiple programming languages
8. **Structured Output**: Consistent, readable output format

## Tips for Success

1. **Test Thoroughly**: Test your command in various scenarios
2. **Handle Edge Cases**: Consider what could go wrong
3. **Provide Examples**: Show real-world usage
4. **Keep It Focused**: Do one thing well
5. **Version Properly**: Use semantic versioning
6. **Document Well**: Clear documentation helps adoption
7. **Gather Feedback**: Iterate based on user feedback
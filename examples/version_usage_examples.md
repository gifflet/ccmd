# Version Specification Examples

## Basic Usage

### 1. Using @ notation (inline version)
```bash
# Install and run specific version
ccmd terraform@1.5.0 init

# The @ notation is part of the command name
ccmd prettier@3.0.0 --write .
```

### 2. Using --version flag
```bash
# Version specified as a flag
ccmd terraform --version 1.5.0 init

# Flag can appear anywhere after the command
ccmd prettier --write . --version 3.0.0
```

### 3. When both are specified
```bash
# --version flag takes precedence
ccmd terraform@1.4.0 --version 1.5.0 init
# Result: Uses terraform version 1.5.0 (from --version flag)

ccmd prettier@2.8.0 --version 3.0.0 --write .
# Result: Uses prettier version 3.0.0 (from --version flag)
```

## Practical Scenarios

### Scenario 1: Script with default version
```bash
#!/bin/bash
# deploy.sh - Always use stable terraform version

# Hardcoded version in script
ccmd terraform@1.4.6 apply -auto-approve
```

### Scenario 2: Override script version for testing
```bash
# Override the hardcoded version without modifying the script
./deploy.sh --version 1.5.0-rc1
```

### Scenario 3: CI/CD Pipeline
```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    steps:
      - name: Deploy with terraform
        run: |
          # Default version in workflow
          ccmd terraform@1.4.6 apply
        
      - name: Test with newer version
        run: |
          # Override for testing
          ccmd terraform@1.4.6 apply --version ${{ matrix.terraform_version }}
```

### Scenario 4: Makefile with configurable version
```makefile
# Default version
TERRAFORM_VERSION ?= 1.4.6

deploy:
	ccmd terraform@$(TERRAFORM_VERSION) apply

# Can be overridden with:
# make deploy TERRAFORM_VERSION=1.5.0
# OR using --version flag:
# ccmd terraform@1.4.6 apply --version 1.5.0
```

## Why --version takes precedence?

1. **Runtime flexibility**: Allows overriding hardcoded versions without code changes
2. **Testing**: Easy to test different versions without modifying scripts
3. **Backwards compatibility**: Scripts with @ notation still work, but can be overridden
4. **CI/CD friendly**: Environment variables or parameters can control versions dynamically

## Command Resolution Order

1. Parse command name and extract @ version if present
2. Parse all flags and look for --version
3. If --version is found, use that version
4. Otherwise, use @ version if present
5. If neither specified, use latest version
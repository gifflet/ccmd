# ccmd.yaml Schema

This package provides the schema definition and parsing functionality for `ccmd.yaml` files.

## Usage

```go
import "github.com/gifflet/ccmd/pkg/project"

// Load from file
config, err := project.LoadConfig("ccmd.yaml")
if err != nil {
    log.Fatal(err)
}

// Parse from reader
config, err := project.ParseConfig(reader)
if err != nil {
    log.Fatal(err)
}

// Access commands
for _, cmd := range config.Commands {
    owner, repo, _ := cmd.ParseOwnerRepo()
    fmt.Printf("Command: %s/%s@%s\n", owner, repo, cmd.Version)
    
    if cmd.IsSemanticVersion() {
        fmt.Println("Using semantic version")
    }
}
```

## Schema Structure

The `ccmd.yaml` file has a simple structure:

```yaml
commands:
  - repo: owner/repository
    version: v1.0.0  # optional
```

### Fields

- **commands**: List of command declarations (required)
  - **repo**: Repository in format "owner/repository" (required)
  - **version**: Version specification (optional, defaults to "latest")

### Version Specification

The version field supports:
- Semantic versions: `v1.0.0`, `1.2.3`, `v2.0.0-beta.1`
- Branch names: `main`, `develop`, `feature/xyz`
- Tag names: `release-1.0`, `stable`
- Special value: `latest` (default)

### Validation Rules

1. At least one command must be defined
2. Repository must be in "owner/repo" format
3. Owner and repo names must contain only alphanumeric characters, hyphens, and underscores
4. Version format is validated for basic correctness

## Example

```yaml
commands:
  # Semantic version
  - repo: example/tool
    version: v1.2.3
  
  # Latest version (explicit)
  - repo: org/cli
    version: latest
  
  # Branch
  - repo: user/command
    version: develop
  
  # Default to latest (version omitted)
  - repo: tools/formatter
```
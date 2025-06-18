# Update Command

The `update` command allows you to update installed commands to their latest versions.

## Usage

### Update a single command
```bash
ccmd update <command-name>
```

### Update all installed commands
```bash
ccmd update --all
```

## Features

- Version comparison using semantic versioning
- Support for both tagged releases and commit hashes
- Dual structure maintenance (directory and .md file)
- Progress indicators with spinners
- Comprehensive error handling and rollback
- Batch updates with summary report

## Implementation Details

The update command:
1. Checks if a command is installed
2. Clones the repository to check for latest version
3. Compares current version with latest available
4. Performs update if newer version is available
5. Updates both directory structure and standalone .md file
6. Updates the lock file with new version and timestamp

### Version Comparison Logic

- Semantic versions are compared properly (e.g., v1.0.0 < v1.1.0)
- Tagged versions are preferred over commit hashes
- If both versions are commits, they are considered different but not comparable

## Testing

The command includes comprehensive tests covering:
- Version comparison logic
- Single command updates
- Batch updates with --all flag
- Error handling scenarios
# Error Handling and Logging Guide

This guide explains how to use the error handling and logging infrastructure in CCMD.

## Error Handling

### Creating Errors

Use the `errors` package to create errors with context:

```go
import (
    "github.com/gifflet/ccmd/pkg/errors"
)

// Resource not found
err := errors.NotFound("command foo")

// Resource already exists  
err := errors.AlreadyExists("command bar")

// Invalid input
err := errors.InvalidInput("version must be semver format")

// Git operation errors
err := errors.GitError("clone", err)

// File operation errors
err := errors.FileError("read", configPath, err)
```

### Sentinel Errors

The package defines the following sentinel errors:

- `ErrNotFound` - Resource not found
- `ErrAlreadyExists` - Resource already exists
- `ErrInvalidInput` - Invalid input
- `ErrGitOperation` - Git operation failed
- `ErrFileOperation` - File operation failed

### Checking Error Types

Use Go's standard `errors.Is()` to check error types:

```go
import (
    "errors"
    errs "github.com/gifflet/ccmd/pkg/errors"
)

if errors.Is(err, errs.ErrNotFound) {
    // Handle not found error
}

if errors.Is(err, errs.ErrAlreadyExists) {
    // Handle already exists error
}

if errors.Is(err, errs.ErrGitOperation) {
    // Handle git-related error
}
```

## Logging

### Creating Loggers

```go
import "github.com/gifflet/ccmd/pkg/logger"

// Use default logger
logger.Info("Application started")

// Create component-specific logger
log := logger.WithField("component", "git")
log.Debug("Cloning repository")

// Create logger with multiple fields
log := logger.WithFields(logger.Fields{
    "user_id": userID,
    "action":  "install",
})
```

### Log Levels

- `Debug` - Detailed information for debugging
- `Info` - General informational messages
- `Warn` - Warning messages
- `Error` - Error messages
- `Fatal` - Fatal errors (exits program)

### Structured Logging

Always use structured logging for better searchability:

```go
log.WithFields(logger.Fields{
    "repository": url,
    "version":    version,
    "duration":   time.Since(start),
}).Info("Installation completed")
```

### Logging Errors

```go
if err != nil {
    log.WithError(err).Error("Operation failed")
    return err
}
```

## Integration with Commands

### Error Handling in Commands

Use the `errors.Handler` for consistent error handling:

```go
import (
    "github.com/gifflet/ccmd/pkg/errors"
    "github.com/spf13/cobra"
)

cmd := &cobra.Command{
    Use:   "install",
    Short: "Install a command",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := runInstall(args[0]); err != nil {
            errors.Handle(err)
            return err
        }
        return nil
    },
}
```

### Command Implementation Pattern

```go
func runInstall(repository string) error {
    // Create command logger
    log := logger.WithField("command", "install")
    
    // Log start
    log.WithField("repository", repository).Debug("starting installation")
    
    // Perform operations
    if err := validateRepository(repository); err != nil {
        log.WithError(err).Error("validation failed")
        return err
    }
    
    // Handle errors with context
    if err := gitClone(repository); err != nil {
        return errors.GitError("clone repository", err)
    }
    
    // Log success
    log.Info("installation completed successfully")
    return nil
}
```

## Best Practices

1. **Use the errors package**: Don't create errors with `fmt.Errorf` alone, use the provided error functions
2. **Add context**: Include relevant information in error messages
3. **Log at appropriate levels**: Use Debug for detailed info, Info for general messages
4. **Structure your logs**: Use fields instead of formatting messages
5. **Handle errors once**: Either log and return, or just return (let caller handle)
6. **Use component loggers**: Create loggers with component field for better filtering

## Migration from Old Code

### Before:
```go
if err != nil {
    return fmt.Errorf("git clone failed: %w", err)
}
```

### After:
```go
if err != nil {
    return errors.GitError("clone", err)
}
```

### Before:
```go
output.PrintErrorf("Command not found: %s", name)
```

### After:
```go
err := errors.NotFound(fmt.Sprintf("command %s", name))
errors.Handle(err)  // This will log and display user-friendly message
```

### Before:
```go
if err := os.ReadFile(path); err != nil {
    return fmt.Errorf("failed to read %s: %w", path, err)
}
```

### After:
```go
if err := os.ReadFile(path); err != nil {
    return errors.FileError("read", path, err)
}
```

## Environment Variables

- `CCMD_DEBUG=1` - Enable debug logging
- `CCMD_LOG_LEVEL=debug` - Set log level (debug, info, warn, error)

## Testing

When testing code that uses logging, note that the logger writes directly to stderr using Go's standard `slog` package. For testing purposes, focus on testing the behavior of your code rather than capturing log output.
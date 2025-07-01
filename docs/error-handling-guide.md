# Error Handling and Logging Guide

This guide explains how to use the comprehensive error handling and logging infrastructure in CCMD.

## Error Handling

### Creating Errors

Use the `errors` package to create structured errors:

```go
import "github.com/gifflet/ccmd/pkg/errors"

// Simple error
err := errors.New(errors.CodeNotFound, "command not found")

// Error with formatted message
err := errors.Newf(errors.CodeInvalidArgument, "invalid value: %s", value)

// Error with details
err := errors.New(errors.CodeConfigInvalid, "invalid configuration").
    WithDetail("file", "ccmd.yaml").
    WithDetail("line", 42)
```

### Error Codes

Use predefined error codes for consistency:

- `CodeNotFound` - Resource not found
- `CodeAlreadyExists` - Resource already exists
- `CodeInvalidArgument` - Invalid input
- `CodePermissionDenied` - Permission denied
- `CodeGitClone` - Git clone failed
- `CodeConfigInvalid` - Invalid configuration
- `CodeNetworkTimeout` - Network timeout

### Wrapping Errors

Wrap lower-level errors with context:

```go
output, err := cmd.CombinedOutput()
if err != nil {
    return errors.Wrap(err, errors.CodeGitClone, "git clone failed").
        WithDetail("repository", url).
        WithDetail("output", string(output))
}
```

### Checking Error Types

Use helper functions to check error types:

```go
if errors.IsNotFound(err) {
    // Handle not found error
}

if errors.IsPermissionDenied(err) {
    // Handle permission error
}

if errors.IsGitError(err) {
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

### Wrapping Commands

Use `WrapCommand` to add automatic error handling and logging:

```go
cmd := &cobra.Command{
    Use:   "install",
    Short: "Install a command",
    RunE: errors.WrapCommand("install", func(cmd *cobra.Command, args []string) error {
        // Command implementation
        return runInstall(args)
    }),
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
        return err  // Error will be handled by WrapCommand
    }
    
    // Log success
    log.Info("installation completed successfully")
    return nil
}
```

## Best Practices

1. **Always use error codes**: Don't create errors with `fmt.Errorf`, use the errors package
2. **Add context**: Include relevant details using `WithDetail`
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
    log.WithError(err).Error("git clone failed")
    return errors.Wrap(err, errors.CodeGitClone, "git clone failed").
        WithDetail("repository", url)
}
```

### Before:
```go
output.PrintErrorf("Command not found: %s", name)
```

### After:
```go
err := errors.New(errors.CodeCommandNotFound, "command not found").
    WithDetail("command", name)
errors.Handle(err)  // This will log and display user-friendly message
```

## Environment Variables

- `CCMD_DEBUG=1` - Enable debug logging
- `CCMD_LOG_LEVEL=debug` - Set log level (debug, info, warn, error)

## Testing

When testing code that uses logging, note that the logger writes directly to stderr using Go's standard `slog` package. For testing purposes, focus on testing the behavior of your code rather than capturing log output.
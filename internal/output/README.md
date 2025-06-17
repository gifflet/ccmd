# Output Package

The `output` package provides a consistent way to handle user feedback, error messages, and progress indicators in the ccmd CLI tool.

## Features

- **Color-coded output**: Success (green), Error (red), Warning (yellow), Info (blue)
- **User-friendly error messages**: Wrap technical errors with user-friendly explanations
- **Progress indicators**: Spinner for indeterminate progress, progress bar for measurable operations
- **Debug output**: Conditional debug messages controlled by environment variable

## Usage

### Basic Output

```go
import "github.com/gifflet/ccmd/internal/output"

// Print different types of messages
output.PrintSuccess("Operation completed successfully")
output.PrintError("Operation failed")
output.PrintWarning("This might cause issues")
output.PrintInfo("Processing files...")

// Fatal error (exits with code 1)
output.Fatal("Critical error: %s", err)
```

### User-Friendly Errors

```go
// Create user-friendly errors
err := output.NewUserError("Failed to load configuration", originalErr)
output.PrintUserError(err)

// Wrap existing errors
wrappedErr := output.WrapError(err, "Failed to initialize application")

// Check if error is user-friendly
if output.IsUserError(err) {
    output.PrintUserError(err)
} else {
    output.PrintError("Unexpected error: %v", err)
}
```

### Progress Indicators

#### Spinner (for indeterminate progress)

```go
spinner := output.NewSpinner("Processing...")
spinner.Start()

// Do work...

spinner.Success("Processing complete")
// or
spinner.Error("Processing failed")
```

#### Progress Bar (for measurable progress)

```go
items := []string{"file1", "file2", "file3"}
progress := output.NewProgressBar(len(items), "Processing files")

for _, item := range items {
    // Process item
    progress.Increment()
}

progress.Complete()
```

### Interactive Input

```go
// Get user input
name := output.Prompt("Enter your name")
```

### Debug Output

```go
// Set CCMD_DEBUG=1 to enable debug output
output.Debug("Detailed information: %v", details)
```

## Environment Variables

- `CCMD_DEBUG=1`: Enable debug output

## Testing

Run tests with:
```bash
go test ./internal/output/...
```
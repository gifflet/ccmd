// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package output_test

import (
	"errors"
	"time"

	"github.com/gifflet/ccmd/internal/output"
)

func ExamplePrintf() {
	// Basic output examples
	output.Printf("Regular message")
	output.PrintInfof("Information message")
	output.PrintSuccessf("Operation completed successfully")
	output.PrintWarningf("This is a warning")
	output.PrintErrorf("Something went wrong")
}

func ExampleUserError() {
	// Creating user-friendly errors
	err := output.NewUserError("Failed to load configuration", errors.New("file not found"))
	output.PrintUserError(err)

	// Using formatted error
	err2 := output.NewUserErrorf("Invalid command: %s", "unknown-cmd")
	output.PrintUserError(err2)
}

func ExampleSpinner() {
	// Using spinner for long operations
	spinner := output.NewSpinner("Processing files...")
	spinner.Start()

	// Simulate work
	time.Sleep(2 * time.Second)

	spinner.Success("Files processed successfully")

	// Or show error
	spinner2 := output.NewSpinner("Downloading...")
	spinner2.Start()
	time.Sleep(1 * time.Second)
	spinner2.Error("Download failed")
}

func ExampleProgressBar() {
	// Using progress bar for measurable operations
	items := []string{"file1.txt", "file2.txt", "file3.txt"}
	progress := output.NewProgressBar(len(items), "Processing files")

	for _, item := range items {
		// Process item
		_ = item
		time.Sleep(500 * time.Millisecond)
		progress.Increment()
	}

	progress.Complete()
}

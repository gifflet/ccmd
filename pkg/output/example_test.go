/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package output_test

import (
	"errors"
	"time"

	ccmderrors "github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/output"
)

func ExamplePrintf() {
	// Basic output examples
	output.Printf("Regular message")
	output.PrintInfof("Information message")
	output.PrintSuccessf("Operation completed successfully")
	output.PrintWarningf("This is a warning")
	output.PrintErrorf("Something went wrong")
}

func ExamplePrintError() {
	// Printing errors with appropriate formatting
	err := errors.New("file not found")
	output.PrintError(err)

	// Using pkg/errors for better error handling
	err2 := ccmderrors.NotFound("command unknown-cmd")
	output.PrintError(err2)
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

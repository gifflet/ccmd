/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package project_test

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/pkg/project"
)

func Example() {
	// Create a new lock file
	lockFile := project.NewLockFile()

	// Add a command entry
	ghCmd := &project.CommandLockInfo{
		Name:         "gh",
		Source:       "github.com/cli/cli",
		Version:      "v2.40.0",
		Commit:       strings.Repeat("a", 40), // Example SHA
		Resolved:     "github.com/cli/cli@v2.40.0",
		InstalledAt:  time.Now(),
		UpdatedAt:    time.Now(),
		Dependencies: []string{"git"},
		Metadata: map[string]string{
			"arch": "amd64",
			"os":   "darwin",
		},
	}

	if err := lockFile.AddCommand(ghCmd); err != nil {
		log.Fatal(err)
	}

	// Save to file
	lockFilePath := filepath.Join("/tmp", "ccmd-lock.yaml")
	fileSystem := fs.OS{}
	if err := lockFile.SaveToFile(lockFilePath, fileSystem); err != nil {
		log.Fatal(err)
	}

	// Load from file
	loaded, err := project.LoadFromFile(lockFilePath, fileSystem)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve a command
	if cmd, exists := loaded.GetCommand("gh"); exists {
		fmt.Printf("Command: %s\n", cmd.Name)
		fmt.Printf("Version: %s\n", cmd.Version)
		fmt.Printf("Repository: %s\n", cmd.Source)
	}

	// Output:
	// Command: gh
	// Version: v2.40.0
	// Repository: github.com/cli/cli
}

func ExampleLockFile_AddCommand() {
	lockFile := project.NewLockFile()

	// Add multiple commands
	commands := []*project.CommandLockInfo{
		{
			Name:        "cobra-cli",
			Source:      "github.com/spf13/cobra-cli",
			Version:     "v1.3.0",
			Commit:      strings.Repeat("c", 40),
			Resolved:    "github.com/spf13/cobra-cli@v1.3.0",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "golangci-lint",
			Source:      "github.com/golangci/golangci-lint",
			Version:     "v1.55.0",
			Commit:      strings.Repeat("e", 40),
			Resolved:    "github.com/golangci/golangci-lint@v1.55.0",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, cmd := range commands {
		if err := lockFile.AddCommand(cmd); err != nil {
			log.Printf("Failed to add %s: %v", cmd.Name, err)
		}
	}

	fmt.Printf("Total commands: %d\n", len(lockFile.Commands))
	// Output: Total commands: 2
}

func ExampleCalculateChecksum() {
	// In practice, this would be the path to an installed binary
	binaryPath := "/usr/local/bin/gh"

	// Calculate the SHA256 checksum
	checksum, err := project.CalculateChecksum(binaryPath)
	if err != nil {
		// Handle error - file might not exist in this example
		fmt.Println("Error calculating checksum")
		return
	}

	fmt.Printf("Checksum length: %d characters\n", len(checksum))
	// Output would be: Checksum length: 64 characters
}

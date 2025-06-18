package project_test

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gifflet/ccmd/pkg/project"
)

func Example() {
	// Create a new lock file
	lockFile := project.NewLockFile()

	// Add a command entry
	ghCmd := &project.Command{
		Name:         "gh",
		Repository:   "github.com/cli/cli",
		Version:      "v2.40.0",
		CommitHash:   strings.Repeat("a", 40), // Example SHA
		InstalledAt:  time.Now(),
		UpdatedAt:    time.Now(),
		FileSize:     45678901,
		Checksum:     strings.Repeat("b", 64), // Example SHA256
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
	if err := lockFile.SaveToFile(lockFilePath); err != nil {
		log.Fatal(err)
	}

	// Load from file
	loaded, err := project.LoadFromFile(lockFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve a command
	if cmd, exists := loaded.GetCommand("gh"); exists {
		fmt.Printf("Command: %s\n", cmd.Name)
		fmt.Printf("Version: %s\n", cmd.Version)
		fmt.Printf("Repository: %s\n", cmd.Repository)
	}

	// Output:
	// Command: gh
	// Version: v2.40.0
	// Repository: github.com/cli/cli
}

func ExampleLockFile_AddCommand() {
	lockFile := project.NewLockFile()

	// Add multiple commands
	commands := []*project.Command{
		{
			Name:        "cobra-cli",
			Repository:  "github.com/spf13/cobra-cli",
			Version:     "v1.3.0",
			CommitHash:  strings.Repeat("c", 40),
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
			FileSize:    12345678,
			Checksum:    strings.Repeat("d", 64),
		},
		{
			Name:        "golangci-lint",
			Repository:  "github.com/golangci/golangci-lint",
			Version:     "v1.55.0",
			CommitHash:  strings.Repeat("e", 40),
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
			FileSize:    87654321,
			Checksum:    strings.Repeat("f", 64),
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

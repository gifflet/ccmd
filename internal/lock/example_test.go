package lock_test

import (
	"fmt"
	"log"

	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
)

func ExampleManager() {
	// Create a new lock manager
	manager := lock.NewManager("/home/user/.claude/commands")

	// Load existing lock file or create new one
	if err := manager.Load(); err != nil {
		log.Fatal(err)
	}

	// Add a new command
	cmd := &models.Command{
		Name:    "git-helper",
		Version: "1.2.0",
		Source:  "github.com/example/git-helper",
		Metadata: map[string]string{
			"author":  "Example Author",
			"license": "MIT",
		},
	}

	if err := manager.AddCommand(cmd); err != nil {
		log.Fatal(err)
	}

	// Save changes to disk
	if err := manager.Save(); err != nil {
		log.Fatal(err)
	}

	// List all installed commands
	commands, err := manager.ListCommands()
	if err != nil {
		log.Fatal(err)
	}

	for _, cmd := range commands {
		fmt.Printf("Command: %s v%s\n", cmd.Name, cmd.Version)
	}
}

func ExampleManager_UpdateCommand() {
	manager := lock.NewManager("/home/user/.claude/commands")

	if err := manager.Load(); err != nil {
		log.Fatal(err)
	}

	// Update a command's version
	err := manager.UpdateCommand("git-helper", func(cmd *models.Command) error {
		cmd.Version = "1.3.0"
		cmd.Metadata["updated"] = "true"
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Save changes
	if err := manager.Save(); err != nil {
		log.Fatal(err)
	}
}

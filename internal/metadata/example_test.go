// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package metadata_test

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gifflet/ccmd/internal/metadata"
	"github.com/gifflet/ccmd/internal/models"
)

func ExampleManager_ReadCommandMetadata() {
	manager := metadata.NewManager()

	// Read metadata from a command directory
	commandDir := "/path/to/command"
	meta, err := manager.ReadCommandMetadata(commandDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Command: %s v%s\n", meta.Name, meta.Version)
	fmt.Printf("Author: %s\n", meta.Author)
	fmt.Printf("Description: %s\n", meta.Description)
}

func ExampleManager_WriteCommandMetadata() {
	manager := metadata.NewManager()

	// Create new metadata
	meta := &models.CommandMetadata{
		Name:        "my-command",
		Version:     "1.0.0",
		Description: "My awesome command",
		Author:      "John Doe",
		Repository:  "https://github.com/johndoe/my-command",
		Entry:       "main.go",
		Tags:        []string{"cli", "tool"},
		License:     "MIT",
	}

	// Write to command directory
	commandDir := "/path/to/command"
	if err := manager.WriteCommandMetadata(commandDir, meta); err != nil {
		log.Fatal(err)
	}
}

func ExampleManager_UpdateCommandMetadata() {
	manager := metadata.NewManager()
	commandDir := "/path/to/command"

	// Update existing metadata
	err := manager.UpdateCommandMetadata(commandDir, func(meta *models.CommandMetadata) error {
		meta.Version = "2.0.0"
		meta.Tags = append(meta.Tags, "updated")
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func Example_packageLevelFunctions() {
	commandDir := "/path/to/command"

	// Check if metadata exists
	if metadata.Exists(commandDir) {
		// Read metadata
		meta, err := metadata.ReadCommandMetadata(commandDir)
		if err != nil {
			log.Fatal(err)
		}

		// Extract command info
		info := metadata.ExtractCommandInfo(meta)
		fmt.Printf("Command info: %v\n", info)

		// Update metadata
		err = metadata.UpdateCommandMetadata(commandDir, func(m *models.CommandMetadata) error {
			m.Version = "1.1.0"
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Create new metadata
		meta := &models.CommandMetadata{
			Name:        "new-command",
			Version:     "1.0.0",
			Description: "A new command",
			Author:      "Author Name",
			Repository:  "https://github.com/author/new-command",
			Entry:       "new-command",
		}

		if err := metadata.WriteCommandMetadata(commandDir, meta); err != nil {
			log.Fatal(err)
		}
	}
}

func ExampleManager_Exists() {
	manager := metadata.NewManager()

	commandDirs := []string{
		"/path/to/command1",
		"/path/to/command2",
		"/path/to/command3",
	}

	for _, dir := range commandDirs {
		if manager.Exists(dir) {
			fmt.Printf("Metadata found in %s\n", filepath.Base(dir))
		} else {
			fmt.Printf("No metadata in %s\n", filepath.Base(dir))
		}
	}
}

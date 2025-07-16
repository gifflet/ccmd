/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gifflet/ccmd/pkg/errors"
)

// CommandDetail represents detailed information about an installed command
type CommandDetail struct {
	Name            string
	Version         string
	Description     string
	Author          string
	Repository      string
	UpdatedAt       string
	InstalledAt     string
	BrokenStructure bool
	StructureError  string
	// Additional metadata from ccmd.yaml
	Tags     []string
	License  string
	Homepage string
	Entry    string
	Requires string
	Resolved string
}

// ListOptions represents options for listing commands
type ListOptions struct {
	ProjectPath string // Path to project root
}

// List returns a list of all installed commands
func List(opts ListOptions) ([]CommandDetail, error) {
	if opts.ProjectPath == "" {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, errors.FileError("get working directory", "", err)
		}
		opts.ProjectPath = cwd
	}

	// Find project root
	projectRoot, err := findProjectRootFrom(opts.ProjectPath)
	if err != nil {
		return nil, err
	}

	// Read lock file
	lockPath := filepath.Join(projectRoot, "ccmd-lock.yaml")
	if !fileExists(lockPath) {
		// No lock file means no commands installed
		return []CommandDetail{}, nil
	}

	lockData, err := ReadLockFile(lockPath)
	if err != nil {
		return nil, err
	}

	// Build command list
	var commands []CommandDetail
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")

	for name, info := range lockData.Commands {
		cmd := CommandDetail{
			Name:        name,
			Version:     info.Version,
			Repository:  info.Source,
			UpdatedAt:   info.UpdatedAt.Format(time.RFC3339),
			InstalledAt: info.InstalledAt.Format(time.RFC3339),
			Resolved:    info.Resolved,
		}

		// Check command structure
		cmdDir := filepath.Join(commandsDir, name)
		mdFile := filepath.Join(commandsDir, name+".md")

		if !dirExists(cmdDir) {
			cmd.BrokenStructure = true
			cmd.StructureError = "command directory not found"
		} else if !fileExists(mdFile) {
			cmd.BrokenStructure = true
			cmd.StructureError = "standalone .md file not found"
		}

		// Read command metadata if available
		if dirExists(cmdDir) {
			metadataPath := filepath.Join(cmdDir, "ccmd.yaml")
			if metadata, err := readCommandMetadata(metadataPath); err == nil {
				// Use metadata values if available
				if metadata.Description != "" {
					cmd.Description = metadata.Description
				}
				if metadata.Author != "" {
					cmd.Author = metadata.Author
				}
				if metadata.Version != "" && cmd.Version == "" {
					cmd.Version = metadata.Version
				}
				cmd.Tags = metadata.Tags
				cmd.License = metadata.License
				cmd.Homepage = metadata.Homepage
				cmd.Entry = metadata.Entry
				// Requires field doesn't exist in current metadata model
			}
		}

		commands = append(commands, cmd)
	}

	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands, nil
}

// GetCommandInfo returns detailed information about a specific command
func GetCommandInfo(name, projectPath string) (*CommandDetail, error) {
	commands, err := List(ListOptions{ProjectPath: projectPath})
	if err != nil {
		return nil, err
	}

	for _, cmd := range commands {
		if cmd.Name == name {
			return &cmd, nil
		}
	}

	return nil, errors.NotFound(fmt.Sprintf("command %q", name))
}

func findProjectRootFrom(startPath string) (string, error) {
	dir := startPath

	for {
		if fileExists(filepath.Join(dir, "ccmd.yaml")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding ccmd.yaml
			// Use start path as project root
			return startPath, nil
		}
		dir = parent
	}
}

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
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/output"
)

// RemoveOptions represents options for removing a command
type RemoveOptions struct {
	Name        string // Command name to remove
	Force       bool   // Remove without confirmation
	UpdateFiles bool   // Update ccmd.yaml and ccmd-lock.yaml
}

// Remove removes an installed command
func Remove(opts RemoveOptions) error {
	if opts.Name == "" {
		return errors.InvalidInput("command name is required")
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	// Check if command exists in lock file
	lockPath := filepath.Join(projectRoot, LockFileName)
	if !fileExists(lockPath) {
		return errors.NotFound("no commands installed (ccmd-lock.yaml not found)")
	}

	// Read lock file using shared function
	lockFile, err := ReadLockFile(lockPath)
	if err != nil {
		return err
	}

	cmdInfo, exists := lockFile.Commands[opts.Name]
	if !exists {
		return errors.NotFound(fmt.Sprintf("command %q", opts.Name))
	}

	// Paths to remove
	commandDir := filepath.Join(projectRoot, ".claude", "commands", opts.Name)
	mdFile := filepath.Join(projectRoot, ".claude", "commands", opts.Name+".md")

	// Show what will be removed
	output.PrintInfof("Will remove command %q", opts.Name)
	output.PrintInfof("Repository: %s", cmdInfo.Source)
	if cmdInfo.Version != "" {
		output.PrintInfof("Version: %s", cmdInfo.Version)
	}

	// Remove command directory
	if dirExists(commandDir) {
		output.PrintInfof("Removing command directory...")
		if err := os.RemoveAll(commandDir); err != nil {
			return errors.FileError("remove command directory", commandDir, err)
		}
	}

	// Remove standalone .md file
	if fileExists(mdFile) {
		output.PrintInfof("Removing documentation file...")
		if err := os.Remove(mdFile); err != nil {
			// Try to restore command directory if md removal fails
			output.PrintWarningf("Failed to remove .md file: %v", err)
		}
	}

	// Update lock file
	delete(lockFile.Commands, opts.Name)

	// Save updated lock file
	if err := WriteLockFile(lockPath, lockFile); err != nil {
		return err
	}

	// Update ccmd.yaml if requested
	if opts.UpdateFiles {
		if err := removeFromConfig(projectRoot, opts.Name, cmdInfo.Source); err != nil {
			output.PrintWarningf("Failed to update ccmd.yaml: %v", err)
		} else {
			output.PrintInfof("Updated ccmd.yaml")
		}
	}

	output.PrintSuccessf("Command %q removed successfully", opts.Name)
	return nil
}

// removeFromConfig removes a command from ccmd.yaml
func removeFromConfig(projectRoot, name, repository string) error {
	configPath := filepath.Join(projectRoot, "ccmd.yaml")
	if !fileExists(configPath) {
		return nil // No config file to update
	}

	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Get commands array
	commandsRaw, ok := config["commands"]
	if !ok {
		return nil // No commands in config
	}

	commands, ok := commandsRaw.([]interface{})
	if !ok {
		return nil // Invalid format
	}

	// Find and remove the command
	newCommands := make([]interface{}, 0, len(commands))
	removed := false

	for _, cmd := range commands {
		// Handle string format only (e.g., "owner/repo@version")
		cmdStr, ok := cmd.(string)
		if !ok {
			// Skip non-string entries
			newCommands = append(newCommands, cmd)
			continue
		}

		// Extract repo from string format
		parts := strings.Split(cmdStr, "@")
		cmdRepo := parts[0]

		// Try to match by repository or by extracting name from repo
		if cmdRepo == repository || extractCommandName(cmdRepo) == name {
			removed = true
			continue // Skip this command
		}

		newCommands = append(newCommands, cmd)
	}

	if !removed {
		return nil // Command not found in config
	}

	// Update config
	config["commands"] = newCommands

	// Write back
	output, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, output, 0644)
}

// ListCommands returns a list of all command names
func ListCommands(projectPath string) ([]string, error) {
	commands, err := List(ListOptions{ProjectPath: projectPath})
	if err != nil {
		return nil, err
	}

	names := make([]string, len(commands))
	for i, cmd := range commands {
		names[i] = cmd.Name
	}

	return names, nil
}

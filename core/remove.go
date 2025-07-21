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
	Name        string
	Force       bool
	UpdateFiles bool
}

// Remove removes an installed command
func Remove(opts RemoveOptions) error {
	if opts.Name == "" {
		return errors.InvalidInput("command name is required")
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	lockPath := filepath.Join(projectRoot, LockFileName)
	if !fileExists(lockPath) {
		return errors.NotFound("no commands installed (ccmd-lock.yaml not found)")
	}

	lockFile, err := ReadLockFile(lockPath)
	if err != nil {
		return err
	}

	cmdInfo, exists := lockFile.Commands[opts.Name]
	if !exists {
		return errors.NotFound(fmt.Sprintf("command %q", opts.Name))
	}

	if err := removeCommandFiles(projectRoot, opts.Name); err != nil {
		return err
	}

	output.PrintInfof("Will remove command %q", opts.Name)
	output.PrintInfof("Repository: %s", cmdInfo.Source)
	if cmdInfo.Version != "" {
		output.PrintInfof("Version: %s", cmdInfo.Version)
	}

	delete(lockFile.Commands, opts.Name)

	if err := WriteLockFile(lockPath, lockFile); err != nil {
		return err
	}

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

func removeFromConfig(projectRoot, name, repository string) error {
	configPath := filepath.Join(projectRoot, "ccmd.yaml")
	if !fileExists(configPath) {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	commandsRaw, ok := config["commands"]
	if !ok {
		return nil
	}

	commands, ok := commandsRaw.([]interface{})
	if !ok {
		return nil
	}

	newCommands := make([]interface{}, 0, len(commands))
	removed := false

	for _, cmd := range commands {
		cmdStr, ok := cmd.(string)
		if !ok {
			newCommands = append(newCommands, cmd)
			continue
		}

		parts := strings.Split(cmdStr, "@")
		cmdRepo := parts[0]

		if cmdRepo == repository || extractCommandName(cmdRepo) == name {
			removed = true
			continue
		}

		newCommands = append(newCommands, cmd)
	}

	if !removed {
		return nil
	}

	config["commands"] = newCommands

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

func removeCommandFiles(projectRoot, name string) error {
	commandDir := filepath.Join(projectRoot, ".claude", "commands", name)
	mdFile := filepath.Join(projectRoot, ".claude", "commands", name+".md")

	if dirExists(commandDir) {
		output.PrintInfof("Removing command directory...")
		if err := os.RemoveAll(commandDir); err != nil {
			return errors.FileError("remove command directory", commandDir, err)
		}
	}

	if fileExists(mdFile) {
		output.PrintInfof("Removing md file...")
		if err := os.Remove(mdFile); err != nil {
			output.PrintWarningf("Failed to remove .md file: %v", err)
		}
	}

	return nil
}

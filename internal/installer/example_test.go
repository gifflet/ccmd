/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package installer_test

import (
	"context"
	"fmt"
	"log"

	"github.com/gifflet/ccmd/internal/installer"
)

// Example demonstrates basic command installation
func Example() {
	// Create installer with options
	opts := installer.Options{
		Repository: "https://github.com/example/mycmd.git",
		Version:    "v1.0.0",
		InstallDir: ".claude/commands",
	}

	inst, err := installer.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	// Install the command
	ctx := context.Background()
	if err := inst.Install(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Command installed successfully")
}

// Example_withCustomName demonstrates installation with a custom command name
func Example_withCustomName() {
	opts := installer.Options{
		Repository: "https://github.com/example/long-command-name.git",
		Name:       "lcn", // Short alias
		InstallDir: ".claude/commands",
	}

	inst, err := installer.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := inst.Install(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Command installed as 'lcn'")
}

// Example_forceReinstall demonstrates force reinstallation
func Example_forceReinstall() {
	opts := installer.Options{
		Repository: "https://github.com/example/mycmd.git",
		Version:    "v2.0.0",
		Force:      true, // Overwrite existing installation
		InstallDir: ".claude/commands",
	}

	inst, err := installer.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := inst.Install(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Command reinstalled with version v2.0.0")
}

// Example_installFromShorthand demonstrates GitHub shorthand installation
func Example_installFromShorthand() {
	// Use the integration function for shorthand support
	ctx := context.Background()

	opts := installer.IntegrationOptions{
		Repository: "user/repo", // GitHub shorthand
		Version:    "latest",
	}

	if err := installer.InstallCommand(ctx, opts, true); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Command installed from user/repo")
}

// Example_installFromConfig demonstrates installing all commands from ccmd.yaml
func Example_installFromConfig() {
	ctx := context.Background()
	projectPath := "." // Current directory

	// Install all commands defined in ccmd.yaml
	if err := installer.InstallFromConfig(ctx, projectPath, false); err != nil {
		log.Fatal(err)
	}

	fmt.Println("All commands from ccmd.yaml installed")
}

// Example_commandManager demonstrates using the CommandManager
func Example_commandManager() {
	// Create a command manager for the current project
	cm := installer.NewCommandManager(".")

	ctx := context.Background()

	// Install a specific command
	if err := cm.Install(ctx, "example/tool", "v1.0.0", "", false); err != nil {
		log.Fatal(err)
	}

	// List installed commands
	commands, err := cm.GetInstalledCommands()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Installed commands:\n")
	for _, cmd := range commands {
		fmt.Printf("- %s (v%s): %s\n", cmd.Name, cmd.Version, cmd.Description)
	}
}

// Example_errorHandling demonstrates error handling during installation
func Example_errorHandling() {
	opts := installer.Options{
		Repository: "https://github.com/nonexistent/repo.git",
		InstallDir: ".claude/commands",
	}

	inst, err := installer.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := inst.Install(ctx); err != nil {
		// Check if error is retryable
		if installer.IsRetryableError(err) {
			fmt.Println("Installation failed but can be retried")
		} else {
			fmt.Println("Installation failed permanently")
		}

		return
	}

	fmt.Println("Command installed successfully")
}

// Example_parseRepositorySpec demonstrates parsing repository specifications
func Example_parseRepositorySpec() {
	specs := []string{
		"user/repo",
		"user/repo@v1.0.0",
		"https://github.com/user/repo.git@feature/branch",
		"git@github.com:user/repo.git",
	}

	for _, spec := range specs {
		repo, version := installer.ParseRepositorySpec(spec)
		if version == "" {
			fmt.Printf("%s -> repo: %s, version:\n", spec, repo)
		} else {
			fmt.Printf("%s -> repo: %s, version: %s\n", spec, repo, version)
		}
	}

	// Output:
	// user/repo -> repo: user/repo, version:
	// user/repo@v1.0.0 -> repo: user/repo, version: v1.0.0
	// https://github.com/user/repo.git@feature/branch -> repo: https://github.com/user/repo.git, version: feature/branch
	// git@github.com:user/repo.git -> repo: git@github.com:user/repo.git, version:
}

// Example_normalizeRepositoryURL demonstrates URL normalization
func Example_normalizeRepositoryURL() {
	urls := []string{
		"user/repo",
		"github.com/user/repo",
		"https://github.com/user/repo",
		"git@github.com:user/repo.git",
	}

	for _, url := range urls {
		normalized := installer.NormalizeRepositoryURL(url)
		fmt.Printf("%s -> %s\n", url, normalized)
	}

	// Output:
	// user/repo -> https://github.com/user/repo.git
	// github.com/user/repo -> https://github.com/user/repo.git
	// https://github.com/user/repo -> https://github.com/user/repo.git
	// git@github.com:user/repo.git -> git@github.com:user/repo.git
}

// Example_extractRepoPath demonstrates extracting repository paths
func Example_extractRepoPath() {
	urls := []string{
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"https://gitlab.com/org/project.git",
	}

	for _, url := range urls {
		path := installer.ExtractRepoPath(url)
		fmt.Printf("%s -> %s\n", url, path)
	}

	// Output:
	// https://github.com/user/repo.git -> user/repo
	// git@github.com:user/repo.git -> user/repo
	// https://gitlab.com/org/project.git -> org/project
}

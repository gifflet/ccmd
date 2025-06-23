/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package version_test

import (
	"fmt"
	"log"

	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/pkg/version"
)

func ExampleResolver_ResolveVersion() {
	// Create a git client
	gitClient := git.NewClient("/tmp")

	// Create a version resolver using the git client's GetTags method
	resolver := version.NewResolver(gitClient.GetTags)

	// Example 1: Resolve "latest" to the highest semantic version
	resolved, err := resolver.ResolveVersion("/path/to/repo", "latest")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Latest version: %s\n", resolved)

	// Example 2: Resolve a semantic version constraint
	resolved, err = resolver.ResolveVersion("/path/to/repo", "^1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Version matching ^1.0.0: %s\n", resolved)

	// Example 3: Resolve an exact tag
	resolved, err = resolver.ResolveVersion("/path/to/repo", "v1.2.3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Exact tag: %s\n", resolved)

	// Example 4: Pass through a branch name
	resolved, err = resolver.ResolveVersion("/path/to/repo", "main")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Branch: %s\n", resolved)
}

// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
// Package version integration example
package version

import (
	"fmt"

	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/models"
)

// IntegrationExample shows how to integrate the version resolver with the rest of ccmd.
func IntegrationExample() {
	// Example of how the install command would use the version resolver

	// 1. Parse command metadata from ccmd.yaml
	metadata := &models.CommandMetadata{
		Version: "^1.0.0", // Could be any version format
	}

	// 2. Create git client
	gitClient := git.NewClient("/tmp")

	// 3. Create version resolver
	resolver := NewResolver(gitClient.GetTags)

	// 4. Resolve the version
	repoPath := "/path/to/cloned/repo"
	resolvedVersion, err := resolver.ResolveVersion(repoPath, metadata.Version)
	if err != nil {
		// Handle error - version could not be resolved
		fmt.Printf("Failed to resolve version %s: %v\n", metadata.Version, err)
		return
	}

	// 5. Use the resolved version to checkout
	if err := gitClient.CheckoutTag(repoPath, resolvedVersion); err != nil {
		fmt.Printf("Failed to checkout %s: %v\n", resolvedVersion, err)
		return
	}

	fmt.Printf("Successfully resolved %s to %s and checked out\n", metadata.Version, resolvedVersion)
}

// ResolveCommandVersion is a helper function that can be used by commands.
func ResolveCommandVersion(gitClient *git.Client, repoPath, version string) (string, error) {
	resolver := NewResolver(gitClient.GetTags)
	return resolver.ResolveVersion(repoPath, version)
}

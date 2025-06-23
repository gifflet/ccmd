/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package sync_test

import (
	"fmt"
	"os"
)

// Example demonstrates how to use the sync command
func Example() {
	// First, create a ccmd.yaml file with declared dependencies
	// ccmd.yaml:
	// commands:
	//   - repo: owner/tool1
	//     version: v1.0.0
	//   - repo: owner/tool2
	//     version: v2.0.0

	// Run sync command
	// This will:
	// 1. Read ccmd.yaml
	// 2. Compare with installed commands
	// 3. Install missing commands
	// 4. Remove commands not in ccmd.yaml (with confirmation)
	// 5. Update ccmd-lock.yaml

	// Example command line usage:
	fmt.Println("ccmd sync")
	fmt.Println("ccmd sync --dry-run  # Preview changes")
	fmt.Println("ccmd sync --force    # Skip confirmation prompts")
}

// Example_dryRun demonstrates using sync with dry-run flag
func Example_dryRun() {
	// Set up environment
	os.Setenv("CCMD_DRY_RUN", "true")
	defer os.Unsetenv("CCMD_DRY_RUN")

	// In dry-run mode, sync will:
	// - Show what would be installed
	// - Show what would be removed
	// - Not make any actual changes

	fmt.Println("Running sync in dry-run mode...")
	// Output would show:
	// Sync Analysis:
	//
	// Commands to install:
	//   + owner/tool3@v3.0.0
	//
	// Commands to remove (not in ccmd.yaml):
	//   - tool4
	//
	// Dry run mode - no changes were made
}

// Example_force demonstrates using sync with force flag
func Example_force() {
	// Set up environment
	os.Setenv("CCMD_FORCE", "true")
	defer os.Unsetenv("CCMD_FORCE")

	// In force mode, sync will:
	// - Skip confirmation prompts for removals
	// - Proceed with all operations automatically

	fmt.Println("Running sync in force mode...")
	// Output would show:
	// Executing sync...
	//
	// Installing commands...
	// ✓ Installed owner/tool1
	//
	// Removing commands...
	// ✓ Removed tool2
	//
	// Updated ccmd-lock.yaml
	// Sync completed successfully!
}

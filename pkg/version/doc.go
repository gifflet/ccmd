/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package version provides version resolution functionality for commands.
//
// # Version Resolution
//
// The version resolver supports multiple formats for specifying versions:
//
// 1. "latest" - Resolves to the highest semantic version tag
//
// 2. Semantic version constraints:
//   - Caret ranges: ^1.2.3 (compatible with 1.2.3)
//   - Tilde ranges: ~1.2.3 (approximately 1.2.3)
//   - Comparisons: >1.0.0, >=1.0.0, <2.0.0, <=2.0.0
//   - Exact: =1.0.0
//   - Ranges: 1.0.0 - 2.0.0
//   - Wildcards: 1.x, 1.2.*
//
// 3. Exact tag names (with or without 'v' prefix):
//   - v1.2.3
//   - 1.2.3
//
// 4. Branch names:
//   - main
//   - develop
//   - feature/new-feature
//
// 5. Commit hashes:
//   - Full: abc123def456...
//   - Short: abc123d
//
// # Precedence
//
// When resolving versions, the resolver follows this precedence:
// 1. "latest" keyword resolution
// 2. Semantic version constraint matching
// 3. Exact tag matching (with v-prefix flexibility)
// 4. Pass-through for branches/commits
//
// # Semantic Version Handling
//
// The resolver handles semantic version tags with or without the 'v' prefix.
// When matching exact versions, it will try both formats automatically.
//
// Example usage:
//
//	resolver := version.NewResolver(gitClient.GetTags)
//	resolved, err := resolver.ResolveVersion("/path/to/repo", "^1.0.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// resolved might be "v1.5.2" - the highest version matching ^1.0.0
package version

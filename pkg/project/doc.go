// Package project provides functionality for parsing and validating ccmd.yaml configuration files.
//
// The ccmd.yaml file declares Claude commands required by a project. It should be placed
// at the root of your project directory.
//
// Schema Format:
//
// The ccmd.yaml file has a simple structure:
//
//	commands:
//	  - repo: owner/repository
//	    version: v1.0.0  # optional, defaults to latest
//
// Example ccmd.yaml:
//
//	commands:
//	  - repo: example/claude-command
//	    version: v1.2.3
//	  - repo: another/command
//	    version: latest
//	  - repo: org/tool
//	    # version omitted, defaults to latest
//
// Version Specification:
//
// The version field supports:
//   - Semantic versions: v1.0.0, 1.2.3, v2.0.0-beta.1
//   - Branch names: main, develop, feature/xyz
//   - Tag names: release-1.0, stable
//   - "latest" or omitted: uses the latest release
//
// Repository Format:
//
// Repositories must be in the format "owner/repository" where:
//   - owner: GitHub username or organization
//   - repository: Repository name
//
// Both must contain only alphanumeric characters, hyphens, and underscores.
package project

// Package project provides functionality for managing ccmd project files,
// including parsing and validating ccmd.yaml configuration files and
// managing the ccmd-lock.yaml lock file format.
//
// # Configuration File (ccmd.yaml)
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
//
// # Lock File Format (ccmd-lock.yaml)
//
// The ccmd-lock.yaml file tracks exact versions and metadata for installed commands.
// It ensures reproducible installations by locking specific commit hashes and
// providing integrity verification through checksums.
//
// Example ccmd-lock.yaml:
//
//	version: "1.0"
//	updated_at: 2024-01-15T10:30:00Z
//	commands:
//	  gh:
//	    name: gh
//	    repository: github.com/cli/cli
//	    version: v2.40.0
//	    commit_hash: abc123def456abc123def456abc123def456abc1
//	    installed_at: 2024-01-10T14:22:00Z
//	    updated_at: 2024-01-15T10:30:00Z
//	    file_size: 45678901
//	    checksum: 1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
//	    dependencies:
//	      - git
//	    metadata:
//	      arch: amd64
//	      os: darwin
//	  cobra-cli:
//	    name: cobra-cli
//	    repository: github.com/spf13/cobra-cli
//	    version: v1.3.0
//	    commit_hash: def456abc123def456abc123def456abc123def4
//	    installed_at: 2024-01-12T09:15:00Z
//	    updated_at: 2024-01-12T09:15:00Z
//	    file_size: 12345678
//	    checksum: abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
//
// Lock File Fields:
//
//   - version: Lock file format version (currently "1.0")
//   - updated_at: Last modification timestamp of the lock file
//   - commands: Map of installed commands by name
//
// Command Fields:
//
//   - name: Command name (must match the map key)
//   - repository: Source repository URL
//   - version: Version specifier used during installation (tag, branch, or commit)
//   - commit_hash: Exact 40-character git commit SHA
//   - installed_at: Initial installation timestamp
//   - updated_at: Last update timestamp
//   - file_size: Size of the installed binary in bytes
//   - checksum: SHA256 hash of the installed binary (64 characters)
//   - dependencies: Optional list of runtime dependencies
//   - metadata: Optional key-value pairs for additional information
//
// The lock file ensures that:
//   - Installations can be reproduced exactly using the commit hash
//   - Binary integrity can be verified using the checksum
//   - Installation history is tracked with timestamps
//   - Dependencies and metadata provide context for troubleshooting
package project

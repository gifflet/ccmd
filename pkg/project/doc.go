// Package project provides functionality for managing ccmd project files,
// including the ccmd-lock.yaml file format.
//
// # Lock File Format
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
// Fields:
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

package project

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

// Config represents the ccmd.yaml configuration file structure
type Config struct {
	Commands []ConfigCommand `yaml:"commands"`
}

// ConfigCommand represents a single command declaration in ccmd.yaml
type ConfigCommand struct {
	Repo    string `yaml:"repo"`
	Version string `yaml:"version,omitempty"`
}

// Validate performs validation on the Config
func (c *Config) Validate() error {
	if len(c.Commands) == 0 {
		return fmt.Errorf("no commands defined")
	}

	for i, cmd := range c.Commands {
		if err := cmd.Validate(); err != nil {
			return fmt.Errorf("command %d: %w", i, err)
		}
	}

	return nil
}

// Validate performs validation on a ConfigCommand
func (c *ConfigCommand) Validate() error {
	if c.Repo == "" {
		return fmt.Errorf("repo is required")
	}

	if err := validateRepoFormat(c.Repo); err != nil {
		return fmt.Errorf("invalid repo format: %w", err)
	}

	if c.Version != "" {
		if err := validateVersion(c.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	return nil
}

// ParseOwnerRepo extracts owner and repo name from the repo field
func (c *ConfigCommand) ParseOwnerRepo() (owner, repo string, err error) {
	parts := strings.Split(c.Repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format: expected owner/repo")
	}
	return parts[0], parts[1], nil
}

// IsSemanticVersion checks if the version is a semantic version
func (c *ConfigCommand) IsSemanticVersion() bool {
	if c.Version == "" || c.Version == "latest" {
		return false
	}
	_, err := semver.NewVersion(c.Version)
	return err == nil
}

// validateRepoFormat validates the repository format (owner/repo)
func validateRepoFormat(repo string) error {
	if repo == "" {
		return fmt.Errorf("repo cannot be empty")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("expected format: owner/repo")
	}

	owner, repoName := parts[0], parts[1]
	if owner == "" || repoName == "" {
		return fmt.Errorf("owner and repo name cannot be empty")
	}

	// Basic validation for GitHub username/org and repo name
	validName := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-_])*[a-zA-Z0-9]?$`)
	if !validName.MatchString(owner) {
		return fmt.Errorf("invalid owner name: %s", owner)
	}
	if !validName.MatchString(repoName) {
		return fmt.Errorf("invalid repo name: %s", repoName)
	}

	return nil
}

// validateVersion validates version format (semantic version, branch, or tag)
func validateVersion(version string) error {
	if version == "" || version == "latest" {
		return nil
	}

	// Try to parse as semantic version
	if _, err := semver.NewVersion(version); err == nil {
		return nil
	}

	// Basic validation for branch/tag names
	if strings.Contains(version, "..") || strings.HasPrefix(version, ".") || strings.HasSuffix(version, ".") {
		return fmt.Errorf("invalid version format")
	}

	return nil
}

// LoadConfig loads and parses a ccmd.yaml file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path) // #nosec G304 - path is expected to be user-provided
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck
	}()

	return ParseConfig(file)
}

// ParseConfig parses ccmd.yaml content from a reader
func ParseConfig(r io.Reader) (*Config, error) {
	var config Config

	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true) // Strict mode - fail on unknown fields

	if err := decoder.Decode(&config); err != nil {
		if err == io.EOF {
			// Empty file - treat as no commands defined
			return nil, fmt.Errorf("invalid configuration: no commands defined")
		}
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

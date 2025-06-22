package project

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
)

// Config represents the ccmd.yaml configuration file structure
type Config struct {
	// Project metadata
	Name        string `yaml:"name,omitempty"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	Author      string `yaml:"author,omitempty"`
	Repository  string `yaml:"repository,omitempty"`
	Entry       string `yaml:"entry,omitempty"`

	// Commands can be either strings or ConfigCommand objects
	Commands interface{} `yaml:"commands"`
}

// ConfigCommand represents a single command declaration in ccmd.yaml
type ConfigCommand struct {
	Repo    string `yaml:"repo"`
	Version string `yaml:"version,omitempty"`
}

// GetCommands returns a normalized list of ConfigCommand objects
func (c *Config) GetCommands() ([]ConfigCommand, error) {
	if c.Commands == nil {
		return []ConfigCommand{}, nil
	}

	var commands []ConfigCommand

	switch v := c.Commands.(type) {
	case []interface{}:
		// Handle array of strings only
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("command %d: must be a string (e.g., \"owner/repo@version\")", i)
			}
			cmd := parseCommandString(str)
			commands = append(commands, cmd)
		}
	case []string:
		// Handle pure string array
		for _, str := range v {
			cmd := parseCommandString(str)
			commands = append(commands, cmd)
		}
	case []ConfigCommand:
		// Already in correct format
		commands = v
	default:
		return nil, fmt.Errorf("commands must be an array of strings")
	}

	return commands, nil
}

// parseCommandString parses a command string in format "repo" or "repo@version"
func parseCommandString(s string) ConfigCommand {
	parts := strings.SplitN(s, "@", 2)
	cmd := ConfigCommand{
		Repo: parts[0],
	}
	if len(parts) > 1 {
		cmd.Version = parts[1]
	}
	return cmd
}

// Validate performs validation on the Config
func (c *Config) Validate() error {
	// Get normalized commands
	commands, err := c.GetCommands()
	if err != nil {
		return err
	}

	// Validate each command
	for i, cmd := range commands {
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
func LoadConfig(path string, fileSystem fs.FileSystem) (*Config, error) {
	data, err := fileSystem.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return ParseConfig(strings.NewReader(string(data)))
}

// ParseConfig parses ccmd.yaml content from a reader
func ParseConfig(r io.Reader) (*Config, error) {
	var config Config

	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true) // Strict mode - fail on unknown fields

	if err := decoder.Decode(&config); err != nil {
		if err == io.EOF {
			// Empty file - create empty config
			config.Commands = []ConfigCommand{}
			return &config, nil
		}
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveConfig saves a Config to a ccmd.yaml file
func SaveConfig(config *Config, path string, fileSystem fs.FileSystem) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Convert to save format
	saveConfig := config.toSaveFormat()

	// Config files are readable by all (0644)
	return writeYAMLFile(path, saveConfig, 0o644, fileSystem)
}

// toSaveFormat converts Config to a format that saves as string array when possible
func (c *Config) toSaveFormat() interface{} {
	// Get normalized commands
	commands, err := c.GetCommands()
	if err != nil || len(commands) == 0 {
		return c
	}

	// Check if all commands can be represented as strings
	allSimple := true
	for _, cmd := range commands {
		if cmd.Version != "" {
			// Still simple if version is embedded in repo
			if !strings.Contains(cmd.Repo, "@") {
				allSimple = false
				break
			}
		}
	}

	if allSimple {
		// Convert to string array format
		stringCommands := make([]string, len(commands))
		for i, cmd := range commands {
			if cmd.Version != "" && !strings.Contains(cmd.Repo, "@") {
				stringCommands[i] = cmd.Repo + "@" + cmd.Version
			} else {
				stringCommands[i] = cmd.Repo
			}
		}

		// Return a simplified structure
		type simpleConfig struct {
			Name        string   `yaml:"name,omitempty"`
			Version     string   `yaml:"version,omitempty"`
			Description string   `yaml:"description,omitempty"`
			Author      string   `yaml:"author,omitempty"`
			Repository  string   `yaml:"repository,omitempty"`
			Entry       string   `yaml:"entry,omitempty"`
			Commands    []string `yaml:"commands"`
		}
		return &simpleConfig{
			Name:        c.Name,
			Version:     c.Version,
			Description: c.Description,
			Author:      c.Author,
			Repository:  c.Repository,
			Entry:       c.Entry,
			Commands:    stringCommands,
		}
	}

	// Keep original format
	return c
}

// WriteConfig writes a Config to an io.Writer
func WriteConfig(config *Config, w io.Writer) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Convert to save format
	saveConfig := config.toSaveFormat()

	encoder := yaml.NewEncoder(w)
	defer func() {
		_ = encoder.Close() //nolint:errcheck // Best effort
	}()

	if err := encoder.Encode(saveConfig); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

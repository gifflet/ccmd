/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/pkg/errors"
)

// InitOptions represents options for initializing a project
type InitOptions struct {
	Name        string
	Version     string
	Description string
	Author      string
	Repository  string
	Entry       string
	Tags        []string
	ProjectPath string
	CommandMode bool // true for command mode, false for project mode
}

// InitDefaults returns default values for init based on current directory
func InitDefaults(projectPath string) InitOptions {
	dirName := filepath.Base(projectPath)

	return InitOptions{
		Name:        dirName,
		Version:     "1.0.0",
		Description: "",
		Author:      "",
		Repository:  "",
		Entry:       "index.md",
		Tags:        []string{},
		ProjectPath: projectPath,
		CommandMode: false,
	}
}

// LoadExistingConfig attempts to load existing ccmd.yaml and returns defaults
func LoadExistingConfig(projectPath string) (InitOptions, interface{}, error) {
	defaults := InitDefaults(projectPath)
	ccmdPath := filepath.Join(projectPath, ConfigFileName)

	// Check if file exists
	if _, err := os.Stat(ccmdPath); os.IsNotExist(err) {
		return defaults, nil, nil
	}

	// Read existing file
	data, err := os.ReadFile(ccmdPath)
	if err != nil {
		return defaults, nil, errors.FileError("read config", ccmdPath, err)
	}

	// Parse as raw YAML to preserve structure
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return defaults, nil, errors.FileError("parse config", ccmdPath, err)
	}

	// Extract fields for defaults
	if name, ok := rawConfig["name"].(string); ok && name != "" {
		defaults.Name = name
	}
	if version, ok := rawConfig["version"].(string); ok && version != "" {
		defaults.Version = version
	}
	if desc, ok := rawConfig["description"].(string); ok {
		defaults.Description = desc
	}
	if author, ok := rawConfig["author"].(string); ok {
		defaults.Author = author
	}
	if repo, ok := rawConfig["repository"].(string); ok {
		defaults.Repository = repo
	}
	if entry, ok := rawConfig["entry"].(string); ok && entry != "" {
		defaults.Entry = entry
	}

	// Extract tags
	if tags, ok := rawConfig["tags"].([]interface{}); ok {
		defaults.Tags = []string{}
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				defaults.Tags = append(defaults.Tags, tagStr)
			}
		}
	}

	// Preserve commands field
	existingCommands := rawConfig["commands"]

	return defaults, existingCommands, nil
}

// orderedConfig ensures consistent field ordering in YAML output
type orderedConfig struct {
	Name        string      `yaml:"name,omitempty"`
	Version     string      `yaml:"version,omitempty"`
	Description string      `yaml:"description"`
	Author      string      `yaml:"author"`
	Repository  string      `yaml:"repository"`
	Entry       string      `yaml:"entry,omitempty"`
	Tags        []string    `yaml:"tags,omitempty"`
	Commands    interface{} `yaml:"commands,omitempty"`
}

// createOrderedConfig creates an orderedConfig from InitOptions
func createOrderedConfig(opts InitOptions, existingCommands interface{}) orderedConfig {
	return orderedConfig{
		Name:        opts.Name,
		Version:     opts.Version,
		Description: opts.Description,
		Author:      opts.Author,
		Repository:  opts.Repository,
		Entry:       opts.Entry,
		Tags:        opts.Tags,
		Commands:    existingCommands,
	}
}

// InitProject creates a new ccmd project with the given options
func InitProject(opts InitOptions) error {
	// Create .claude/commands directory
	claudeDir := filepath.Join(opts.ProjectPath, ".claude", "commands")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return errors.FileError("create .claude directory", claudeDir, err)
	}

	// Create config
	config := &ProjectConfig{
		Name:        opts.Name,
		Version:     opts.Version,
		Description: opts.Description,
		Author:      opts.Author,
		Repository:  opts.Repository,
		Entry:       opts.Entry,
		Tags:        opts.Tags,
	}

	// Save config
	if err := SaveProjectConfig(opts.ProjectPath, config); err != nil {
		return err
	}

	return nil
}

// InitProjectWithCommands creates a project with existing commands preserved
func InitProjectWithCommands(opts InitOptions, existingCommands interface{}) error {
	// Create .claude/commands directory
	claudeDir := filepath.Join(opts.ProjectPath, ".claude", "commands")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return errors.FileError("create .claude directory", claudeDir, err)
	}

	// Create ordered structure to ensure commands field comes last
	config := createOrderedConfig(opts, existingCommands)

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return errors.FileError("marshal config", ConfigFileName, err)
	}

	// Write file
	configPath := filepath.Join(opts.ProjectPath, ConfigFileName)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return errors.FileError("write config", configPath, err)
	}

	return nil
}

// GenerateConfigPreview generates a YAML preview of the configuration
func GenerateConfigPreview(opts InitOptions, existingCommands interface{}) (string, error) {
	// Create ordered structure
	config := createOrderedConfig(opts, existingCommands)

	data, err := yaml.Marshal(&config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(data), nil
}

// ParseTags parses a comma-separated string of tags
func ParseTags(input string) []string {
	if input == "" {
		return []string{}
	}

	var tags []string
	for _, tag := range strings.Split(input, ",") {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			tags = append(tags, trimmed)
		}
	}

	return tags
}

// FormatTags formats a slice of tags as comma-separated string
func FormatTags(tags []string) string {
	return strings.Join(tags, ", ")
}

// GetInstallCommand generates the install command for the repository
func GetInstallCommand(repository string) string {
	if repository == "" {
		return "ccmd install github.com/your-username/your-repo"
	}

	// Extract GitHub path from full URL
	if strings.HasPrefix(repository, "https://github.com/") {
		repoPath := strings.TrimPrefix(repository, "https://github.com/")
		repoPath = strings.TrimSuffix(repoPath, ".git")
		return "ccmd install " + repoPath
	}

	return "ccmd install " + repository
}

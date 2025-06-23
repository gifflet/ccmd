/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/project"
)

// SearchOptions contains options for searching commands.
type SearchOptions struct {
	Keyword    string
	Tags       []string
	Author     string
	ShowAll    bool
	BaseDir    string
	FileSystem fs.FileSystem
}

// SearchResult represents a command found in the search.
type SearchResult struct {
	Name        string
	Version     string
	Description string
	Author      string
	Tags        []string
	Source      string
}

// Search searches for commands based on the provided options.
func Search(opts SearchOptions) ([]SearchResult, error) {
	if opts.FileSystem == nil {
		opts.FileSystem = fs.OS{}
	}

	if opts.BaseDir == "" {
		opts.BaseDir = "."
	}

	lockPath := filepath.Join(opts.BaseDir, "ccmd-lock.yaml")
	lockManager := project.NewLockManagerWithFS(lockPath, opts.FileSystem)
	if err := lockManager.Load(); err != nil {
		if os.IsNotExist(err) {
			return []SearchResult{}, nil
		}
		return nil, fmt.Errorf("failed to load lock file: %w", err)
	}

	cmds, err := lockManager.ListCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}

	var results []SearchResult
	for _, cmd := range cmds {
		// Load metadata for each command
		metadata := loadCommandMetadata(cmd.Name, opts.BaseDir, opts.FileSystem)
		if matches(cmd, metadata, opts) {
			results = append(results, toSearchResult(cmd, metadata))
		}
	}

	return results, nil
}

// matches checks if a command matches the search criteria.
func matches(cmd *project.CommandLockInfo, metadata *models.CommandMetadata, opts SearchOptions) bool {
	// If ShowAll is true and no other filters, return all
	if opts.ShowAll && opts.Keyword == "" && len(opts.Tags) == 0 && opts.Author == "" {
		return true
	}

	// If no criteria specified, don't match
	if opts.Keyword == "" && len(opts.Tags) == 0 && opts.Author == "" && !opts.ShowAll {
		return false
	}

	// Check each filter - all must match (AND logic)

	// Check keyword match if specified
	if opts.Keyword != "" {
		keyword := strings.ToLower(opts.Keyword)
		keywordMatch := false

		// Check name
		if strings.Contains(strings.ToLower(cmd.Name), keyword) {
			keywordMatch = true
		}

		// Check source
		if !keywordMatch {
			if strings.Contains(strings.ToLower(cmd.Source), keyword) {
				keywordMatch = true
			}
		}

		// Check description if metadata available
		if !keywordMatch && metadata != nil {
			if strings.Contains(strings.ToLower(metadata.Description), keyword) {
				keywordMatch = true
			}
		}

		// If keyword doesn't match, command doesn't match
		if !keywordMatch {
			return false
		}
	}

	// Check author match if specified
	if opts.Author != "" {
		if metadata == nil || !strings.Contains(strings.ToLower(metadata.Author), strings.ToLower(opts.Author)) {
			return false
		}
	}

	// Check tags match if specified
	if len(opts.Tags) > 0 {
		if metadata == nil || len(metadata.Tags) == 0 {
			return false
		}

		// Check if all requested tags are present
		for _, searchTag := range opts.Tags {
			found := false
			for _, cmdTag := range metadata.Tags {
				if strings.EqualFold(cmdTag, searchTag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// toSearchResult converts a Command to a SearchResult.
func toSearchResult(cmd *project.CommandLockInfo, metadata *models.CommandMetadata) SearchResult {
	result := SearchResult{
		Name:    cmd.Name,
		Version: cmd.Version,
		Source:  cmd.Source,
	}

	if metadata != nil {
		result.Description = metadata.Description
		result.Author = metadata.Author
		result.Tags = metadata.Tags
	}

	return result
}

// loadCommandMetadata loads metadata from a command's ccmd.yaml file
func loadCommandMetadata(name, baseDir string, filesystem fs.FileSystem) *models.CommandMetadata {
	commandDir := filepath.Join(baseDir, ".claude", "commands", name)
	metadataPath := filepath.Join(commandDir, "ccmd.yaml")

	data, err := filesystem.ReadFile(metadataPath)
	if err != nil {
		return nil
	}

	var metadata models.CommandMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil
	}

	return &metadata
}

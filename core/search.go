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
	"strings"
)

// SearchOptions contains options for searching commands
type SearchOptions struct {
	Keyword string
	Tags    []string
	Author  string
	ShowAll bool
}

// SearchResult represents a command found in the search
type SearchResult struct {
	Name        string
	Version     string
	Description string
	Author      string
	Tags        []string
	Repository  string
}

// Search searches for installed commands based on the provided options
func Search(opts SearchOptions) ([]SearchResult, error) {
	// Get all installed commands
	commands, err := List(ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, cmd := range commands {
		if matches(cmd, opts) {
			results = append(results, toSearchResult(cmd))
		}
	}

	return results, nil
}

// matches checks if a command matches the search criteria
func matches(cmd CommandDetail, opts SearchOptions) bool {
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

		// Check repository
		if !keywordMatch {
			if strings.Contains(strings.ToLower(cmd.Repository), keyword) {
				keywordMatch = true
			}
		}

		// Check description
		if !keywordMatch {
			if strings.Contains(strings.ToLower(cmd.Description), keyword) {
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
		if !strings.Contains(strings.ToLower(cmd.Author), strings.ToLower(opts.Author)) {
			return false
		}
	}

	// Check tags match if specified
	if len(opts.Tags) > 0 {
		if len(cmd.Tags) == 0 {
			return false
		}

		// Check if all requested tags are present
		for _, searchTag := range opts.Tags {
			found := false
			for _, cmdTag := range cmd.Tags {
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

// toSearchResult converts a CommandDetail to a SearchResult
func toSearchResult(cmd CommandDetail) SearchResult {
	return SearchResult{
		Name:        cmd.Name,
		Version:     cmd.Version,
		Description: cmd.Description,
		Author:      cmd.Author,
		Tags:        cmd.Tags,
		Repository:  cmd.Repository,
	}
}

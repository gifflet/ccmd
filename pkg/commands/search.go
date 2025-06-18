package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/lock"
	"github.com/gifflet/ccmd/internal/models"
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
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.BaseDir = filepath.Join(homeDir, ".claude")
	}

	lockManager := lock.NewManagerWithFS(opts.BaseDir, opts.FileSystem)
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
		if matches(cmd, opts) {
			results = append(results, toSearchResult(cmd))
		}
	}

	return results, nil
}

// matches checks if a command matches the search criteria.
func matches(cmd *models.Command, opts SearchOptions) bool {
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

		// Check description
		if !keywordMatch {
			if desc, ok := cmd.Metadata["description"]; ok {
				if strings.Contains(strings.ToLower(desc), keyword) {
					keywordMatch = true
				}
			}
		}

		// Check tags
		if !keywordMatch {
			if tagsStr, ok := cmd.Metadata["tags"]; ok {
				if strings.Contains(strings.ToLower(tagsStr), keyword) {
					keywordMatch = true
				}
			}
		}

		// If keyword doesn't match, command doesn't match
		if !keywordMatch {
			return false
		}
	}

	// Check author match if specified
	if opts.Author != "" {
		author, ok := cmd.Metadata["author"]
		if !ok || !strings.EqualFold(author, opts.Author) {
			return false
		}
	}

	// Check tags match if specified
	if len(opts.Tags) > 0 {
		tagsStr, ok := cmd.Metadata["tags"]
		if !ok {
			return false
		}

		cmdTags := parseTags(tagsStr)
		for _, searchTag := range opts.Tags {
			found := false
			for _, cmdTag := range cmdTags {
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

// parseTags parses a comma-separated string of tags.
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}

	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// toSearchResult converts a Command to a SearchResult.
func toSearchResult(cmd *models.Command) SearchResult {
	result := SearchResult{
		Name:    cmd.Name,
		Version: cmd.Version,
		Source:  cmd.Source,
	}

	// Extract metadata
	if desc, ok := cmd.Metadata["description"]; ok {
		result.Description = desc
	}

	if author, ok := cmd.Metadata["author"]; ok {
		result.Author = author
	}

	if tagsStr, ok := cmd.Metadata["tags"]; ok {
		result.Tags = parseTags(tagsStr)
	}

	return result
}

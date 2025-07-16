/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package search

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/core"
	"github.com/gifflet/ccmd/pkg/output"
)

// NewCommand creates a new search command.
func NewCommand() *cobra.Command {
	var (
		tags   []string
		author string
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "search [keyword]",
		Short: "Search for installed commands",
		Long: `Search for installed commands by keyword, tags, or author.
		
This command searches through locally installed commands.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var keyword string
			if len(args) > 0 {
				keyword = args[0]
			}
			return runSearch(keyword, tags, author, all)
		},
	}

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")
	cmd.Flags().StringVarP(&author, "author", "a", "", "Filter by author")
	cmd.Flags().BoolVar(&all, "all", false, "Show all commands (ignore keyword)")

	return cmd
}

func runSearch(keyword string, tags []string, author string, showAll bool) error {
	// Get search results
	opts := core.SearchOptions{
		Keyword: keyword,
		Tags:    tags,
		Author:  author,
		ShowAll: showAll,
	}

	results, err := core.Search(opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(results) == 0 {
		output.PrintInfof("No commands found matching your criteria.")
		if !showAll && keyword == "" && len(tags) == 0 && author == "" {
			output.PrintInfof("\nTip: Use 'ccmd search --all' to list all installed commands.")
		}
		return nil
	}

	output.PrintSuccessf("Found %d command(s):\n", len(results))

	for _, cmd := range results {
		displayCommand(&cmd)
	}

	if len(results) >= 10 {
		output.PrintInfof("\n💡 Note: Command registry search is coming soon for discovering more commands!")
	}

	return nil
}

func displayCommand(cmd *core.SearchResult) {
	// Display command name and version
	output.PrintInfof("📦 %s (v%s)", cmd.Name, cmd.Version)

	// Display description if available
	if cmd.Description != "" {
		output.PrintInfof("   %s", cmd.Description)
	}

	// Display author if available
	if cmd.Author != "" {
		output.PrintInfof("   Author: %s", cmd.Author)
	}

	// Display tags if available
	if len(cmd.Tags) > 0 {
		output.PrintInfof("   Tags: %s", strings.Join(cmd.Tags, ", "))
	}

	// Display repository
	if cmd.Repository != "" {
		output.PrintInfof("   Repository: %s", cmd.Repository)
	}

	output.PrintInfof("") // Empty line for spacing
}

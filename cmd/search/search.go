package search

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/internal/output"
	"github.com/gifflet/ccmd/pkg/commands"
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
		
This command searches through locally installed commands. In the future,
it will also search the command registry for available commands to install.`,
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
	opts := commands.SearchOptions{
		Keyword: keyword,
		Tags:    tags,
		Author:  author,
		ShowAll: showAll,
	}

	results, err := commands.Search(opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(results) == 0 {
		output.Info("No commands found matching your criteria.")
		if !showAll && keyword == "" && len(tags) == 0 && author == "" {
			output.Info("\nTip: Use 'ccmd search --all' to list all installed commands.")
		}
		output.Info("\n💡 Note: Command registry search is coming soon!")
		return nil
	}

	output.Success("Found %d command(s):\n", len(results))

	for _, cmd := range results {
		displayCommand(&cmd)
	}

	if len(results) >= 10 {
		output.Info("\n💡 Note: Command registry search is coming soon for discovering more commands!")
	}

	return nil
}

func displayCommand(cmd *commands.SearchResult) {
	// Display command name and version
	output.Info("📦 %s (v%s)", cmd.Name, cmd.Version)

	// Display description if available
	if cmd.Description != "" {
		output.Info("   %s", cmd.Description)
	}

	// Display author if available
	if cmd.Author != "" {
		output.Info("   Author: %s", cmd.Author)
	}

	// Display tags if available
	if len(cmd.Tags) > 0 {
		output.Info("   Tags: %s", strings.Join(cmd.Tags, ", "))
	}

	// Display source
	if cmd.Source != "" {
		output.Info("   Source: %s", cmd.Source)
	}

	output.Info("") // Empty line for spacing
}

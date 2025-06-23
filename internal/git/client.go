// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// Client provides git operations
type Client struct {
	workDir string
	logger  logger.Logger
}

// NewClient creates a new git client
func NewClient(workDir string) *Client {
	return &Client{
		workDir: workDir,
		logger:  logger.WithField("component", "git"),
	}
}

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL     string
	Target  string
	Branch  string
	Tag     string
	Shallow bool
	Depth   int
}

// Clone clones a repository with the given options
func (c *Client) Clone(opts CloneOptions) error {
	if opts.URL == "" {
		return errors.New(errors.CodeInvalidArgument, "repository URL is required")
	}
	if opts.Target == "" {
		return errors.New(errors.CodeInvalidArgument, "target directory is required")
	}

	args := []string{"clone"}

	if opts.Shallow || opts.Depth > 0 {
		depth := opts.Depth
		if depth == 0 {
			depth = 1
		}
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}

	if opts.Branch != "" && opts.Tag == "" {
		args = append(args, "--branch", opts.Branch)
	} else if opts.Tag != "" {
		args = append(args, "--branch", opts.Tag)
	}

	args = append(args, opts.URL, opts.Target)

	cmd := exec.Command("git", args...)
	// Use current directory for git clone, not the client's workDir
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.CodeGitClone, "git clone failed").
			WithDetail("repository", opts.URL).
			WithDetail("output", strings.TrimSpace(string(output)))
	}

	return nil
}

// CheckoutTag checks out a specific tag in the repository
func (c *Client) CheckoutTag(repoPath, tag string) error {
	cmd := exec.Command("git", "checkout", tag)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.CodeGitInvalidRepo, "git checkout failed").
			WithDetail("tag", tag).
			WithDetail("output", strings.TrimSpace(string(output)))
	}

	return nil
}

// GetTags returns all tags in the repository
func (c *Client) GetTags(repoPath string) ([]string, error) {
	// Convert to absolute path to avoid issues
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeFileIO, "failed to get absolute path").
			WithDetail("path", repoPath)
	}

	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = absPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeGitInvalidRepo, "git tag failed").
			WithDetail("repo", absPath).
			WithDetail("output", strings.TrimSpace(string(output)))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	tags := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			tags = append(tags, line)
		}
	}

	return tags, nil
}

// GetLatestTag returns the latest tag in the repository
func (c *Client) GetLatestTag(repoPath string) (string, error) {
	// Convert to absolute path to avoid issues
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", errors.Wrap(err, errors.CodeFileIO, "failed to get absolute path").
			WithDetail("path", repoPath)
	}

	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = absPath

	output, err := cmd.Output()
	if err != nil {
		// Try alternative method if describe fails
		tags, err := c.GetTags(absPath)
		if err != nil {
			return "", err
		}
		if len(tags) == 0 {
			return "", errors.New(errors.CodeGitNotFound, "no tags found in repository").
				WithDetail("repo", absPath)
		}
		return tags[len(tags)-1], nil
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentCommit returns the current commit hash
func (c *Client) GetCurrentCommit(repoPath string) (string, error) {
	// Convert to absolute path to avoid issues
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", errors.Wrap(err, errors.CodeFileIO, "failed to get absolute path").
			WithDetail("path", repoPath)
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = absPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, errors.CodeGitInvalidRepo, "git rev-parse failed").
			WithDetail("repo", absPath).
			WithDetail("output", strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
}

// IsGitRepository checks if a directory is a git repository
func (c *Client) IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ParseRepositoryURL parses a repository URL and extracts repo name
func ParseRepositoryURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", errors.New(errors.CodeInvalidArgument, "empty repository URL")
	}

	// Handle git@ URLs
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) < 2 {
			return "", errors.New(errors.CodeGitInvalidRepo, "invalid git URL format").
				WithDetail("url", url)
		}
		url = parts[1]
	}

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Extract repository name
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", errors.New(errors.CodeGitInvalidRepo, "invalid repository URL format").
			WithDetail("url", url)
	}

	return parts[len(parts)-1], nil
}

// ValidateRemoteRepository checks if a remote repository exists
func (c *Client) ValidateRemoteRepository(url string) error {
	cmd := exec.Command("git", "ls-remote", "--heads", url)
	cmd.Dir = c.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.CodeGitInvalidRepo, "repository validation failed").
			WithDetail("repository", url).
			WithDetail("output", strings.TrimSpace(string(output)))
	}

	return nil
}

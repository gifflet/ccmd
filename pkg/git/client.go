// Package git provides a simple wrapper around go-git for common Git operations.
package git

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Client provides Git operations using go-git.
type Client struct {
	auth transport.AuthMethod
}

// Options holds configuration for Git operations.
type Options struct {
	// Authentication options
	Username string
	Password string
	SSHKey   string
}

// New creates a new Git client with optional authentication.
func New(opts *Options) (*Client, error) {
	client := &Client{}

	if opts != nil {
		if opts.SSHKey != "" {
			auth, err := ssh.NewPublicKeysFromFile("git", opts.SSHKey, opts.Password)
			if err != nil {
				return nil, fmt.Errorf("failed to create SSH auth: %w", err)
			}
			client.auth = auth
		} else if opts.Username != "" && opts.Password != "" {
			client.auth = &http.BasicAuth{
				Username: opts.Username,
				Password: opts.Password,
			}
		}
	}

	return client, nil
}

// Clone clones a repository to the specified directory.
func (c *Client) Clone(repoURL, targetDir string) error {
	normalizedURL, err := normalizeURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:  normalizedURL,
		Auth: c.auth,
	}

	_, err = git.PlainClone(targetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Fetch fetches updates from the remote repository.
func (c *Client) Fetch(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	fetchOpts := &git.FetchOptions{
		Auth: c.auth,
		Tags: git.AllTags,
	}

	err = repo.Fetch(fetchOpts)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	return nil
}

// Checkout checks out a specific version, tag, or branch.
func (c *Client) Checkout(repoPath, ref string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to resolve as tag first
	tagRef, err := repo.Tag(ref)
	if err == nil {
		checkoutOpts := &git.CheckoutOptions{
			Hash: tagRef.Hash(),
		}
		if err := worktree.Checkout(checkoutOpts); err != nil {
			return fmt.Errorf("failed to checkout tag %s: %w", ref, err)
		}
		return nil
	}

	// Try as branch
	branchRef := plumbing.NewBranchReferenceName(ref)
	checkoutOpts := &git.CheckoutOptions{
		Branch: branchRef,
		Create: false,
	}

	if err := worktree.Checkout(checkoutOpts); err != nil {
		// Try as commit hash
		hash := plumbing.NewHash(ref)
		checkoutOpts = &git.CheckoutOptions{
			Hash: hash,
		}
		if err := worktree.Checkout(checkoutOpts); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", ref, err)
		}
	}

	return nil
}

// ListTags returns all tags in the repository.
func (c *Client) ListTags(repoPath string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var tagList []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		tagList = append(tagList, tagName)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tagList, nil
}

// GetRemoteURL returns the URL of the origin remote.
func (c *Client) GetRemoteURL(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get origin remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", errors.New("no URL configured for origin remote")
	}

	return urls[0], nil
}

// normalizeURL validates and normalizes a Git repository URL.
func normalizeURL(repoURL string) (string, error) {
	if repoURL == "" {
		return "", errors.New("repository URL cannot be empty")
	}

	// Check if it's a local file path
	if filepath.IsAbs(repoURL) || strings.HasPrefix(repoURL, "./") || strings.HasPrefix(repoURL, "../") {
		// Local path is valid
		return repoURL, nil
	}

	// Handle GitHub shorthand (e.g., "user/repo")
	if !strings.Contains(repoURL, "://") && !strings.HasPrefix(repoURL, "git@") {
		if strings.Count(repoURL, "/") == 1 && !strings.Contains(repoURL, "..") {
			repoURL = "https://github.com/" + repoURL
		}
	}

	// Add .git suffix if missing for common Git hosts
	if strings.Contains(repoURL, "github.com") || strings.Contains(repoURL, "gitlab.com") || strings.Contains(repoURL, "bitbucket.org") {
		if !strings.HasSuffix(repoURL, ".git") && !strings.Contains(repoURL, ".git/") {
			repoURL = repoURL + ".git"
		}
	}

	// Validate URL format
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		_, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}
	} else if !strings.HasPrefix(repoURL, "git@") && !strings.HasPrefix(repoURL, "ssh://") && !strings.HasPrefix(repoURL, "file://") {
		// If it's not a recognized scheme and not a local path, it's invalid
		if !filepath.IsAbs(repoURL) {
			return "", errors.New("unsupported URL scheme")
		}
	}

	return repoURL, nil
}

// SetRemote adds or updates a remote in the repository.
func (c *Client) SetRemote(repoPath, name, url string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	normalizedURL, err := normalizeURL(url)
	if err != nil {
		return fmt.Errorf("invalid remote URL: %w", err)
	}

	// Try to get existing remote
	remote, err := repo.Remote(name)
	if err == nil {
		// Update existing remote
		cfg := remote.Config()
		cfg.URLs = []string{normalizedURL}
		err = repo.DeleteRemote(name)
		if err != nil {
			return fmt.Errorf("failed to delete existing remote: %w", err)
		}
	}

	// Create remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{normalizedURL},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}

	return nil
}

// IsValidRepository checks if a directory is a valid Git repository.
func IsValidRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

// GetRepositoryName extracts the repository name from a Git URL.
func GetRepositoryName(repoURL string) string {
	// Remove .git suffix
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Extract the last part of the URL
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	// If URL parsing fails, try to extract from the path
	return filepath.Base(repoURL)
}

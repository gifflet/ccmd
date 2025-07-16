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
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var (
	gitPath     string
	gitPathOnce sync.Once
	gitPathErr  error
)

// getGitPath returns the absolute path to git executable after validating it's in a secure location
func getGitPath() (string, error) {
	gitPathOnce.Do(func() {
		path, err := exec.LookPath("git")
		if err != nil {
			gitPathErr = fmt.Errorf("git not found in PATH: %w", err)
			return
		}

		trustedPaths := []string{
			"/usr/bin/",
			"/usr/local/bin/",
			"/opt/homebrew/bin/",
			"/opt/local/bin/",
			"C:\\Program Files\\Git\\",
			"C:\\Program Files (x86)\\Git\\",
		}

		trusted := false
		for _, tp := range trustedPaths {
			if strings.HasPrefix(path, tp) {
				trusted = true
				break
			}
		}

		if !trusted {
			gitPathErr = fmt.Errorf("git found in untrusted location: %s", path)
			return
		}

		gitPath = path
	})

	return gitPath, gitPathErr
}

// isCommitHash checks if a string looks like a git commit SHA-1 hash
func isCommitHash(s string) bool {
	// Git commit hashes are hexadecimal strings of 7-40 characters
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	// Check if all characters are valid hexadecimal
	matched, err := regexp.MatchString("^[a-f0-9]+$", s)
	if err != nil {
		return false
	}
	return matched
}

// gitClone clones a repository to the specified destination
func gitClone(repo, dest, version string) error {
	git, err := getGitPath()
	if err != nil {
		return err
	}

	if version != "" && isCommitHash(version) {
		// For commit hashes, we need to clone first then checkout
		// Clone without depth limit to access all commits
		cmd := exec.Command(git, "clone", repo, dest)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
		}

		// Checkout the specific commit
		checkoutCmd := exec.Command(git, "-C", dest, "checkout", version)
		checkoutOutput, checkoutErr := checkoutCmd.CombinedOutput()
		if checkoutErr != nil {
			return fmt.Errorf("git checkout failed: %w\nOutput: %s", checkoutErr, string(checkoutOutput))
		}

		return nil
	}

	// For branches and tags, use shallow clone
	args := []string{"clone", "--depth", "1"}

	if version != "" {
		args = append(args, "--branch", version)
	}

	args = append(args, repo, dest)

	cmd := exec.Command(git, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// gitGetCurrentCommit returns the current commit hash of a repository
func gitGetCurrentCommit(repoPath string) (string, error) {
	git, err := getGitPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(git, "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// gitGetRefCommit returns the commit hash for a specific ref (tag, branch or commit)
func gitGetRefCommit(repoPath, ref string) (string, error) {
	git, err := getGitPath()
	if err != nil {
		return "", err
	}

	// Get local ref commit
	cmd := exec.Command(git, "-C", repoPath, "rev-list", "-n", "1", ref)
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get ref commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// gitGetRemoteRefCommit returns the commit hash for a remote ref (tag or branch)
func gitGetRemoteRefCommit(repoPath, ref string) (string, error) {
	git, err := getGitPath()
	if err != nil {
		return "", err
	}

	// Try as tag first
	cmd := exec.Command(git, "-C", repoPath, "ls-remote", "origin", fmt.Sprintf("refs/tags/%s", ref))
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		parts := strings.Fields(string(output))
		if len(parts) > 0 {
			return parts[0], nil
		}
	}

	// Try as branch
	cmd = exec.Command(git, "-C", repoPath, "ls-remote", "origin", fmt.Sprintf("refs/heads/%s", ref))
	output, err = cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get remote ref commit: %w", err)
	}

	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("ref %s not found in remote", ref)
}

// gitGetDefaultBranch returns the default branch name of a repository
func gitGetDefaultBranch(repoPath string) (string, error) {
	git, err := getGitPath()
	if err != nil {
		return "", err
	}

	// Try to get the default branch from remote
	cmd := exec.Command(git, "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()

	if err == nil {
		// Output format: refs/remotes/origin/main
		branch := strings.TrimSpace(string(output))
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		return branch, nil
	}

	// Fallback: try common default branch names
	for _, branch := range []string{"main", "master"} {
		ref := fmt.Sprintf("refs/remotes/origin/%s", branch)
		checkCmd := exec.Command(git, "-C", repoPath, "show-ref", "--verify", ref)
		if err := checkCmd.Run(); err == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

// ExtractRepoPath extracts the owner/repo path from a Git URL
func ExtractRepoPath(gitURL string) string {
	// Remove protocol
	if idx := strings.Index(gitURL, "://"); idx != -1 {
		gitURL = gitURL[idx+3:]
	}

	// Remove git@ prefix
	gitURL = strings.TrimPrefix(gitURL, "git@")

	// Replace : with / for SSH URLs
	gitURL = strings.Replace(gitURL, ":", "/", 1)

	// Remove .git suffix
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Extract path after domain
	parts := strings.Split(gitURL, "/")
	if len(parts) >= 3 {
		// Return owner/repo
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return gitURL
}

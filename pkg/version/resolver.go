// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
// Package version provides version resolution functionality for commands.
package version

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Resolver handles version resolution for commands.
type Resolver struct {
	// getTags is a function that returns available tags for a repository
	getTags func(repoPath string) ([]string, error)
}

// NewResolver creates a new version resolver.
func NewResolver(getTagsFunc func(string) ([]string, error)) *Resolver {
	return &Resolver{
		getTags: getTagsFunc,
	}
}

// ResolveVersion resolves a version specification to a concrete Git reference.
// It handles the following formats:
// - "latest" - resolves to the latest semantic version tag
// - Semantic version constraint (e.g., "^1.0.0", "~2.1.0", ">=1.0.0")
// - Exact tag name (e.g., "v1.2.3" or "1.2.3")
// - Branch name (e.g., "main", "feature/xyz")
// - Commit hash (e.g., "abc123def")
func (r *Resolver) ResolveVersion(repoPath, version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version cannot be empty")
	}

	// Handle "latest" keyword
	if version == "latest" {
		return r.resolveLatest(repoPath)
	}

	// Get all available tags
	tags, err := r.getTags(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	// Check if it's a semantic version constraint
	if isSemverConstraint(version) {
		return r.resolveSemverConstraint(version, tags)
	}

	// Check if it's an exact tag match
	if resolved := r.findExactTag(version, tags); resolved != "" {
		return resolved, nil
	}

	// If not a tag, it could be a branch or commit hash
	// Return as-is and let Git handle the validation
	return version, nil
}

// resolveLatest finds the latest semantic version tag.
func (r *Resolver) resolveLatest(repoPath string) (string, error) {
	tags, err := r.getTags(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in repository")
	}

	// Extract semantic versions
	var versions []*semver.Version
	tagMap := make(map[string]string) // maps normalized version to original tag

	for _, tag := range tags {
		v, err := parseSemverTag(tag)
		if err == nil {
			versions = append(versions, v)
			tagMap[v.String()] = tag
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no semantic version tags found")
	}

	// Sort versions in descending order
	sort.Sort(sort.Reverse(semver.Collection(versions)))

	// Return the original tag for the latest version
	return tagMap[versions[0].String()], nil
}

// resolveSemverConstraint resolves a semantic version constraint to a specific tag.
func (r *Resolver) resolveSemverConstraint(constraint string, tags []string) (string, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return "", fmt.Errorf("invalid semantic version constraint: %w", err)
	}

	// Extract semantic versions from tags
	type versionTag struct {
		version *semver.Version
		tag     string
	}
	var versions []versionTag

	for _, tag := range tags {
		v, err := parseSemverTag(tag)
		if err == nil {
			versions = append(versions, versionTag{version: v, tag: tag})
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no semantic version tags found")
	}

	// Sort versions in descending order
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].version.GreaterThan(versions[j].version)
	})

	// Find the first version that satisfies the constraint
	for _, vt := range versions {
		if c.Check(vt.version) {
			return vt.tag, nil
		}
	}

	return "", fmt.Errorf("no version found matching constraint: %s", constraint)
}

// findExactTag checks if the version matches any tag exactly.
func (r *Resolver) findExactTag(version string, tags []string) string {
	// Direct match
	for _, tag := range tags {
		if tag == version {
			return tag
		}
	}

	// Try with/without 'v' prefix
	if strings.HasPrefix(version, "v") {
		withoutV := strings.TrimPrefix(version, "v")
		for _, tag := range tags {
			if tag == withoutV {
				return tag
			}
		}
	} else {
		withV := "v" + version
		for _, tag := range tags {
			if tag == withV {
				return tag
			}
		}
	}

	return ""
}

// parseSemverTag attempts to parse a tag as a semantic version.
// It handles tags with or without 'v' prefix.
func parseSemverTag(tag string) (*semver.Version, error) {
	// Try parsing as-is
	v, err := semver.NewVersion(tag)
	if err == nil {
		return v, nil
	}

	// Try without 'v' prefix
	if strings.HasPrefix(tag, "v") {
		return semver.NewVersion(strings.TrimPrefix(tag, "v"))
	}

	// Try with 'v' prefix
	return semver.NewVersion("v" + tag)
}

// isSemverConstraint checks if a string looks like a semantic version constraint.
func isSemverConstraint(s string) bool {
	// Common constraint prefixes
	constraintPrefixes := []string{"^", "~", ">=", "<=", "="}
	for _, prefix := range constraintPrefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	// Check for comparison operators (but not as part of arrows like <->)
	if strings.HasPrefix(s, ">") && !strings.HasPrefix(s, ">=") {
		return true
	}
	if strings.HasPrefix(s, "<") && !strings.HasPrefix(s, "<=") {
		return true
	}

	// Check for range constraints
	if strings.Contains(s, " - ") || strings.Contains(s, " || ") {
		return true
	}

	// Check if it's a wildcard version
	// Must look like a version with wildcards, not just any string with these characters
	parts := strings.Split(s, ".")
	if len(parts) >= 2 {
		for _, part := range parts {
			if part == "*" || part == "x" || part == "X" {
				return true
			}
		}
	}

	return false
}

// ValidateReference checks if a Git reference exists in the repository.
// This is useful for validating branches and commit hashes.
func (r *Resolver) ValidateReference(_, _ string) error {
	// This would typically use git commands to validate
	// For now, we'll leave it as a placeholder
	// The actual validation will happen during checkout
	return nil
}

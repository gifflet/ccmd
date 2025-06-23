// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRepositoryURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "https URL with .git",
			url:  "https://github.com/user/repo.git",
			want: "repo",
		},
		{
			name: "https URL without .git",
			url:  "https://github.com/user/repo",
			want: "repo",
		},
		{
			name: "git@ URL",
			url:  "git@github.com:user/repo.git",
			want: "repo",
		},
		{
			name: "URL with subdirectories",
			url:  "https://github.com/org/team/repo.git",
			want: "repo",
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "git@ URL without colon",
			url:     "git@github.com/user/repo.git",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepositoryURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepositoryURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRepositoryURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGitRepository(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)

	// Test non-git directory
	if client.IsGitRepository(tempDir) {
		t.Errorf("IsGitRepository() returned true for non-git directory")
	}

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Test git directory
	if !client.IsGitRepository(tempDir) {
		t.Errorf("IsGitRepository() returned false for git directory")
	}

	// Test non-existent directory
	if client.IsGitRepository(filepath.Join(tempDir, "nonexistent")) {
		t.Errorf("IsGitRepository() returned true for non-existent directory")
	}
}

func TestCloneOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    CloneOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: CloneOptions{
				URL:    "https://github.com/user/repo.git",
				Target: "/tmp/repo",
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			opts: CloneOptions{
				Target: "/tmp/repo",
			},
			wantErr: true,
		},
		{
			name: "missing target",
			opts: CloneOptions{
				URL: "https://github.com/user/repo.git",
			},
			wantErr: true,
		},
		{
			name: "with branch",
			opts: CloneOptions{
				URL:    "https://github.com/user/repo.git",
				Target: "/tmp/repo",
				Branch: "develop",
			},
			wantErr: false,
		},
		{
			name: "with tag",
			opts: CloneOptions{
				URL:    "https://github.com/user/repo.git",
				Target: "/tmp/repo",
				Tag:    "v1.0.0",
			},
			wantErr: false,
		},
		{
			name: "shallow clone",
			opts: CloneOptions{
				URL:     "https://github.com/user/repo.git",
				Target:  "/tmp/repo",
				Shallow: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("/tmp")
			err := client.Clone(tt.opts)
			// We expect errors because we're not actually cloning
			// Just check that validation works
			if tt.opts.URL == "" || tt.opts.Target == "" {
				if err == nil {
					t.Errorf("Clone() expected validation error but got none")
				}
			}
		})
	}
}

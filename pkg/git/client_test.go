// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "empty options",
			opts:    &Options{},
			wantErr: false,
		},
		{
			name: "basic auth",
			opts: &Options{
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "invalid SSH key",
			opts: &Options{
				SSHKey: "/nonexistent/key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:  "GitHub shorthand",
			input: "user/repo",
			want:  "https://github.com/user/repo.git",
		},
		{
			name:  "HTTPS URL without .git",
			input: "https://github.com/user/repo",
			want:  "https://github.com/user/repo.git",
		},
		{
			name:  "HTTPS URL with .git",
			input: "https://github.com/user/repo.git",
			want:  "https://github.com/user/repo.git",
		},
		{
			name:  "SSH URL",
			input: "git@github.com:user/repo.git",
			want:  "git@github.com:user/repo.git",
		},
		{
			name:  "GitLab URL",
			input: "https://gitlab.com/user/repo",
			want:  "https://gitlab.com/user/repo.git",
		},
		{
			name:  "Bitbucket URL",
			input: "https://bitbucket.org/user/repo",
			want:  "https://bitbucket.org/user/repo.git",
		},
		{
			name:    "invalid URL scheme",
			input:   "ftp://example.com/repo",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://[invalid",
			wantErr: true,
		},
		{
			name:  "absolute path",
			input: "/path/to/repo",
			want:  "/path/to/repo",
		},
		{
			name:  "relative path with ./",
			input: "./repo",
			want:  "./repo",
		},
		{
			name:  "relative path with ../",
			input: "../repo",
			want:  "../repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepositoryName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "HTTPS URL with .git",
			input: "https://github.com/user/repo.git",
			want:  "repo",
		},
		{
			name:  "HTTPS URL without .git",
			input: "https://github.com/user/repo",
			want:  "repo",
		},
		{
			name:  "SSH URL",
			input: "git@github.com:user/repo.git",
			want:  "repo",
		},
		{
			name:  "simple name",
			input: "myrepo",
			want:  "myrepo",
		},
		{
			name:  "path",
			input: "/path/to/repo",
			want:  "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRepositoryName(tt.input); got != tt.want {
				t.Errorf("GetRepositoryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientOperations(t *testing.T) {
	// Create temporary directory for test repository
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo")
	cloneDir := filepath.Join(tempDir, "clone")

	// Initialize a test repository
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatalf("failed to init test repo: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(repoDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add and commit
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := worktree.Add("test.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	commit, err := worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create a tag
	if _, err := repo.CreateTag("v1.0.0", commit, nil); err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	// Test client operations
	client, err := New(nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("Clone", func(t *testing.T) {
		err := client.Clone(repoDir, cloneDir)
		if err != nil {
			t.Errorf("Clone() error = %v", err)
		}

		// Verify clone exists
		if !IsValidRepository(cloneDir) {
			t.Error("cloned directory is not a valid repository")
		}
	})

	t.Run("ListTags", func(t *testing.T) {
		tags, err := client.ListTags(cloneDir)
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
			return
		}

		if len(tags) != 1 || tags[0] != "v1.0.0" {
			t.Errorf("ListTags() = %v, want [v1.0.0]", tags)
		}
	})

	t.Run("GetRemoteURL", func(t *testing.T) {
		// Set remote for clone
		err := client.SetRemote(cloneDir, "origin", repoDir)
		if err != nil {
			t.Errorf("SetRemote() error = %v", err)
			return
		}

		url, err := client.GetRemoteURL(cloneDir)
		if err != nil {
			t.Errorf("GetRemoteURL() error = %v", err)
			return
		}

		if url != repoDir {
			t.Errorf("GetRemoteURL() = %v, want %v", url, repoDir)
		}
	})

	t.Run("Fetch", func(t *testing.T) {
		// Create another commit in original repo
		testFile2 := filepath.Join(repoDir, "test2.txt")
		if err := os.WriteFile(testFile2, []byte("test content 2"), 0o644); err != nil {
			t.Fatalf("failed to create test file 2: %v", err)
		}

		if _, err := worktree.Add("test2.txt"); err != nil {
			t.Fatalf("failed to add file 2: %v", err)
		}

		if _, err := worktree.Commit("second commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
			},
		}); err != nil {
			t.Fatalf("failed to commit 2: %v", err)
		}

		// Fetch updates
		err := client.Fetch(cloneDir)
		if err != nil {
			t.Errorf("Fetch() error = %v", err)
		}
	})

	t.Run("Checkout", func(t *testing.T) {
		// Checkout tag
		err := client.Checkout(cloneDir, "v1.0.0")
		if err != nil {
			t.Errorf("Checkout() tag error = %v", err)
		}

		// Verify file from second commit doesn't exist
		if _, err := os.Stat(filepath.Join(cloneDir, "test2.txt")); !os.IsNotExist(err) {
			t.Error("test2.txt should not exist after checking out v1.0.0")
		}
	})
}

func TestIsValidRepository(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	nonRepoDir := filepath.Join(tempDir, "not-repo")

	// Create a valid repository
	if _, err := git.PlainInit(repoDir, false); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Create a non-repository directory
	if err := os.MkdirAll(nonRepoDir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid repository",
			path: repoDir,
			want: true,
		},
		{
			name: "non-repository directory",
			path: nonRepoDir,
			want: false,
		},
		{
			name: "non-existent directory",
			path: filepath.Join(tempDir, "nonexistent"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRepository(tt.path); got != tt.want {
				t.Errorf("IsValidRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gifflet/ccmd/internal/fs"
)

func TestInstall(t *testing.T) {
	// Skip install tests as they require actual git operations
	// These would be better as integration tests
	t.Skip("Skipping install tests that require git operations")
}

func TestInstall_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        InstallOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "missing repository URL",
			opts: InstallOptions{
				Repository: "",
			},
			wantErr:     true,
			errContains: "repository URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation
			err := Install(tt.opts)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Install() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestParseRepositorySpec(t *testing.T) {
	tests := []struct {
		name           string
		spec           string
		wantRepository string
		wantVersion    string
	}{
		{
			name:           "repository only",
			spec:           "github.com/user/repo",
			wantRepository: "github.com/user/repo",
			wantVersion:    "",
		},
		{
			name:           "repository with version",
			spec:           "github.com/user/repo@v1.0.0",
			wantRepository: "github.com/user/repo",
			wantVersion:    "v1.0.0",
		},
		{
			name:           "repository with commit hash",
			spec:           "github.com/user/repo@abc123",
			wantRepository: "github.com/user/repo",
			wantVersion:    "abc123",
		},
		{
			name:           "repository with @ in name",
			spec:           "github.com/user@org/repo@v1.0.0",
			wantRepository: "github.com/user@org/repo",
			wantVersion:    "v1.0.0",
		},
		{
			name:           "ssh url without version",
			spec:           "git@github.com:user/repo.git",
			wantRepository: "git@github.com:user/repo.git",
			wantVersion:    "",
		},
		{
			name:           "ssh url with version",
			spec:           "git@github.com:user/repo.git@v2.0.0",
			wantRepository: "git@github.com:user/repo.git",
			wantVersion:    "v2.0.0",
		},
		{
			name:           "tag without v prefix",
			spec:           "github.com/user/repo@modular-architecture-improvements",
			wantRepository: "github.com/user/repo",
			wantVersion:    "modular-architecture-improvements",
		},
		{
			name:           "branch name with slashes",
			spec:           "github.com/user/repo@feature/new-feature",
			wantRepository: "github.com/user/repo",
			wantVersion:    "feature/new-feature",
		},
		{
			name:           "release tag without v",
			spec:           "github.com/user/repo@release-2.0",
			wantRepository: "github.com/user/repo",
			wantVersion:    "release-2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, version := ParseRepositorySpec(tt.spec)
			if repo != tt.wantRepository {
				t.Errorf("ParseRepositorySpec() repository = %v, want %v", repo, tt.wantRepository)
			}
			if version != tt.wantVersion {
				t.Errorf("ParseRepositorySpec() version = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestCopyDirectory(t *testing.T) {
	mockFS := fs.NewMemoryFileSystem()

	// Create source directory structure
	srcDir := "/src"
	_ = mockFS.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755)
	_ = mockFS.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0o644)
	_ = mockFS.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0o644)

	// Create .git directory that should be skipped
	_ = mockFS.MkdirAll(filepath.Join(srcDir, ".git"), 0o755)
	_ = mockFS.WriteFile(filepath.Join(srcDir, ".git", "config"), []byte("git config"), 0o644)

	// Copy directory
	dstDir := "/dst"
	err := copyDirectory(mockFS, srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyDirectory() error = %v", err)
	}

	// Verify files were copied
	tests := []struct {
		path        string
		expected    string
		shouldExist bool
	}{
		{filepath.Join(dstDir, "file1.txt"), "content1", true},
		{filepath.Join(dstDir, "subdir", "file2.txt"), "content2", true},
		{filepath.Join(dstDir, ".git", "config"), "", false}, // .git should be skipped
	}

	for _, tt := range tests {
		data, err := mockFS.ReadFile(tt.path)
		if tt.shouldExist {
			if err != nil {
				t.Errorf("expected file %s to exist, but got error: %v", tt.path, err)
				continue
			}
			if string(data) != tt.expected {
				t.Errorf("file %s content = %q, want %q", tt.path, string(data), tt.expected)
			}
		} else {
			if err == nil {
				t.Errorf("expected file %s to not exist, but it does", tt.path)
			}
		}
	}
}

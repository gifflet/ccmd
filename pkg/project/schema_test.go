package project

import (
	"os"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with single command",
			yaml: `commands:
  - repo: owner/repo
    version: v1.0.0`,
			wantErr: false,
		},
		{
			name: "valid config with multiple commands",
			yaml: `commands:
  - repo: owner/repo1
    version: v1.0.0
  - repo: owner/repo2
    version: latest
  - repo: owner/repo3`,
			wantErr: false,
		},
		{
			name: "valid config with branch version",
			yaml: `commands:
  - repo: owner/repo
    version: main`,
			wantErr: false,
		},
		{
			name:    "empty config",
			yaml:    ``,
			wantErr: true,
			errMsg:  "no commands defined",
		},
		{
			name:    "empty commands list",
			yaml:    `commands: []`,
			wantErr: true,
			errMsg:  "no commands defined",
		},
		{
			name: "missing repo",
			yaml: `commands:
  - version: v1.0.0`,
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name: "invalid repo format - no slash",
			yaml: `commands:
  - repo: invalidrepo
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "invalid repo format",
		},
		{
			name: "invalid repo format - multiple slashes",
			yaml: `commands:
  - repo: owner/repo/extra
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "invalid repo format",
		},
		{
			name: "invalid repo format - empty owner",
			yaml: `commands:
  - repo: /repo
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "owner and repo name cannot be empty",
		},
		{
			name: "invalid repo format - empty repo name",
			yaml: `commands:
  - repo: owner/
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "owner and repo name cannot be empty",
		},
		{
			name: "invalid owner name - starts with dash",
			yaml: `commands:
  - repo: -owner/repo
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "invalid owner name",
		},
		{
			name: "invalid repo name - invalid chars",
			yaml: `commands:
  - repo: owner/repo@123
    version: v1.0.0`,
			wantErr: true,
			errMsg:  "invalid repo name",
		},
		{
			name: "unknown field",
			yaml: `commands:
  - repo: owner/repo
    version: v1.0.0
    unknown: field`,
			wantErr: true,
			errMsg:  "field unknown not found",
		},
		{
			name: "invalid version format",
			yaml: `commands:
  - repo: owner/repo
    version: ..invalid`,
			wantErr: true,
			errMsg:  "invalid version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(strings.NewReader(tt.yaml))

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseConfig() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseConfig() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestCommand_ParseOwnerRepo(t *testing.T) {
	tests := []struct {
		name      string
		repo      string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid repo",
			repo:      "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "valid repo with dashes",
			repo:      "my-owner/my-repo",
			wantOwner: "my-owner",
			wantRepo:  "my-repo",
			wantErr:   false,
		},
		{
			name:      "valid repo with underscores",
			repo:      "my_owner/my_repo",
			wantOwner: "my_owner",
			wantRepo:  "my_repo",
			wantErr:   false,
		},
		{
			name:    "invalid - no slash",
			repo:    "invalidrepo",
			wantErr: true,
		},
		{
			name:    "invalid - multiple slashes",
			repo:    "owner/repo/extra",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			repo:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Repo: tt.repo}
			owner, repo, err := cmd.ParseOwnerRepo()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOwnerRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseOwnerRepo() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseOwnerRepo() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestCommand_IsSemanticVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "valid semantic version",
			version: "v1.0.0",
			want:    true,
		},
		{
			name:    "valid semantic version without v",
			version: "1.0.0",
			want:    true,
		},
		{
			name:    "valid semantic version with prerelease",
			version: "v1.0.0-alpha.1",
			want:    true,
		},
		{
			name:    "valid semantic version with metadata",
			version: "v1.0.0+build.123",
			want:    true,
		},
		{
			name:    "latest is not semantic",
			version: "latest",
			want:    false,
		},
		{
			name:    "branch name is not semantic",
			version: "main",
			want:    false,
		},
		{
			name:    "tag name is not semantic",
			version: "release-1.0",
			want:    false,
		},
		{
			name:    "empty version is not semantic",
			version: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Version: tt.version}
			if got := cmd.IsSemanticVersion(); got != tt.want {
				t.Errorf("IsSemanticVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "empty version is valid",
			version: "",
			wantErr: false,
		},
		{
			name:    "latest is valid",
			version: "latest",
			wantErr: false,
		},
		{
			name:    "semantic version is valid",
			version: "v1.2.3",
			wantErr: false,
		},
		{
			name:    "branch name is valid",
			version: "main",
			wantErr: false,
		},
		{
			name:    "tag name is valid",
			version: "release-1.0",
			wantErr: false,
		},
		{
			name:    "version with dots at start",
			version: ".invalid",
			wantErr: true,
		},
		{
			name:    "version with dots at end",
			version: "invalid.",
			wantErr: true,
		},
		{
			name:    "version with double dots",
			version: "invalid..version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		_, err := LoadConfig("/non/existent/file.yaml")
		if err == nil {
			t.Error("LoadConfig() should fail for non-existent file")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		// Create temporary test file
		tmpfile, err := os.CreateTemp("", "test_ccmd_*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		content := `commands:
  - repo: test/repo
    version: v1.0.0`
		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		config, err := LoadConfig(tmpfile.Name())
		if err != nil {
			t.Errorf("LoadConfig() error = %v, want nil", err)
		}
		if config == nil {
			t.Error("LoadConfig() returned nil config")
		}
		if len(config.Commands) != 1 {
			t.Errorf("LoadConfig() got %d commands, want 1", len(config.Commands))
		}
	})
}

func TestParseConfigWithInvalidYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid YAML syntax",
			yaml:    `commands: [invalid yaml`,
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
		{
			name:    "wrong type for commands",
			yaml:    `commands: "not a list"`,
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(strings.NewReader(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ParseConfig() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

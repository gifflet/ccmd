package installer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/logger"
	"github.com/gifflet/ccmd/pkg/project"
)

func TestParseRepositorySpec(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		wantRepo string
		wantVer  string
	}{
		{
			name:     "simple URL",
			spec:     "https://github.com/user/repo.git",
			wantRepo: "https://github.com/user/repo.git",
			wantVer:  "",
		},
		{
			name:     "URL with version",
			spec:     "https://github.com/user/repo.git@v1.0.0",
			wantRepo: "https://github.com/user/repo.git",
			wantVer:  "v1.0.0",
		},
		{
			name:     "shorthand with version",
			spec:     "user/repo@v2.0.0",
			wantRepo: "user/repo",
			wantVer:  "v2.0.0",
		},
		{
			name:     "SSH URL",
			spec:     "git@github.com:user/repo.git",
			wantRepo: "git@github.com:user/repo.git",
			wantVer:  "",
		},
		{
			name:     "SSH URL should not split on @",
			spec:     "git@github.com:user/repo.git@v1.0.0",
			wantRepo: "git@github.com:user/repo.git",
			wantVer:  "v1.0.0",
		},
		{
			name:     "complex version",
			spec:     "https://example.com/repo.git@feature/new-thing",
			wantRepo: "https://example.com/repo.git",
			wantVer:  "feature/new-thing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ver := ParseRepositorySpec(tt.spec)
			assert.Equal(t, tt.wantRepo, repo)
			assert.Equal(t, tt.wantVer, ver)
		})
	}
}

func TestNormalizeRepositoryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub shorthand",
			input:    "user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "full HTTPS URL",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL without protocol",
			input:    "github.com/user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL without .git suffix",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "SSH URL unchanged",
			input:    "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "GitLab URL",
			input:    "gitlab.com/user/repo",
			expected: "https://gitlab.com/user/repo.git",
		},
		{
			name:     "Bitbucket URL",
			input:    "bitbucket.org/user/repo",
			expected: "https://bitbucket.org/user/repo.git",
		},
		{
			name:     "custom domain not modified",
			input:    "git.company.com/user/repo",
			expected: "git.company.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRepositoryURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRepoPath(t *testing.T) {
	tests := []struct {
		name     string
		gitURL   string
		expected string
	}{
		{
			name:     "HTTPS URL",
			gitURL:   "https://github.com/user/repo.git",
			expected: "user/repo",
		},
		{
			name:     "HTTP URL",
			gitURL:   "http://github.com/user/repo.git",
			expected: "user/repo",
		},
		{
			name:     "SSH URL",
			gitURL:   "git@github.com:user/repo.git",
			expected: "user/repo",
		},
		{
			name:     "git protocol",
			gitURL:   "git://github.com/user/repo.git",
			expected: "user/repo",
		},
		{
			name:     "URL without .git",
			gitURL:   "https://github.com/user/repo",
			expected: "user/repo",
		},
		{
			name:     "nested path",
			gitURL:   "https://github.com/org/team/repo.git",
			expected: "org/team",
		},
		{
			name:     "invalid URL",
			gitURL:   "not-a-url",
			expected: "",
		},
		{
			name:     "domain only",
			gitURL:   "https://github.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRepoPath(tt.gitURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstallFromConfig(t *testing.T) {
	// Create test filesystem
	memFS := fs.NewMemFS()

	// Create project directory
	projectDir := "/test/project"
	require.NoError(t, memFS.MkdirAll(projectDir, 0o755))

	// Create ccmd.yaml with test commands
	config := &project.Config{
		Commands: []project.ConfigCommand{
			{Repo: "test/cmd1", Version: "v1.0.0"},
			{Repo: "test/cmd2", Version: ""},
		},
	}

	// Marshal config to YAML
	configData, err := yaml.Marshal(config)
	require.NoError(t, err)
	require.NoError(t, memFS.WriteFile(filepath.Join(projectDir, "ccmd.yaml"), configData, 0o644))

	// Mock git client for testing
	gitClient := &mockGitClient{
		cloneFunc: func(opts git.CloneOptions) error {
			// Extract command name from URL
			var cmdName string
			if opts.URL == "https://github.com/test/cmd1.git" {
				cmdName = "cmd1"
			} else if opts.URL == "https://github.com/test/cmd2.git" {
				cmdName = "cmd2"
			}

			// Create test repository structure
			metadata := models.CommandMetadata{
				Name:        cmdName,
				Version:     "1.0.0",
				Description: "Test command",
				Author:      "Test",
				Repository:  opts.URL,
			}

			yamlData, _ := metadata.MarshalYAML()
			return memFS.WriteFile(filepath.Join(opts.Target, "ccmd.yaml"), yamlData, 0o644)
		},
	}

	// Note: This test is more of an integration test outline
	// In practice, InstallFromConfig would need to be refactored to accept
	// a custom filesystem and git client for proper unit testing

	ctx := context.Background()

	// This would fail in current implementation because InstallFromConfig
	// creates its own installer instances. For proper testing, we'd need
	// to refactor to accept factory functions or dependency injection.
	_ = ctx
	_ = gitClient
}

func TestCommandManager_GetInstalledCommands(t *testing.T) {
	// Create test filesystem
	memFS := fs.NewMemFS()

	// Create project structure
	projectDir := "/test/project"
	commandsDir := filepath.Join(projectDir, ".ccmd", "commands")

	// Create test commands
	testCommands := []struct {
		name     string
		metadata models.CommandMetadata
	}{
		{
			name: "cmd1",
			metadata: models.CommandMetadata{
				Name:        "cmd1",
				Version:     "v1.0.0",
				Description: "First test command",
				Author:      "Test Author 1",
				Repository:  "https://github.com/test/cmd1.git",
			},
		},
		{
			name: "cmd2",
			metadata: models.CommandMetadata{
				Name:        "cmd2",
				Version:     "v2.0.0",
				Description: "Second test command",
				Author:      "Test Author 2",
				Repository:  "https://github.com/test/cmd2.git",
			},
		},
	}

	// Install test commands
	for _, tc := range testCommands {
		cmdDir := filepath.Join(commandsDir, tc.name)
		require.NoError(t, memFS.MkdirAll(cmdDir, 0o755))

		yamlData, err := tc.metadata.MarshalYAML()
		require.NoError(t, err)
		require.NoError(t, memFS.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), yamlData, 0o644))
	}

	// Create non-command directory (should be ignored)
	require.NoError(t, memFS.MkdirAll(filepath.Join(commandsDir, "not-a-command"), 0o755))

	// Create manager with custom filesystem
	cm := &CommandManager{
		projectPath: projectDir,
		fileSystem:  memFS,
	}

	// Get installed commands
	commands, err := cm.GetInstalledCommands()

	assert.NoError(t, err)

	// Debug: Print commands to see what's happening
	t.Logf("Got %d commands:", len(commands))
	for i, cmd := range commands {
		t.Logf("  [%d] %s - %s", i, cmd.Name, cmd.Path)
	}

	assert.Len(t, commands, 2)

	// Verify command details
	for _, cmd := range commands {
		assert.Contains(t, []string{"cmd1", "cmd2"}, cmd.Name)
		if cmd.Name == "cmd1" {
			assert.Equal(t, "v1.0.0", cmd.Version)
			assert.Equal(t, "First test command", cmd.Description)
			assert.Equal(t, "Test Author 1", cmd.Author)
		} else if cmd.Name == "cmd2" {
			assert.Equal(t, "v2.0.0", cmd.Version)
			assert.Equal(t, "Second test command", cmd.Description)
			assert.Equal(t, "Test Author 2", cmd.Author)
		}
	}
}

func TestCommandManager_GetInstalledCommands_EmptyDir(t *testing.T) {
	memFS := fs.NewMemFS()
	projectDir := "/test/project"

	cm := &CommandManager{
		projectPath: projectDir,
		fileSystem:  memFS,
	}

	// Get commands from non-existent directory
	commands, err := cm.GetInstalledCommands()

	assert.NoError(t, err)
	assert.Empty(t, commands)
}

func TestCommandManager_GetInstalledCommands_InvalidMetadata(t *testing.T) {
	memFS := fs.NewMemFS()
	projectDir := "/test/project"
	commandsDir := filepath.Join(projectDir, ".ccmd", "commands")

	// Create command with invalid metadata
	cmdDir := filepath.Join(commandsDir, "badcmd")
	require.NoError(t, memFS.MkdirAll(cmdDir, 0o755))
	require.NoError(t, memFS.WriteFile(filepath.Join(cmdDir, "ccmd.yaml"), []byte("invalid yaml: ["), 0o644))

	cm := &CommandManager{
		projectPath: projectDir,
		fileSystem:  memFS,
		logger:      logger.WithField("test", "true"),
	}

	// Should skip invalid command
	commands, err := cm.GetInstalledCommands()

	assert.NoError(t, err)
	assert.Empty(t, commands)
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/installer"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/pkg/project"
)

// TestInstallationProcess tests the complete installation workflow
func TestInstallationProcess(t *testing.T) {
	// Create in-memory filesystem
	memFS := fs.NewMemFS()

	// Create test project directory
	projectDir := "/test/project"
	require.NoError(t, memFS.MkdirAll(projectDir, 0o755))

	// Create ccmd.yaml with test commands
	config := &project.Config{
		Commands: []project.ConfigCommand{
			{Repo: "test/cmd1", Version: "v1.0.0"},
			{Repo: "test/cmd2", Version: "v2.0.0"},
		},
	}

	configData, err := yaml.Marshal(config)
	require.NoError(t, err)
	require.NoError(t, memFS.WriteFile(filepath.Join(projectDir, "ccmd.yaml"), configData, 0o644))

	// Create .claude directory
	ccmdDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, memFS.MkdirAll(ccmdDir, 0o755))

	// Test individual command installation
	t.Run("InstallSingleCommand", func(t *testing.T) {
		// Create installer options
		opts := installer.Options{
			Repository:  "https://github.com/test/example.git",
			Version:     "v1.0.0",
			InstallDir:  filepath.Join(ccmdDir, "commands"),
			ProjectPath: projectDir,
			FileSystem:  memFS,
			GitClient: &mockGitClient{
				cloneFunc: func(opts git.CloneOptions) error {
					// Simulate successful clone with valid metadata
					metadata := &models.CommandMetadata{
						Name:        "example",
						Version:     "1.0.0",
						Description: "Example command",
						Author:      "Test Author",
						Repository:  "https://github.com/test/example.git",
						Entry:       "example",
					}

					yamlData, _ := metadata.MarshalYAML()
					memFS.WriteFile(filepath.Join(opts.Target, "ccmd.yaml"), yamlData, 0o644)
					// Create index.md (required by validator)
					return memFS.WriteFile(filepath.Join(opts.Target, "index.md"), []byte("# Example Command\n\nTest documentation."), 0o644)
				},
			},
		}

		inst, err := installer.New(opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = inst.Install(ctx)
		assert.NoError(t, err)

		// Verify installation
		commandPath := filepath.Join(ccmdDir, "commands", "example")
		assert.True(t, memFS.DirExists(commandPath))
		assert.True(t, memFS.FileExists(filepath.Join(commandPath, "ccmd.yaml")))

		// Verify lock file (in project root)
		lockPath := filepath.Join(projectDir, "ccmd-lock.yaml")
		assert.True(t, memFS.FileExists(lockPath))
	})

	t.Run("InstallFromConfig", func(t *testing.T) {
		// Mock implementation would go here
		// This demonstrates the structure but would need proper mocking setup
		ctx := context.Background()

		// In a real test, we'd use dependency injection or a test factory
		// to provide a custom installer that uses our mock filesystem
		_ = ctx
	})

	t.Run("ForceReinstall", func(t *testing.T) {
		// Create existing command
		existingCmd := filepath.Join(ccmdDir, "commands", "example")
		require.NoError(t, memFS.MkdirAll(existingCmd, 0o755))
		require.NoError(t, memFS.WriteFile(filepath.Join(existingCmd, "old.txt"), []byte("old"), 0o644))

		// Install with force
		opts := installer.Options{
			Repository:  "https://github.com/test/example.git",
			Version:     "v2.0.0",
			Force:       true,
			InstallDir:  filepath.Join(ccmdDir, "commands"),
			ProjectPath: projectDir,
			FileSystem:  memFS,
			GitClient: &mockGitClient{
				cloneFunc: func(opts git.CloneOptions) error {
					metadata := &models.CommandMetadata{
						Name:        "example",
						Version:     "2.0.0",
						Description: "Updated example",
						Author:      "Test Author",
						Repository:  "https://github.com/test/example.git",
						Entry:       "example",
					}

					yamlData, _ := metadata.MarshalYAML()
					memFS.WriteFile(filepath.Join(opts.Target, "ccmd.yaml"), yamlData, 0o644)
					memFS.WriteFile(filepath.Join(opts.Target, "new.txt"), []byte("new"), 0o644)
					// Create index.md (required by validator)
					return memFS.WriteFile(filepath.Join(opts.Target, "index.md"), []byte("# Example Command\n\nTest documentation."), 0o644)
				},
			},
		}

		inst, err := installer.New(opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = inst.Install(ctx)
		assert.NoError(t, err)

		// Verify old file is gone and new file exists
		assert.False(t, memFS.FileExists(filepath.Join(existingCmd, "old.txt")))
		assert.True(t, memFS.FileExists(filepath.Join(existingCmd, "new.txt")))
	})
}

// TestInstallationErrorHandling tests error scenarios
func TestInstallationErrorHandling(t *testing.T) {
	memFS := fs.NewMemFS()
	ccmdDir := "/test/.claude"

	t.Run("RepositoryNotFound", func(t *testing.T) {
		opts := installer.Options{
			Repository: "https://github.com/nonexistent/repo.git",
			InstallDir: filepath.Join(ccmdDir, "commands"),
			FileSystem: memFS,
			GitClient: &mockGitClient{
				validateFunc: func(url string) error {
					return installer.NewInstallationError(
						"repository not found",
						url,
						"",
						installer.PhaseValidation,
					)
				},
			},
		}

		inst, err := installer.New(opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = inst.Install(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository not found")
	})

	t.Run("InvalidMetadata", func(t *testing.T) {
		opts := installer.Options{
			Repository: "https://github.com/test/invalid.git",
			InstallDir: filepath.Join(ccmdDir, "commands"),
			FileSystem: memFS,
			GitClient: &mockGitClient{
				cloneFunc: func(opts git.CloneOptions) error {
					// Create invalid ccmd.yaml
					return memFS.WriteFile(filepath.Join(opts.Target, "ccmd.yaml"), []byte("invalid: [yaml"), 0o644)
				},
			},
		}

		inst, err := installer.New(opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = inst.Install(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse metadata file")
	})

	t.Run("MissingMetadata", func(t *testing.T) {
		opts := installer.Options{
			Repository: "https://github.com/test/no-metadata.git",
			InstallDir: filepath.Join(ccmdDir, "commands"),
			FileSystem: memFS,
			GitClient: &mockGitClient{
				cloneFunc: func(opts git.CloneOptions) error {
					// Don't create ccmd.yaml
					return memFS.WriteFile(filepath.Join(opts.Target, "README.md"), []byte("Test"), 0o644)
				},
			},
		}

		inst, err := installer.New(opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = inst.Install(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ccmd.yaml not found")
	})
}

// TestRepositoryParsing tests repository URL parsing
func TestRepositoryParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedRepo  string
		expectedVer   string
		normalizedURL string
		extractedPath string
	}{
		{
			name:          "GitHub shorthand",
			input:         "user/repo",
			expectedRepo:  "user/repo",
			expectedVer:   "",
			normalizedURL: "https://github.com/user/repo.git",
			extractedPath: "user/repo",
		},
		{
			name:          "GitHub shorthand with version",
			input:         "user/repo@v1.0.0",
			expectedRepo:  "user/repo",
			expectedVer:   "v1.0.0",
			normalizedURL: "https://github.com/user/repo.git",
			extractedPath: "user/repo",
		},
		{
			name:          "Full HTTPS URL",
			input:         "https://github.com/user/repo.git",
			expectedRepo:  "https://github.com/user/repo.git",
			expectedVer:   "",
			normalizedURL: "https://github.com/user/repo.git",
			extractedPath: "user/repo",
		},
		{
			name:          "SSH URL",
			input:         "git@github.com:user/repo.git",
			expectedRepo:  "git@github.com:user/repo.git",
			expectedVer:   "",
			normalizedURL: "git@github.com:user/repo.git",
			extractedPath: "user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ParseRepositorySpec
			repo, ver := installer.ParseRepositorySpec(tt.input)
			assert.Equal(t, tt.expectedRepo, repo)
			assert.Equal(t, tt.expectedVer, ver)

			// Test NormalizeRepositoryURL
			normalized := installer.NormalizeRepositoryURL(repo)
			assert.Equal(t, tt.normalizedURL, normalized)

			// Test ExtractRepoPath
			path := installer.ExtractRepoPath(normalized)
			assert.Equal(t, tt.extractedPath, path)
		})
	}
}

// mockGitClient for testing
type mockGitClient struct {
	cloneFunc    func(opts git.CloneOptions) error
	validateFunc func(url string) error
}

func (m *mockGitClient) Clone(opts git.CloneOptions) error {
	if m.cloneFunc != nil {
		return m.cloneFunc(opts)
	}
	return nil
}

func (m *mockGitClient) ValidateRemoteRepository(url string) error {
	if m.validateFunc != nil {
		return m.validateFunc(url)
	}
	return nil
}

func (m *mockGitClient) GetLatestTag(path string) (string, error) {
	return "v1.0.0", nil
}

func (m *mockGitClient) GetCurrentCommit(path string) (string, error) {
	return "1234567890abcdef1234567890abcdef12345678", nil
}

func (m *mockGitClient) IsGitRepository(path string) bool {
	return true
}

// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package installer

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/git"
	"github.com/gifflet/ccmd/internal/metadata"
	"github.com/gifflet/ccmd/internal/models"
	"github.com/gifflet/ccmd/internal/validation"
	"github.com/gifflet/ccmd/pkg/project"
)

// BenchmarkInstallCommand benchmarks the full installation process
func BenchmarkInstallCommand(b *testing.B) {
	sizes := []struct {
		name      string
		fileCount int
		fileSize  int
	}{
		{"Small", 10, 1024},    // 10 files, 1KB each
		{"Medium", 50, 10240},  // 50 files, 10KB each
		{"Large", 100, 102400}, // 100 files, 100KB each
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			// Setup
			memFS := fs.NewMemFS()
			gitClient := &mockGitClient{
				cloneFunc: func(opts git.CloneOptions) error {
					return createBenchmarkRepository(memFS, opts.Target, size.fileCount, size.fileSize)
				},
			}

			opts := Options{
				Repository: "https://github.com/bench/test.git",
				InstallDir: ".claude/commands",
				FileSystem: memFS,
				GitClient:  gitClient,
			}

			// Create lock directory
			_ = memFS.MkdirAll(".claude", 0o755)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Clean up from previous iteration
				_ = memFS.RemoveAll(".claude/commands/benchcmd")

				installer, err := New(opts)
				if err != nil {
					b.Fatal(err)
				}
				ctx := context.Background()
				_ = installer.Install(ctx)
			}

			b.ReportMetric(float64(size.fileCount), "files")
			b.ReportMetric(float64(size.fileSize*size.fileCount)/1024, "KB_total")
		})
	}
}

// BenchmarkCopyDirectory benchmarks directory copying performance
func BenchmarkCopyDirectory(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dFiles", size), func(b *testing.B) {
			memFS := fs.NewMemFS()
			installer := &Installer{
				fileSystem: memFS,
			}

			// Create source directory with files
			srcDir := "/tmp/src"
			_ = memFS.MkdirAll(srcDir, 0o755)

			for i := 0; i < size; i++ {
				content := make([]byte, 1024) // 1KB files
				_ = memFS.WriteFile(filepath.Join(srcDir, fmt.Sprintf("file%d.txt", i)), content, 0o644)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				dstDir := fmt.Sprintf("/tmp/dst%d", i)
				_ = installer.copyDirectory(srcDir, dstDir)
			}

			b.ReportMetric(float64(size), "files")
		})
	}
}

// BenchmarkLockFileOperations benchmarks lock file read/write operations
func BenchmarkLockFileOperations(b *testing.B) {
	commandCounts := []int{10, 50, 100, 500}

	for _, count := range commandCounts {
		b.Run(fmt.Sprintf("%dCommands", count), func(b *testing.B) {
			memFS := fs.NewMemFS()
			_ = memFS.MkdirAll(".claude", 0o755)

			// Create lock manager
			lockManager := project.NewLockManagerWithFS(".claude", memFS)
			_ = lockManager.Load()

			// Add commands
			for i := 0; i < count; i++ {
				cmd := &project.CommandLockInfo{
					Name:     fmt.Sprintf("cmd%d", i),
					Version:  "v1.0.0",
					Source:   fmt.Sprintf("https://github.com/test/cmd%d.git", i),
					Resolved: fmt.Sprintf("https://github.com/test/cmd%d.git@v1.0.0", i),
				}
				_ = lockManager.AddCommand(cmd)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Benchmark save and load cycle
				_ = lockManager.Save()
				_ = lockManager.Load()
			}

			b.ReportMetric(float64(count), "commands")
		})
	}
}

// BenchmarkRepositoryValidation benchmarks repository structure validation
func BenchmarkRepositoryValidation(b *testing.B) {
	memFS := fs.NewMemFS()
	validator := validation.NewValidator(memFS)
	metaManager := metadata.NewManager()

	// Create test repository
	repoPath := "/test/repo"
	_ = memFS.MkdirAll(repoPath, 0o755)

	// Create valid metadata
	meta := &models.CommandMetadata{
		Name:        "testcmd",
		Version:     "1.0.0",
		Description: "Test command",
		Author:      "Test",
		Repository:  "https://github.com/test/repo.git",
		Entry:       "testcmd",
	}

	yamlData, _ := meta.MarshalYAML()
	_ = memFS.WriteFile(filepath.Join(repoPath, "ccmd.yaml"), yamlData, 0o644)

	// Add some files
	for i := 0; i < 10; i++ {
		_ = memFS.WriteFile(filepath.Join(repoPath, fmt.Sprintf("file%d.sh", i)), []byte("#!/bin/bash"), 0o755)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Validate repository structure
		_, _ = metaManager.ReadCommandMetadata(repoPath)
		_ = validator.ValidateCommandStructure(repoPath)
	}
}

// createBenchmarkRepository creates a repository with specified number and size of files
func createBenchmarkRepository(fs fs.FileSystem, path string, fileCount, fileSize int) error {
	// Create metadata
	metadata := &models.CommandMetadata{
		Name:        "benchcmd",
		Version:     "1.0.0",
		Description: "Benchmark command",
		Author:      "Benchmark",
		Repository:  "https://github.com/bench/test.git",
		Entry:       "benchcmd",
	}

	yamlData, err := metadata.MarshalYAML()
	if err != nil {
		return err
	}

	if err := fs.WriteFile(filepath.Join(path, "ccmd.yaml"), yamlData, 0o644); err != nil {
		return err
	}

	// Create index.md (required by validator)
	if err := fs.WriteFile(filepath.Join(path, "index.md"), []byte("# Benchmark Command"), 0o644); err != nil {
		return err
	}

	// Create files
	content := make([]byte, fileSize)
	for i := 0; i < fileCount; i++ {
		filename := filepath.Join(path, fmt.Sprintf("file%d.dat", i))
		if err := fs.WriteFile(filename, content, 0o644); err != nil {
			return err
		}
	}

	return nil
}

// BenchmarkNormalizeRepositoryURL benchmarks URL normalization
func BenchmarkNormalizeRepositoryURL(b *testing.B) {
	urls := []string{
		"user/repo",
		"github.com/user/repo",
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"https://gitlab.com/user/repo",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, url := range urls {
			_ = NormalizeRepositoryURL(url)
		}
	}

	b.ReportMetric(float64(len(urls)), "urls_per_op")
}

// BenchmarkParseRepositorySpec benchmarks repository spec parsing
func BenchmarkParseRepositorySpec(b *testing.B) {
	specs := []string{
		"user/repo",
		"user/repo@v1.0.0",
		"https://github.com/user/repo.git@v2.0.0",
		"git@github.com:user/repo.git",
		"git@github.com:user/repo.git@feature/branch",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, spec := range specs {
			_, _ = ParseRepositorySpec(spec)
		}
	}

	b.ReportMetric(float64(len(specs)), "specs_per_op")
}

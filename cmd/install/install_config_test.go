package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gifflet/ccmd/pkg/project"
)

func TestRunInstallFromConfig(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	// Test 1: No ccmd.yaml file
	err = runInstallFromConfig(false)
	if err == nil {
		t.Error("expected error when no ccmd.yaml exists")
	}
	if err != nil && err.Error() != "no ccmd.yaml found in current directory" {
		t.Errorf("unexpected error: %v", err)
	}

	// Test 2: Empty ccmd.yaml
	config := &project.Config{
		Commands: []project.ConfigCommand{},
	}
	if err := project.SaveConfig(config, filepath.Join(tempDir, project.ConfigFileName)); err != nil {
		t.Fatal(err)
	}

	err = runInstallFromConfig(false)
	if err != nil {
		t.Errorf("unexpected error with empty config: %v", err)
	}

	// Test 3: ccmd.yaml with commands (will fail due to non-existent repos, but we can test the flow)
	config = &project.Config{
		Commands: []project.ConfigCommand{
			{
				Repo:    "test/repo1",
				Version: "v1.0.0",
			},
			{
				Repo:    "test/repo2",
				Version: "",
			},
		},
	}
	if err := project.SaveConfig(config, filepath.Join(tempDir, project.ConfigFileName)); err != nil {
		t.Fatal(err)
	}

	// This will fail because the repos don't exist, but we're testing the flow
	_ = runInstallFromConfig(false)
	// No assertions here since we expect failures due to non-existent repos
}

func TestUpdateProjectLockFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	// Create project manager
	pm := project.NewManager(".")

	// Test creating new lock file
	err = updateProjectLockFile(pm, "test-cmd", "https://github.com/test/test-cmd.git", "v1.0.0")
	if err != nil {
		t.Errorf("unexpected error creating lock file: %v", err)
	}

	// Verify lock file was created
	if !pm.LockExists() {
		t.Error("expected lock file to be created")
	}

	// Load lock file and verify content
	lockFile, err := pm.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}

	cmd, exists := lockFile.GetCommand("test-cmd")
	if !exists {
		t.Error("expected command to exist in lock file")
	}

	if cmd.Name != "test-cmd" {
		t.Errorf("expected command name 'test-cmd', got %q", cmd.Name)
	}

	if cmd.Repository != "https://github.com/test/test-cmd.git" {
		t.Errorf("expected repository 'https://github.com/test/test-cmd.git', got %q", cmd.Repository)
	}

	if cmd.Version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got %q", cmd.Version)
	}

	// Test updating existing lock file
	err = updateProjectLockFile(pm, "another-cmd", "https://github.com/test/another-cmd.git", "v2.0.0")
	if err != nil {
		t.Errorf("unexpected error updating lock file: %v", err)
	}

	// Reload and verify both commands exist
	lockFile, err = pm.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}

	if len(lockFile.Commands) != 2 {
		t.Errorf("expected 2 commands in lock file, got %d", len(lockFile.Commands))
	}
}

package integration_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gifflet/ccmd/pkg/project"
)

func TestInstallFromProjectConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build ccmd binary
	buildCmd := exec.Command("go", "build", "-o", "ccmd-test", "../../cmd/ccmd")
	if err := buildCmd.Run(); err != nil {
		t.Fatal("failed to build ccmd:", err)
	}
	defer os.Remove("ccmd-test")

	// Create temporary directory for test
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

	// Create a test ccmd.yaml
	config := &project.Config{
		Commands: []project.ConfigCommand{
			{
				Repo:    "gifflet/hello-world",
				Version: "v1.0.0",
			},
			{
				Repo:    "gifflet/test-cmd",
				Version: "", // Test latest version
			},
		},
	}

	// Save config
	if err := project.SaveConfig(config, filepath.Join(tempDir, project.ConfigFileName)); err != nil {
		t.Fatal(err)
	}

	// Run install command without arguments
	cmd := exec.Command(filepath.Join(oldDir, "ccmd-test"), "install")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	output := out.String()

	// Note: We expect this to fail because the test repositories don't exist
	// But we can still verify the command behavior
	if err == nil {
		t.Log("Note: install succeeded unexpectedly, test repos might exist")
	}

	// Verify output contains expected messages
	expectedMessages := []string{
		"Installing 2 command(s) from ccmd.yaml",
		"Installing gifflet/hello-world",
		"Installing gifflet/test-cmd",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("expected output to contain %q, got:\n%s", msg, output)
		}
	}
}

func TestInstallWithArgsUpdatesProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build ccmd binary
	buildCmd := exec.Command("go", "build", "-o", "ccmd-test", "../../cmd/ccmd")
	if err := buildCmd.Run(); err != nil {
		t.Fatal("failed to build ccmd:", err)
	}
	defer os.Remove("ccmd-test")

	// Create temporary directory for test
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

	// Create empty ccmd.yaml
	config := &project.Config{
		Commands: []project.ConfigCommand{},
	}

	// Save config
	if err := project.SaveConfig(config, filepath.Join(tempDir, project.ConfigFileName)); err != nil {
		t.Fatal(err)
	}

	// Run install command with repository argument
	cmd := exec.Command(filepath.Join(oldDir, "ccmd-test"), "install", "github.com/gifflet/hello-world@v1.0.0")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	output := out.String()

	// Note: We expect this to fail because the test repository doesn't exist
	// But we can still verify some command behavior
	if err == nil {
		t.Log("Note: install succeeded unexpectedly, test repo might exist")
	}

	// Verify output contains expected messages
	expectedMessages := []string{
		"Installing command from: https://github.com/gifflet/hello-world.git",
		"Version: v1.0.0",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("expected output to contain %q, got:\n%s", msg, output)
		}
	}
}

func TestInstallNoConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build ccmd binary
	buildCmd := exec.Command("go", "build", "-o", "ccmd-test", "../../cmd/ccmd")
	if err := buildCmd.Run(); err != nil {
		t.Fatal("failed to build ccmd:", err)
	}
	defer os.Remove("ccmd-test")

	// Create temporary directory for test
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

	// Run install command without arguments (no ccmd.yaml exists)
	cmd := exec.Command(filepath.Join(oldDir, "ccmd-test"), "install")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	output := out.String()

	if err == nil {
		t.Fatal("expected error when no ccmd.yaml exists")
	}

	// Verify error message
	if !strings.Contains(output, "no ccmd.yaml found") {
		t.Errorf("expected error about missing ccmd.yaml, got output:\n%s", output)
	}
}

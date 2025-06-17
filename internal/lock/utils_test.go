package lock

import (
	"testing"
	"time"

	"github.com/gifflet/ccmd/internal/fs"
	"github.com/gifflet/ccmd/internal/models"
)

func TestManager_ListCommandsSorted(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add commands with different install times
	now := time.Now()
	commands := []*models.Command{
		{
			Name:        "zebra",
			InstalledAt: now.Add(-3 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			Name:        "alpha",
			InstalledAt: now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-30 * time.Minute),
		},
		{
			Name:        "beta",
			InstalledAt: now.Add(-1 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
	}

	// Add commands directly to bypass AddCommand's time setting
	for _, cmd := range commands {
		manager.lockFile.Commands[cmd.Name] = cmd
	}

	tests := []struct {
		name     string
		sortBy   string
		expected []string
	}{
		{
			name:     "sort by name",
			sortBy:   "name",
			expected: []string{"alpha", "beta", "zebra"},
		},
		{
			name:     "sort by install date",
			sortBy:   "installed",
			expected: []string{"beta", "alpha", "zebra"},
		},
		{
			name:     "sort by update date",
			sortBy:   "updated",
			expected: []string{"alpha", "zebra", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted, err := manager.ListCommandsSorted(tt.sortBy)
			if err != nil {
				t.Fatalf("ListCommandsSorted() error = %v", err)
			}

			if len(sorted) != len(tt.expected) {
				t.Fatalf("expected %d commands, got %d", len(tt.expected), len(sorted))
			}

			for i, name := range tt.expected {
				if sorted[i].Name != name {
					t.Errorf("position %d: expected %s, got %s", i, name, sorted[i].Name)
				}
			}
		})
	}
}

func TestManager_SearchCommands(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add test commands
	commands := []*models.Command{
		{
			Name:   "git-helper",
			Source: "github.com/example/git-helper",
			Metadata: map[string]string{
				"description": "A helper tool for version control operations",
				"author":      "John Doe",
			},
		},
		{
			Name:   "docker-compose-helper",
			Source: "gitlab.com/example/docker-tools",
			Metadata: map[string]string{
				"description": "Docker compose utilities",
				"author":      "Jane Smith",
			},
		},
		{
			Name:   "test-runner",
			Source: "gitlab.com/testing/runner",
			Metadata: map[string]string{
				"description": "Test execution framework",
				"type":        "testing",
			},
		},
	}

	for _, cmd := range commands {
		_ = manager.AddCommand(cmd)
	}

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "search by name",
			query:    "git-h",
			expected: []string{"git-helper"},
		},
		{
			name:     "search by partial name",
			query:    "helper",
			expected: []string{"docker-compose-helper", "git-helper"},
		},
		{
			name:     "search by source",
			query:    "github",
			expected: []string{"git-helper"},
		},
		{
			name:     "search by metadata",
			query:    "testing",
			expected: []string{"test-runner"},
		},
		{
			name:     "search case insensitive",
			query:    "DOCKER",
			expected: []string{"docker-compose-helper"},
		},
		{
			name:     "search not found",
			query:    "notfound",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := manager.SearchCommands(tt.query)
			if err != nil {
				t.Fatalf("SearchCommands() error = %v", err)
			}

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(results))
			}

			for i, name := range tt.expected {
				if results[i].Name != name {
					t.Errorf("position %d: expected %s, got %s", i, name, results[i].Name)
				}
			}
		})
	}
}

func TestManager_GetCommandsBySource(t *testing.T) {
	memFS := fs.NewMemFS()
	manager := NewManagerWithFS(".", memFS)
	_ = manager.Load()

	// Add commands from different sources
	sources := map[string][]string{
		"github.com/example/tools": {"tool1", "tool2"},
		"github.com/other/utils":   {"util1"},
		"github.com/test/apps":     {"app1", "app2", "app3"},
	}

	for source, names := range sources {
		for _, name := range names {
			cmd := &models.Command{
				Name:   name,
				Source: source,
			}
			_ = manager.AddCommand(cmd)
		}
	}

	// Test getting commands by source
	for source, expectedNames := range sources {
		t.Run(source, func(t *testing.T) {
			commands, err := manager.GetCommandsBySource(source)
			if err != nil {
				t.Fatalf("GetCommandsBySource() error = %v", err)
			}

			if len(commands) != len(expectedNames) {
				t.Fatalf("expected %d commands, got %d", len(expectedNames), len(commands))
			}

			// Verify all commands have the correct source
			for _, cmd := range commands {
				if cmd.Source != source {
					t.Errorf("expected source %s, got %s", source, cmd.Source)
				}
			}
		})
	}
}

func TestManager_ImportExport(t *testing.T) {
	// Create source manager with commands
	sourceFS := fs.NewMemFS()
	sourceManager := NewManagerWithFS(".", sourceFS)
	_ = sourceManager.Load()

	commands := map[string]*models.Command{
		"cmd1": {Name: "cmd1", Version: "1.0.0"},
		"cmd2": {Name: "cmd2", Version: "2.0.0"},
		"cmd3": {Name: "cmd3", Version: "3.0.0"},
	}

	for _, cmd := range commands {
		_ = sourceManager.AddCommand(cmd)
	}

	// Export commands
	exported, err := sourceManager.ExportCommands()
	if err != nil {
		t.Fatalf("ExportCommands() error = %v", err)
	}

	// Create destination manager
	destFS := fs.NewMemFS()
	destManager := NewManagerWithFS(".", destFS)
	_ = destManager.Load()

	// Add one command that will be overwritten
	_ = destManager.AddCommand(&models.Command{
		Name:    "cmd1",
		Version: "0.5.0", // Old version
	})

	// Import with overwrite
	if err := destManager.ImportCommands(exported, true); err != nil {
		t.Fatalf("ImportCommands() error = %v", err)
	}

	// Verify all commands were imported
	count, _ := destManager.CountCommands()
	if count != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), count)
	}

	// Verify cmd1 was overwritten
	cmd1, _ := destManager.GetCommand("cmd1")
	if cmd1.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cmd1.Version)
	}

	// Test import without overwrite
	destManager2 := NewManagerWithFS(".", fs.NewMemFS())
	_ = destManager2.Load()
	_ = destManager2.AddCommand(&models.Command{
		Name:    "cmd1",
		Version: "0.5.0",
	})

	if err := destManager2.ImportCommands(exported, false); err != nil {
		t.Fatalf("ImportCommands() error = %v", err)
	}

	// Verify cmd1 was NOT overwritten
	cmd1, _ = destManager2.GetCommand("cmd1")
	if cmd1.Version != "0.5.0" {
		t.Errorf("expected version 0.5.0 (not overwritten), got %s", cmd1.Version)
	}
}

package lock

import (
	"sort"
	"strings"

	"github.com/gifflet/ccmd/internal/models"
)

// CommandSorter provides different sorting methods for commands
type CommandSorter struct {
	commands []*models.Command
}

// ByName sorts commands alphabetically by name
func (s CommandSorter) ByName() {
	sort.Slice(s.commands, func(i, j int) bool {
		return s.commands[i].Name < s.commands[j].Name
	})
}

// ByInstallDate sorts commands by installation date (newest first)
func (s CommandSorter) ByInstallDate() {
	sort.Slice(s.commands, func(i, j int) bool {
		return s.commands[i].InstalledAt.After(s.commands[j].InstalledAt)
	})
}

// ByUpdateDate sorts commands by update date (newest first)
func (s CommandSorter) ByUpdateDate() {
	sort.Slice(s.commands, func(i, j int) bool {
		return s.commands[i].UpdatedAt.After(s.commands[j].UpdatedAt)
	})
}

// ListCommandsSorted returns a sorted list of commands
func (m *Manager) ListCommandsSorted(sortBy string) ([]*models.Command, error) {
	commands, err := m.ListCommands()
	if err != nil {
		return nil, err
	}

	sorter := CommandSorter{commands: commands}

	switch strings.ToLower(sortBy) {
	case "name":
		sorter.ByName()
	case "installed", "install-date":
		sorter.ByInstallDate()
	case "updated", "update-date":
		sorter.ByUpdateDate()
	default:
		sorter.ByName() // Default to name sorting
	}

	return sorter.commands, nil
}

// SearchCommands returns commands matching the search query
func (m *Manager) SearchCommands(query string) ([]*models.Command, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, ErrNotLoaded
	}

	query = strings.ToLower(query)
	var matches []*models.Command

	for _, cmd := range m.lockFile.Commands {
		// Search in name, source, and metadata
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Source), query) {
			cmdCopy := *cmd
			matches = append(matches, &cmdCopy)
			continue
		}

		// Search in metadata
		for key, value := range cmd.Metadata {
			if strings.Contains(strings.ToLower(key), query) ||
				strings.Contains(strings.ToLower(value), query) {
				cmdCopy := *cmd
				matches = append(matches, &cmdCopy)
				break
			}
		}
	}

	// Sort results by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	return matches, nil
}

// GetCommandsBySource returns all commands from a specific source
func (m *Manager) GetCommandsBySource(source string) ([]*models.Command, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, ErrNotLoaded
	}

	var commands []*models.Command
	for _, cmd := range m.lockFile.Commands {
		if cmd.Source == source {
			cmdCopy := *cmd
			commands = append(commands, &cmdCopy)
		}
	}

	return commands, nil
}

// CountCommands returns the total number of installed commands
func (m *Manager) CountCommands() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return 0, ErrNotLoaded
	}

	return len(m.lockFile.Commands), nil
}

// ExportCommands exports all commands as a slice for backup or transfer
func (m *Manager) ExportCommands() (map[string]*models.Command, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.loaded {
		return nil, ErrNotLoaded
	}

	// Create a deep copy of the commands map
	export := make(map[string]*models.Command)
	for name, cmd := range m.lockFile.Commands {
		cmdCopy := *cmd
		export[name] = &cmdCopy
	}

	return export, nil
}

// ImportCommands imports commands from a backup or transfer
func (m *Manager) ImportCommands(commands map[string]*models.Command, overwrite bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		return ErrNotLoaded
	}

	for name, cmd := range commands {
		if _, exists := m.lockFile.Commands[name]; exists && !overwrite {
			continue
		}
		cmdCopy := *cmd
		m.lockFile.Commands[name] = &cmdCopy
	}

	return nil
}

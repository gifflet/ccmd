package fs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MemFS is an in-memory implementation of FileSystem for testing
type MemFS struct {
	mu    sync.RWMutex
	files map[string]*memFile
	env   map[string]string
}

// memFile represents a file in memory
type memFile struct {
	data    []byte
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// memFileInfo implements fs.FileInfo
type memFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *memFileInfo) Name() string       { return fi.name }
func (fi *memFileInfo) Size() int64        { return fi.size }
func (fi *memFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *memFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *memFileInfo) IsDir() bool        { return fi.isDir }
func (fi *memFileInfo) Sys() interface{}   { return nil }

// memDirEntry implements fs.DirEntry
type memDirEntry struct {
	name  string
	isDir bool
	mode  os.FileMode
}

func (de *memDirEntry) Name() string      { return de.name }
func (de *memDirEntry) IsDir() bool       { return de.isDir }
func (de *memDirEntry) Type() os.FileMode { return de.mode.Type() }
func (de *memDirEntry) Info() (fs.FileInfo, error) {
	return &memFileInfo{
		name:    de.name,
		size:    0,
		mode:    de.mode,
		modTime: time.Now(),
		isDir:   de.isDir,
	}, nil
}

// NewMemFS creates a new in-memory file system
func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string]*memFile),
		env:   make(map[string]string),
	}
}

// ReadFile reads the named file and returns its contents
func (m *MemFS) ReadFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name = filepath.Clean(name)
	file, exists := m.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}

	if file.isDir {
		return nil, fmt.Errorf("read %s: is a directory", name)
	}

	data := make([]byte, len(file.data))
	copy(data, file.data)
	return data, nil
}

// WriteFile writes data to the named file, creating it if necessary
func (m *MemFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = filepath.Clean(name)

	// Create parent directories if necessary
	dir := filepath.Dir(name)
	if dir != "." && dir != "/" {
		if err := m.mkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	m.files[name] = &memFile{
		data:    bytes.Clone(data),
		mode:    perm,
		modTime: time.Now(),
		isDir:   false,
	}

	return nil
}

// Remove removes the named file or directory
func (m *MemFS) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = filepath.Clean(name)
	if _, exists := m.files[name]; !exists {
		return os.ErrNotExist
	}

	delete(m.files, name)
	return nil
}

// RemoveAll removes path and any children it contains
func (m *MemFS) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path = filepath.Clean(path)
	found := false

	// Remove all files that start with the path
	toDelete := []string{}
	for name := range m.files {
		if name == path || strings.HasPrefix(name, path+string(filepath.Separator)) {
			toDelete = append(toDelete, name)
			found = true
		}
	}

	for _, name := range toDelete {
		delete(m.files, name)
	}

	if !found {
		return os.ErrNotExist
	}

	return nil
}

// Rename renames (moves) oldpath to newpath
func (m *MemFS) Rename(oldpath, newpath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldpath = filepath.Clean(oldpath)
	newpath = filepath.Clean(newpath)

	file, exists := m.files[oldpath]
	if !exists {
		return os.ErrNotExist
	}

	// Create parent directories for newpath if necessary
	dir := filepath.Dir(newpath)
	if dir != "." && dir != "/" {
		if err := m.mkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	m.files[newpath] = file
	delete(m.files, oldpath)
	return nil
}

// Stat returns a FileInfo describing the named file
func (m *MemFS) Stat(name string) (fs.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name = filepath.Clean(name)
	file, exists := m.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}

	size := int64(0)
	if !file.isDir {
		size = int64(len(file.data))
	}

	return &memFileInfo{
		name:    filepath.Base(name),
		size:    size,
		mode:    file.mode,
		modTime: file.modTime,
		isDir:   file.isDir,
	}, nil
}

// MkdirAll creates a directory named path, along with any necessary parents
func (m *MemFS) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.mkdirAll(path, perm)
}

// ReadDir reads the directory named by dirname and returns a list of directory entries
func (m *MemFS) ReadDir(name string) ([]fs.DirEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name = filepath.Clean(name)

	// Check if directory exists
	if name != "." && name != "/" {
		file, exists := m.files[name]
		if !exists || !file.isDir {
			return nil, os.ErrNotExist
		}
	}

	entries := make([]fs.DirEntry, 0)
	seen := make(map[string]bool)

	for path, file := range m.files {
		// Skip if not in this directory
		if !strings.HasPrefix(path, name) {
			continue
		}

		// Get relative path
		rel, err := filepath.Rel(name, path)
		if err != nil {
			continue
		}

		// Skip if in subdirectory
		if strings.Contains(rel, string(filepath.Separator)) {
			// But add the subdirectory itself
			parts := strings.Split(rel, string(filepath.Separator))
			dirName := parts[0]
			if !seen[dirName] {
				seen[dirName] = true
				entries = append(entries, &memDirEntry{
					name:  dirName,
					isDir: true,
					mode:  0o755,
				})
			}
			continue
		}

		// Skip self
		if rel == "." {
			continue
		}

		entries = append(entries, &memDirEntry{
			name:  filepath.Base(path),
			isDir: file.isDir,
			mode:  file.mode,
		})
	}

	return entries, nil
}

// Exists checks if a path exists
func (m *MemFS) Exists(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path = filepath.Clean(path)
	_, exists := m.files[path]
	return exists, nil
}

// Setenv sets an environment variable for testing
func (m *MemFS) Setenv(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.env[key] = value
	return nil
}

// Getenv gets an environment variable for testing
func (m *MemFS) Getenv(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.env[key]
}

// mkdirAll is the internal implementation that assumes the lock is held
func (m *MemFS) mkdirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	if path == "." || path == "/" {
		return nil
	}

	// Check if already exists
	if file, exists := m.files[path]; exists {
		if !file.isDir {
			return fmt.Errorf("mkdir %s: not a directory", path)
		}
		return nil
	}

	// Create parent directories
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		if err := m.mkdirAll(parent, perm); err != nil {
			return err
		}
	}

	m.files[path] = &memFile{
		mode:    perm | os.ModeDir,
		modTime: time.Now(),
		isDir:   true,
	}

	return nil
}

// List returns all files matching the pattern
func (m *MemFS) List(pattern string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var files []string
	for path := range m.files {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			files = append(files, path)
		}
	}
	return files
}

// Clear removes all files from the file system
func (m *MemFS) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files = make(map[string]*memFile)
}

// String returns a string representation of the file system for debugging
func (m *MemFS) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	paths := make([]string, 0, len(m.files))
	for path := range m.files {
		paths = append(paths, path)
	}
	return "MemFS{" + strings.Join(paths, ", ") + "}"
}

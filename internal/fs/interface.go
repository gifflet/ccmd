package fs

import (
	"io/fs"
	"os"
)

// FileSystem is an interface for file system operations
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (fs.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
}

// OS implements FileSystem using the real file system
type OS struct{}

// ReadFile reads the named file and returns its contents
func (OS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to the named file, creating it if necessary
func (OS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Remove removes the named file or directory
func (OS) Remove(name string) error {
	return os.Remove(name)
}

// Rename renames (moves) oldpath to newpath
func (OS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat returns a FileInfo describing the named file
func (OS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll creates a directory named path, along with any necessary parents
func (OS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

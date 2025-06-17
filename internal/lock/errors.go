package lock

import "errors"

// Common errors
var (
	// ErrNotLoaded is returned when operations are attempted before loading the lock file
	ErrNotLoaded = errors.New("lock file not loaded")

	// ErrCommandNotFound is returned when a command is not found in the lock file
	ErrCommandNotFound = errors.New("command not found")

	// ErrCommandExists is returned when trying to add a command that already exists
	ErrCommandExists = errors.New("command already exists")
)

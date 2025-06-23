/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package repository

import "fmt"

// Common error types for repository operations
var (
	// ErrRepositoryNotFound indicates the repository doesn't exist or is inaccessible
	ErrRepositoryNotFound = fmt.Errorf("repository not found")

	// ErrCommandNotFound indicates the command doesn't exist in the repository
	ErrCommandNotFound = fmt.Errorf("command not found")

	// ErrInvalidConfiguration indicates the repository configuration is invalid
	ErrInvalidConfiguration = fmt.Errorf("invalid repository configuration")

	// ErrAuthenticationFailed indicates authentication to the repository failed
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")

	// ErrResourceNotFound indicates the requested resource doesn't exist
	ErrResourceNotFound = fmt.Errorf("resource not found")

	// ErrUnsupportedOperation indicates the operation is not supported by this repository type
	ErrUnsupportedOperation = fmt.Errorf("unsupported operation")

	// ErrCacheExpired indicates the cache is outdated and needs refresh
	ErrCacheExpired = fmt.Errorf("cache expired")

	// ErrNetworkError indicates a network-related error occurred
	ErrNetworkError = fmt.Errorf("network error")

	// ErrVersionConflict indicates multiple versions exist and resolution is ambiguous
	ErrVersionConflict = fmt.Errorf("version conflict")

	// ErrPermissionDenied indicates insufficient permissions to access the repository
	ErrPermissionDenied = fmt.Errorf("permission denied")
)

// Error provides detailed error information for repository operations
type Error struct {
	Op         string // Operation that failed
	Repository string // Repository identifier
	Err        error  // Underlying error
	Details    string // Additional details
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s %s: %v (%s)", e.Op, e.Repository, e.Err, e.Details)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.Repository, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository Error
func NewRepositoryError(op, repo string, err error, details string) error {
	return &Error{
		Op:         op,
		Repository: repo,
		Err:        err,
		Details:    details,
	}
}

// CommandError provides detailed error information for command operations
type CommandError struct {
	CommandID string // Command identifier
	Op        string // Operation that failed
	Err       error  // Underlying error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command %s: %s failed: %v", e.CommandID, e.Op, e.Err)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// NewCommandError creates a new CommandError
func NewCommandError(commandID, op string, err error) error {
	return &CommandError{
		CommandID: commandID,
		Op:        op,
		Err:       err,
	}
}

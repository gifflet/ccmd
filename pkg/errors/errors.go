/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package errors provides simple error handling for CCMD.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error types
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrGitOperation  = errors.New("git operation failed")
	ErrFileOperation = errors.New("file operation failed")
)

// errorWithContextFormat is the format string for errors with context
const errorWithContextFormat = "%w: %s"

// NotFound creates a not found error with context
func NotFound(resource string) error {
	return fmt.Errorf(errorWithContextFormat, ErrNotFound, resource)
}

// AlreadyExists creates an already exists error with context
func AlreadyExists(resource string) error {
	return fmt.Errorf(errorWithContextFormat, ErrAlreadyExists, resource)
}

// InvalidInput creates an invalid input error with context
func InvalidInput(msg string) error {
	return fmt.Errorf(errorWithContextFormat, ErrInvalidInput, msg)
}

// GitError creates a git operation error with context
func GitError(operation string, err error) error {
	if err == nil {
		return fmt.Errorf("%w during %s", ErrGitOperation, operation)
	}
	return fmt.Errorf("%w during %s: %v", ErrGitOperation, operation, err)
}

// FileError creates a file operation error with context
func FileError(operation, path string, err error) error {
	if err == nil {
		return fmt.Errorf("%w: %s on %s", ErrFileOperation, operation, path)
	}
	return fmt.Errorf("%w: %s on %s: %v", ErrFileOperation, operation, path, err)
}

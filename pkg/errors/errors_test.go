/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		sentinel error
	}{
		{"NotFound", NotFound("user"), ErrNotFound},
		{"AlreadyExists", AlreadyExists("command"), ErrAlreadyExists},
		{"InvalidInput", InvalidInput("bad format"), ErrInvalidInput},
		{"GitError", GitError("clone", nil), ErrGitOperation},
		{"FileError", FileError("read", "/tmp/test", nil), ErrFileOperation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.sentinel) {
				t.Errorf("expected error to match sentinel %v", tt.sentinel)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"NotFound", NotFound("user"), "not found: user"},
		{"AlreadyExists", AlreadyExists("command"), "already exists: command"},
		{"InvalidInput", InvalidInput("bad format"), "invalid input: bad format"},
		{"GitError", GitError("clone", nil), "git operation failed during clone"},
		{"FileError", FileError("read", "/tmp/test", nil), "file operation failed: read on /tmp/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorWithCause(t *testing.T) {
	cause := errors.New("network timeout")

	gitErr := GitError("clone", cause)
	if !errors.Is(gitErr, ErrGitOperation) {
		t.Error("GitError should wrap ErrGitOperation")
	}
	if gitErr.Error() != "git operation failed during clone: network timeout" {
		t.Errorf("unexpected error message: %v", gitErr)
	}

	fileErr := FileError("write", "/data/file", cause)
	if !errors.Is(fileErr, ErrFileOperation) {
		t.Error("FileError should wrap ErrFileOperation")
	}
	if fileErr.Error() != "file operation failed: write on /data/file: network timeout" {
		t.Errorf("unexpected error message: %v", fileErr)
	}
}

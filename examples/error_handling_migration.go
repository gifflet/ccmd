/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package main demonstrates how to migrate from old error handling to the new system
package main

import (
	"fmt"

	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// OldWay shows the old error handling approach
func OldWay() error {
	// Before: Simple error with fmt.Errorf
	if err := someOperation(); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}
	return nil
}

// NewWay shows the new structured error handling
func NewWay() error {
	log := logger.WithField("function", "NewWay")

	// After: Structured error with logging
	if err := someOperation(); err != nil {
		log.WithError(err).Error("operation failed")
		return errors.Wrap(err, errors.CodeInternal, "operation failed").
			WithDetail("operation", "someOperation").
			WithDetail("context", "example")
	}

	log.Debug("operation completed successfully")
	return nil
}

// ExampleGitError shows git-specific error handling
func ExampleGitError(repoURL string) error {
	log := logger.WithField("function", "ExampleGitError")

	// Simulate a git clone error
	if err := gitClone(repoURL); err != nil {
		log.WithError(err).WithField("repository", repoURL).Error("git clone failed")

		// Return a structured error with context
		return errors.Wrap(err, errors.CodeGitClone, "failed to clone repository").
			WithDetail("repository", repoURL).
			WithDetail("suggestion", "check your network connection and repository URL")
	}

	return nil
}

// ExampleValidationError shows validation error handling
func ExampleValidationError(config map[string]interface{}) error {
	// Validate configuration
	if config["version"] == nil {
		return errors.New(errors.CodeValidationFailed, "missing required field: version").
			WithDetail("file", "ccmd.yaml").
			WithDetail("field", "version").
			WithDetail("suggestion", "add 'version: 1.0.0' to your ccmd.yaml")
	}

	// Check version format
	version, ok := config["version"].(string)
	if !ok {
		return errors.New(errors.CodeValidationFailed, "version must be a string").
			WithDetail("file", "ccmd.yaml").
			WithDetail("field", "version").
			WithDetail("actual_type", fmt.Sprintf("%T", config["version"]))
	}

	logger.WithField("version", version).Debug("configuration validated")
	return nil
}

// ExampleCommandError shows command-specific error handling
func ExampleCommandError(cmdName string) error {
	// Check if command exists
	if !commandExists(cmdName) {
		// This error will be displayed user-friendly by the handler
		return errors.New(errors.CodeCommandNotFound, "command not found").
			WithDetail("command", cmdName)
	}

	// Execute command
	if err := executeCommand(cmdName); err != nil {
		return errors.Wrap(err, errors.CodeCommandExecute, "command execution failed").
			WithDetail("command", cmdName)
	}

	return nil
}

// Helper functions for examples
func someOperation() error             { return nil }
func gitClone(url string) error        { return nil }
func commandExists(name string) bool   { return true }
func executeCommand(name string) error { return nil }

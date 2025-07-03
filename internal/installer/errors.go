/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package installer

import (
	stderrors "errors"
	"fmt"

	"github.com/gifflet/ccmd/pkg/errors"
)

// Common installation error messages
const (
	ErrMsgRepositoryNotFound = "repository not found or inaccessible"
	ErrMsgInvalidRepository  = "invalid repository structure"
	ErrMsgMissingMetadata    = "ccmd.yaml not found in repository"
	ErrMsgInvalidMetadata    = "invalid ccmd.yaml format"
	ErrMsgCommandExists      = "command already exists"
	ErrMsgInstallationFailed = "installation failed"
	ErrMsgRollbackFailed     = "failed to rollback installation"
	ErrMsgVersionNotFound    = "specified version not found"
	ErrMsgNoVersionAvailable = "no version information available"
	ErrMsgPermissionDenied   = "permission denied"
	ErrMsgDiskFull           = "insufficient disk space"
	ErrMsgNetworkError       = "network error during download"
	ErrMsgInvalidCommandName = "invalid command name"
	ErrMsgLockFileCorrupted  = "lock file is corrupted"
	ErrMsgConcurrentInstall  = "another installation is in progress"
)

// Installation phases
const (
	PhaseValidation   = "validation"
	PhaseClone        = "clone"
	PhaseVerification = "verification"
	PhaseInstall      = "install"
	PhaseMetadata     = "metadata"
	PhaseLockFile     = "lock_file"
	PhaseCleanup      = "cleanup"
	PhaseRollback     = "rollback"
)

// WrapInstallationError adds installation context to an error
func WrapInstallationError(err error, repository, version, phase string) error {
	if err == nil {
		return nil
	}

	// Build context message
	var context string
	if repository != "" {
		context = fmt.Sprintf(" (repository: %s", repository)
		if version != "" {
			context += fmt.Sprintf(", version: %s", version)
		}
		context += ")"
	}
	if phase != "" {
		context += fmt.Sprintf(" [phase: %s]", phase)
	}

	// Return error with context appended
	return fmt.Errorf("%w%s", err, context)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Git and file errors are typically retryable
	return stderrors.Is(err, errors.ErrGitOperation) || stderrors.Is(err, errors.ErrFileOperation)
}

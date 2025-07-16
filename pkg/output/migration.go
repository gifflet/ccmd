/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package output

import (
	"errors"

	ccmderrors "github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// PrintError prints an error to the user in appropriate format
func PrintError(err error) {
	if err == nil {
		return
	}

	// Check error type
	switch {
	case errors.Is(err, ccmderrors.ErrNotFound):
		PrintWarningf("%s", err.Error())
	case errors.Is(err, ccmderrors.ErrAlreadyExists):
		PrintWarningf("%s", err.Error())
	case errors.Is(err, ccmderrors.ErrInvalidInput):
		PrintErrorf("Validation failed: %s", err.Error())
	case errors.Is(err, ccmderrors.ErrFileOperation):
		PrintErrorf("File operation failed: %s", err.Error())
	case errors.Is(err, ccmderrors.ErrGitOperation):
		PrintErrorf("Git operation failed: %s", err.Error())
	default:
		PrintErrorf("Error: %s", err.Error())
	}
}

// LogAndPrintError logs an error with context and outputs to user
func LogAndPrintError(log logger.Logger, err error, message string) {
	if err == nil {
		return
	}

	// Log with context
	log.WithError(err).Error(message)

	// Output to user
	PrintError(err)
}

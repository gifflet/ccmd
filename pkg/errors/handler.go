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
	"fmt"
	"os"

	"github.com/gifflet/ccmd/pkg/logger"
)

// DefaultHandler is the default error handler instance
var DefaultHandler = &Handler{
	logger: logger.GetDefault(),
}

// Handler provides centralized error handling with logging
type Handler struct {
	logger logger.Logger
}

// Handle processes an error and returns true if it was handled
func (h *Handler) Handle(err error) bool {
	if err == nil {
		return false
	}

	// Log the error
	h.logger.WithError(err).Error("error occurred")

	// Display user-friendly message based on error type
	switch {
	case errors.Is(err, ErrNotFound):
		fmt.Fprintf(os.Stderr, "Not found: %v\n", err)
	case errors.Is(err, ErrAlreadyExists):
		fmt.Fprintf(os.Stderr, "Already exists: %v\n", err)
	case errors.Is(err, ErrInvalidInput):
		fmt.Fprintf(os.Stderr, "Invalid input: %v\n", err)
	case errors.Is(err, ErrGitOperation):
		fmt.Fprintf(os.Stderr, "Git operation failed: %v\n", err)
	case errors.Is(err, ErrFileOperation):
		fmt.Fprintf(os.Stderr, "File operation failed: %v\n", err)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	return true
}

// HandleFatal processes a fatal error and exits
func (h *Handler) HandleFatal(err error) {
	if err == nil {
		return
	}

	h.Handle(err)
	os.Exit(1)
}

// Handle is a convenience function using the default handler
func Handle(err error) bool {
	return DefaultHandler.Handle(err)
}

// HandleFatal is a convenience function using the default handler
func HandleFatal(err error) {
	DefaultHandler.HandleFatal(err)
}

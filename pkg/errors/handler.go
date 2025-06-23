// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package errors

import (
	"errors"
	"fmt"
	"os"

	"github.com/gifflet/ccmd/pkg/logger"
)

// OutputFunc is a function that outputs messages to the user
type OutputFunc func(format string, args ...interface{})

// Handler provides centralized error handling with logging
type Handler struct {
	logger       logger.Logger
	printError   OutputFunc
	printWarning OutputFunc
	printInfo    OutputFunc
}

// NewHandler creates a new error handler
func NewHandler(log logger.Logger) *Handler {
	if log == nil {
		log = logger.GetDefault()
	}
	return &Handler{
		logger:       log,
		printError:   defaultPrintErrorf,
		printWarning: defaultPrintWarningf,
		printInfo:    defaultPrintInfof,
	}
}

// SetOutputFuncs sets custom output functions
func (h *Handler) SetOutputFuncs(printError, printWarning, printInfo OutputFunc) {
	if printError != nil {
		h.printError = printError
	}
	if printWarning != nil {
		h.printWarning = printWarning
	}
	if printInfo != nil {
		h.printInfo = printInfo
	}
}

// Default output functions
func defaultPrintErrorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func defaultPrintWarningf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func defaultPrintInfof(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// DefaultHandler is the default error handler instance
var DefaultHandler = NewHandler(nil)

// Handle processes an error and returns true if it was handled
func (h *Handler) Handle(err error) bool {
	if err == nil {
		return false
	}

	// Extract structured error if available
	var ccmdErr *Error
	if As(err, &ccmdErr) {
		h.handleStructuredError(ccmdErr)
	} else {
		h.handleGenericError(err)
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

// handleStructuredError handles our structured errors
func (h *Handler) handleStructuredError(err *Error) {
	// Build fields for logging
	fields := logger.Fields{
		"code": string(err.Code),
	}

	// Add details if present
	for k, v := range err.Details {
		fields[k] = v
	}

	// Add cause if present
	if err.Cause != nil {
		fields["cause"] = err.Cause.Error()
	}

	// Log based on error code severity
	switch {
	case IsNotFound(err):
		h.logger.WithFields(fields).Warn(err.Message)
		h.printWarning("%s", h.getUserMessage(err))

	case IsAlreadyExists(err):
		h.logger.WithFields(fields).Warn(err.Message)
		h.printWarning("%s", h.getUserMessage(err))

	case IsValidationError(err):
		h.logger.WithFields(fields).Error(err.Message)
		h.printError("Validation failed: %s", h.getUserMessage(err))

	case IsPermissionDenied(err):
		h.logger.WithFields(fields).Error(err.Message)
		h.printError("Permission denied: %s", h.getUserMessage(err))

	case IsGitError(err):
		h.logger.WithFields(fields).Error(err.Message)
		h.printError("Git operation failed: %s", h.getUserMessage(err))

	case IsNetworkError(err):
		h.logger.WithFields(fields).Error(err.Message)
		h.printError("Network error: %s", h.getUserMessage(err))

	default:
		h.logger.WithFields(fields).Error(err.Message)
		h.printError("Error: %s", h.getUserMessage(err))
	}
}

// handleGenericError handles standard errors
func (h *Handler) handleGenericError(err error) {
	h.logger.WithError(err).Error("unhandled error")
	h.printError("Error: %v", err)
}

// getUserMessage formats an error for user display
func (h *Handler) getUserMessage(err *Error) string {
	msg := err.Message

	// Add helpful context based on error code
	switch err.Code {
	case CodeCommandNotFound:
		if cmd, ok := err.Details["command"].(string); ok {
			msg = fmt.Sprintf("Command '%s' not found. Use 'ccmd list' to see available commands.", cmd)
		}

	case CodeGitClone:
		if repo, ok := err.Details["repository"].(string); ok {
			msg = fmt.Sprintf("Failed to clone repository '%s'. Check the URL and connection.", repo)
		}

	case CodeFilePermission:
		if path, ok := err.Details["path"].(string); ok {
			msg = fmt.Sprintf("Permission denied accessing '%s'. Please check file permissions.", path)
		}

	case CodeConfigInvalid:
		if file, ok := err.Details["file"].(string); ok {
			msg = fmt.Sprintf("Invalid configuration in '%s': %s", file, err.Message)
		}

	case CodeNetworkTimeout:
		msg = "Network operation timed out. Please check your internet connection and try again."
	}

	return msg
}

// Handle is a convenience function using the default handler
func Handle(err error) bool {
	return DefaultHandler.Handle(err)
}

// HandleFatal is a convenience function using the default handler
func HandleFatal(err error) {
	DefaultHandler.HandleFatal(err)
}

// As is a wrapper around errors.As for convenience
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is is a wrapper around errors.Is for convenience
func Is(err, target error) bool {
	return errors.Is(err, target)
}

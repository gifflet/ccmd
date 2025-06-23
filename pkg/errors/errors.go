/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package errors provides comprehensive error handling for CCMD.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// CodeUnknown represents an unknown error
	CodeUnknown ErrorCode = "UNKNOWN"
	// CodeInternal represents an internal error
	CodeInternal ErrorCode = "INTERNAL"
	// CodeInvalidArgument represents an invalid argument error
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	// CodeNotFound represents a not found error
	CodeNotFound ErrorCode = "NOT_FOUND"
	// CodeAlreadyExists represents an already exists error
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	// CodePermissionDenied represents a permission denied error
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// CodeGitClone represents a git clone error
	CodeGitClone ErrorCode = "GIT_CLONE"
	// CodeGitInvalidRepo represents an invalid git repository error
	CodeGitInvalidRepo ErrorCode = "GIT_INVALID_REPO"
	// CodeGitAuth represents a git authentication error
	CodeGitAuth ErrorCode = "GIT_AUTH"
	// CodeGitNotFound represents a git not found error
	CodeGitNotFound ErrorCode = "GIT_NOT_FOUND"

	// CodeCommandNotFound represents a command not found error
	CodeCommandNotFound ErrorCode = "COMMAND_NOT_FOUND"
	// CodeCommandExists represents a command already exists error
	CodeCommandExists ErrorCode = "COMMAND_EXISTS"
	// CodeCommandInvalid represents an invalid command error
	CodeCommandInvalid ErrorCode = "COMMAND_INVALID"
	// CodeCommandExecute represents a command execution error
	CodeCommandExecute ErrorCode = "COMMAND_EXECUTE"

	// CodeConfigInvalid represents an invalid configuration error
	CodeConfigInvalid ErrorCode = "CONFIG_INVALID"
	// CodeConfigNotFound represents a configuration not found error
	CodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	// CodeConfigParse represents a configuration parse error
	CodeConfigParse ErrorCode = "CONFIG_PARSE"

	// CodeFileNotFound represents a file not found error
	CodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	// CodeFileExists represents a file already exists error
	CodeFileExists ErrorCode = "FILE_EXISTS"
	// CodeFilePermission represents a file permission error
	CodeFilePermission ErrorCode = "FILE_PERMISSION"
	// CodeFileIO represents a file I/O error
	CodeFileIO ErrorCode = "FILE_IO"

	// CodeValidationFailed represents a validation failure
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	// CodeValidationSchema represents a schema validation error
	CodeValidationSchema ErrorCode = "VALIDATION_SCHEMA"

	// CodeNetworkTimeout represents a network timeout error
	CodeNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"
	// CodeNetworkUnavailable represents a network unavailable error
	CodeNetworkUnavailable ErrorCode = "NETWORK_UNAVAILABLE"

	// CodeValidation represents a general validation error
	CodeValidation ErrorCode = "VALIDATION"
	// CodeLockConflict represents a lock conflict error
	CodeLockConflict ErrorCode = "LOCK_CONFLICT"
	// CodeTimeout represents a general timeout error
	CodeTimeout ErrorCode = "TIMEOUT"
	// CodePartialFailure represents a partial failure where some operations succeeded
	CodePartialFailure ErrorCode = "PARTIAL_FAILURE"
	// CodeNotImplemented represents a not implemented error
	CodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// Error represents a structured error with code, message and context
type Error struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
	Cause   error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New creates a new error with the given code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Newf creates a new error with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Details: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Details: make(map[string]interface{}),
		Cause:   err,
	}
}

// WithDetail adds a detail to the error
func (e *Error) WithDetail(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeUnknown
}

// IsCode checks if an error has a specific code
func IsCode(err error, code ErrorCode) bool {
	return GetCode(err) == code
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	code := GetCode(err)
	return code == CodeNotFound || code == CodeGitNotFound ||
		code == CodeCommandNotFound || code == CodeConfigNotFound ||
		code == CodeFileNotFound
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	code := GetCode(err)
	return code == CodeAlreadyExists || code == CodeCommandExists ||
		code == CodeFileExists
}

// IsPermissionDenied checks if an error is a permission denied error
func IsPermissionDenied(err error) bool {
	code := GetCode(err)
	return code == CodePermissionDenied || code == CodeFilePermission
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	code := GetCode(err)
	return code == CodeValidationFailed || code == CodeValidationSchema ||
		code == CodeConfigInvalid || code == CodeCommandInvalid
}

// IsGitError checks if an error is git related
func IsGitError(err error) bool {
	code := GetCode(err)
	return code == CodeGitClone || code == CodeGitInvalidRepo ||
		code == CodeGitAuth || code == CodeGitNotFound
}

// IsNetworkError checks if an error is network related
func IsNetworkError(err error) bool {
	code := GetCode(err)
	return code == CodeNetworkTimeout || code == CodeNetworkUnavailable
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors occurred (%d errors)", len(e.Errors))
}

// NewMulti creates a new MultiError from a list of errors
func NewMulti(errs ...error) error {
	// Filter out nil errors
	var nonNilErrors []error
	for _, err := range errs {
		if err != nil {
			nonNilErrors = append(nonNilErrors, err)
		}
	}

	if len(nonNilErrors) == 0 {
		return nil
	}
	if len(nonNilErrors) == 1 {
		return nonNilErrors[0]
	}

	return &MultiError{Errors: nonNilErrors}
}

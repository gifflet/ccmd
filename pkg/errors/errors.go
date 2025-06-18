// Package errors provides comprehensive error handling for CCMD.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// General errors
	CodeUnknown          ErrorCode = "UNKNOWN"
	CodeInternal         ErrorCode = "INTERNAL"
	CodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// Git related errors
	CodeGitClone       ErrorCode = "GIT_CLONE"
	CodeGitInvalidRepo ErrorCode = "GIT_INVALID_REPO"
	CodeGitAuth        ErrorCode = "GIT_AUTH"
	CodeGitNotFound    ErrorCode = "GIT_NOT_FOUND"

	// Command related errors
	CodeCommandNotFound ErrorCode = "COMMAND_NOT_FOUND"
	CodeCommandExists   ErrorCode = "COMMAND_EXISTS"
	CodeCommandInvalid  ErrorCode = "COMMAND_INVALID"
	CodeCommandExecute  ErrorCode = "COMMAND_EXECUTE"

	// Configuration errors
	CodeConfigInvalid  ErrorCode = "CONFIG_INVALID"
	CodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	CodeConfigParse    ErrorCode = "CONFIG_PARSE"

	// File system errors
	CodeFileNotFound   ErrorCode = "FILE_NOT_FOUND"
	CodeFileExists     ErrorCode = "FILE_EXISTS"
	CodeFilePermission ErrorCode = "FILE_PERMISSION"
	CodeFileIO         ErrorCode = "FILE_IO"

	// Validation errors
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	CodeValidationSchema ErrorCode = "VALIDATION_SCHEMA"

	// Network errors
	CodeNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	CodeNetworkUnavailable ErrorCode = "NETWORK_UNAVAILABLE"
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

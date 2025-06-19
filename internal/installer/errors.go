package installer

import (
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

// InstallationError represents an error that occurred during installation
type InstallationError struct {
	err        error
	code       errors.Code
	repository string
	version    string
	phase      string
}

// NewInstallationError creates a new installation error
func NewInstallationError(code errors.Code, message, repository, version, phase string) *InstallationError {
	return &InstallationError{
		err:        errors.New(code, message),
		code:       code,
		repository: repository,
		version:    version,
		phase:      phase,
	}
}

// Error implements the error interface
func (e *InstallationError) Error() string {
	msg := e.err.Error()
	if e.repository != "" {
		msg = fmt.Sprintf("%s (repository: %s", msg, e.repository)
		if e.version != "" {
			msg += fmt.Sprintf(", version: %s", e.version)
		}
		msg += ")"
	}
	if e.phase != "" {
		msg += fmt.Sprintf(" [phase: %s]", e.phase)
	}
	return msg
}

// Unwrap returns the underlying error
func (e *InstallationError) Unwrap() error {
	return e.err
}

// GetCode returns the error code
func (e *InstallationError) GetCode() errors.Code {
	return e.code
}

// Phase returns the installation phase where the error occurred
func (e *InstallationError) Phase() string {
	return e.phase
}

// Repository returns the repository that was being installed
func (e *InstallationError) Repository() string {
	return e.repository
}

// Version returns the version that was being installed
func (e *InstallationError) Version() string {
	return e.version
}

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

// WrapInstallationError wraps an error with installation context
func WrapInstallationError(err error, code errors.Code, repository, version, phase string) error {
	if err == nil {
		return nil
	}

	// If it's already an installation error, update fields
	if instErr, ok := err.(*InstallationError); ok {
		if repository != "" {
			instErr.repository = repository
		}
		if version != "" {
			instErr.version = version
		}
		if phase != "" {
			instErr.phase = phase
		}
		return instErr
	}

	// Create new installation error
	return &InstallationError{
		err:        err,
		code:       code,
		repository: repository,
		version:    version,
		phase:      phase,
	}
}

// IsRetryableError checks if an installation error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an errors.Error
	if e, ok := err.(*errors.Error); ok {
		switch e.Code {
		case errors.CodeGitClone, errors.CodeGitAuth, errors.CodeTimeout:
			return true
		case errors.CodeFileIO:
			// Check if it's a temporary filesystem error
			if e.Details != nil && e.Details["temporary"] == "true" {
				return true
			}
		}
	}

	// Check if it's an InstallationError
	if instErr, ok := err.(*InstallationError); ok {
		return IsRetryableError(instErr.err)
	}

	return false
}

// GetInstallationPhase extracts the installation phase from an error
func GetInstallationPhase(err error) string {
	if err == nil {
		return ""
	}

	if instErr, ok := err.(*InstallationError); ok {
		return instErr.phase
	}

	return ""
}

// ValidationError represents a validation error during installation
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError
}

// Error implements the error interface
func (e *MultiValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d issues found", len(e.Errors))
}

// Add adds a validation error
func (e *MultiValidationError) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are any validation errors
func (e *MultiValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

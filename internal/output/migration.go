package output

import (
	"github.com/gifflet/ccmd/pkg/errors"
	"github.com/gifflet/ccmd/pkg/logger"
)

// ErrorToOutput converts an error to appropriate output format
func ErrorToOutput(err error) {
	if err == nil {
		return
	}

	// Check if it's a structured error
	var ccmdErr *errors.Error
	if errors.As(err, &ccmdErr) {
		switch {
		case errors.IsNotFound(err):
			PrintWarningf("%s", err.Error())
		case errors.IsAlreadyExists(err):
			PrintWarningf("%s", err.Error())
		case errors.IsValidationError(err):
			PrintErrorf("Validation failed: %s", err.Error())
		case errors.IsPermissionDenied(err):
			PrintErrorf("Permission denied: %s", err.Error())
		default:
			PrintErrorf("Error: %s", err.Error())
		}
	} else {
		// Standard error
		PrintErrorf("Error: %v", err)
	}
}

// DebugError logs an error at debug level with context
func DebugError(err error, context string) {
	if err == nil {
		return
	}

	logger.WithError(err).WithField("context", context).Debug("error occurred")
	Debugf("%s: %v", context, err)
}

// LogAndPrintf logs at the appropriate level and prints to output
func LogAndPrintf(level, format string, args ...interface{}) {
	switch level {
	case "success":
		logger.Infof("[SUCCESS] "+format, args...)
		PrintSuccessf(format, args...)
	case "error":
		logger.Errorf("[ERROR] "+format, args...)
		PrintErrorf(format, args...)
	case "warning":
		logger.Warnf("[WARNING] "+format, args...)
		PrintWarningf(format, args...)
	case "info":
		logger.Infof("[INFO] "+format, args...)
		PrintInfof(format, args...)
	default:
		logger.Infof(format, args...)
		Printf(format, args...)
	}
}

// WithLogger creates a context-aware output function
func WithLogger(log logger.Logger) func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		log.Infof(format, args...)
		Printf(format, args...)
	}
}

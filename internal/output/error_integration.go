package output

import (
	"github.com/gifflet/ccmd/pkg/errors"
)

// InitializeErrorHandler sets up the error handler with output functions
func InitializeErrorHandler() {
	handler := errors.DefaultHandler
	handler.SetOutputFuncs(
		func(format string, args ...interface{}) {
			PrintErrorf(format, args...)
		},
		func(format string, args ...interface{}) {
			PrintWarningf(format, args...)
		},
		func(format string, args ...interface{}) {
			PrintInfof(format, args...)
		},
	)
}

// HandleError is a convenience function to handle errors with output
func HandleError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Handle(err)
}

// HandleFatalError handles a fatal error and exits
func HandleFatalError(err error) {
	if err != nil {
		errors.HandleFatal(err)
	}
}

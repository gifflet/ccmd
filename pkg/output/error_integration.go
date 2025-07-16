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
	"github.com/gifflet/ccmd/pkg/errors"
)

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

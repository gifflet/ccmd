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
	"github.com/spf13/cobra"

	"github.com/gifflet/ccmd/pkg/logger"
)

// CommandRunner wraps a cobra command with error handling
type CommandRunner func(cmd *cobra.Command, args []string) error

// WrapCommand wraps a command runner with error handling and logging
func WrapCommand(name string, runner CommandRunner) CommandRunner {
	return func(cmd *cobra.Command, args []string) error {
		log := logger.WithField("command", name)

		// Log command execution
		log.Debugf("executing command with args: %v", args)

		// Run the command
		err := runner(cmd, args)

		if err != nil {
			// Log the error with context
			log.WithError(err).Error("command failed")

			// Handle the error for user display
			Handle(err)
			return err
		}

		log.Debug("command completed successfully")
		return nil
	}
}

// RecoverPanic recovers from panics and converts them to errors
func RecoverPanic(name string) {
	if r := recover(); r != nil {
		log := logger.WithField("command", name)

		// Convert panic to error
		var err error
		switch v := r.(type) {
		case error:
			err = v
		case string:
			err = New(CodeInternal, v)
		default:
			err = Newf(CodeInternal, "panic: %v", v)
		}

		// Log the panic
		log.WithField("panic", r).Fatal("command panicked")

		// Handle as fatal error
		HandleFatal(err)
	}
}

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package logger_test

import (
	"bytes"
	"fmt"

	"github.com/gifflet/ccmd/pkg/logger"
)

func ExampleLogger_basic() {
	// Create a logger writing to a buffer for testing
	var buf bytes.Buffer
	log := logger.New(&buf, logger.InfoLevel)

	// Log at different levels
	log.Debug("This won't be shown") // Below threshold
	log.Info("Application started")
	log.Warn("Low memory")
	log.Error("Connection failed")

	// Check output (timestamps will vary)
	output := buf.String()
	fmt.Println(len(output) > 0)
	// Output: true
}

func ExampleLogger_withFields() {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.InfoLevel)

	// Add fields to provide context
	log.WithFields(logger.Fields{
		"user_id": 123,
		"action":  "login",
	}).Info("User logged in")

	// Fields are included in output
	output := buf.String()
	fmt.Println(len(output) > 0)
	// Output: true
}

func ExampleLogger_chaining() {
	var buf bytes.Buffer
	baseLog := logger.New(&buf, logger.InfoLevel)

	// Create specialized loggers
	apiLog := baseLog.WithField("component", "api")
	authLog := apiLog.WithField("module", "auth")

	// Each logger inherits parent fields
	authLog.Info("Authentication successful")

	output := buf.String()
	fmt.Println(len(output) > 0)
	// Output: true
}

func ExampleLogger_withError() {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.ErrorLevel)

	// Log with error context
	err := fmt.Errorf("database connection failed")
	log.WithError(err).Error("Failed to process request")

	output := buf.String()
	fmt.Println(len(output) > 0)
	// Output: true
}

func ExampleDefault() {
	// Use the default logger
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// With fields
	logger.WithField("request_id", "abc123").Info("Processing request")

	// With multiple fields
	logger.WithFields(logger.Fields{
		"user":   "john",
		"action": "upload",
		"size":   1024,
	}).Info("File uploaded")
}

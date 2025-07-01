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
	"errors"
	"fmt"

	"github.com/gifflet/ccmd/pkg/logger"
)

func ExampleLogger_basic() {
	// Create a logger
	log := logger.New()

	// Log at different levels
	log.Debug("This might not be shown depending on log level")
	log.Info("Application started")
	log.Warn("Low memory")
	log.Error("Connection failed")

	// Note: Since we're using slog internally, output goes to stderr
}

func ExampleLogger_withFields() {
	log := logger.New()

	// Add fields to provide context
	log.WithFields(logger.Fields{
		"user":   "john",
		"action": "login",
		"ip":     "127.0.0.1",
	}).Info("User authentication successful")

	// Chain multiple fields
	log.WithField("request_id", "12345").
		WithField("method", "GET").
		WithField("path", "/api/users").
		Info("API request received")
}

func ExampleLogger_contextual() {
	baseLog := logger.New()

	// Create component-specific loggers
	dbLog := baseLog.WithField("component", "database")
	apiLog := baseLog.WithField("component", "api")

	// Use them throughout the component
	dbLog.Info("Connecting to database")
	dbLog.WithField("pool_size", 10).Info("Connection pool initialized")

	apiLog.Info("Starting API server")
	apiLog.WithField("port", 8080).Info("Listening for requests")
}

func ExampleLogger_errorHandling() {
	log := logger.New()

	// Log with error context
	err := errors.New("connection timeout")
	log.WithError(err).Error("Failed to connect to database")

	// Log with multiple contexts
	log.WithFields(logger.Fields{
		"host":    "db.example.com",
		"port":    5432,
		"timeout": "30s",
	}).WithError(err).Error("Database connection failed")

	// Conditional error logging
	if err != nil {
		log.WithField("operation", "user_create").
			WithError(err).
			Error("Operation failed")
	}
}

func ExampleLogger_formatted() {
	log := logger.New()

	// Use formatted logging methods
	count := 42
	name := "test-command"

	log.Infof("Processing %d items", count)
	log.Debugf("Installing command: %s", name)
	log.Warnf("Retrying operation, attempt %d of %d", 2, 3)
	log.Errorf("Failed to download %s: timeout after %d seconds", name, 30)
}

func Example_globalFunctions() {
	// Use global logger functions
	logger.Info("Application starting")
	logger.Debug("Debug mode enabled")

	// With formatting
	logger.Infof("Listening on port %d", 8080)
	logger.Errorf("Failed to bind to port %d", 8080)

	// With fields
	logger.WithField("version", "1.0.0").Info("Application initialized")
	logger.WithFields(logger.Fields{
		"user": "admin",
		"role": "superuser",
	}).Info("User logged in")

	// With error
	err := errors.New("file not found")
	logger.WithError(err).Error("Failed to load configuration")
}

func Example_packageInstallation() {
	// Real-world example: package installation logging
	log := logger.WithField("component", "installer")

	packageName := "example-cli"
	repository := "github.com/user/example-cli"

	log.WithFields(logger.Fields{
		"package":    packageName,
		"repository": repository,
	}).Info("Starting package installation")

	// Simulate installation steps
	log.WithField("step", "download").Info("Downloading package")
	log.WithField("step", "extract").Info("Extracting files")
	log.WithField("step", "build").Info("Building package")

	// Handle potential error
	err := fmt.Errorf("build failed: exit code 1")
	if err != nil {
		log.WithFields(logger.Fields{
			"package": packageName,
			"step":    "build",
		}).WithError(err).Error("Installation failed")
		return
	}

	log.WithField("package", packageName).Info("Installation completed successfully")
}

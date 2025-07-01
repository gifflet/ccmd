/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package logger provides structured logging capabilities for CCMD.
package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Fields represents structured fields for logging
type Fields map[string]interface{}

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
}

// logger is the default implementation
type logger struct {
	slogger *slog.Logger
}

// New creates a new logger instance
func New() Logger {
	level := slog.LevelInfo
	if os.Getenv("CCMD_DEBUG") == "1" || os.Getenv("CCMD_LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	return &logger{
		slogger: slog.New(handler),
	}
}

// Default creates a logger with default settings
func Default() Logger {
	return New()
}

// Debug logs a debug message
func (l *logger) Debug(msg string) {
	l.slogger.Debug(msg)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(format string, args ...interface{}) {
	l.slogger.Debug(fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *logger) Info(msg string) {
	l.slogger.Info(msg)
}

// Warn logs a warning message
func (l *logger) Warn(msg string) {
	l.slogger.Warn(msg)
}

// Error logs an error message
func (l *logger) Error(msg string) {
	l.slogger.Error(msg)
}

// Infof logs a formatted info message
func (l *logger) Infof(format string, args ...interface{}) {
	l.slogger.Info(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *logger) Errorf(format string, args ...interface{}) {
	l.slogger.Error(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *logger) Warnf(format string, args ...interface{}) {
	l.slogger.Warn(fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string) {
	l.slogger.Error(msg)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.slogger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// WithField creates a new logger with an additional field
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{
		slogger: l.slogger.With(key, value),
	}
}

// WithFields creates a new logger with additional fields
func (l *logger) WithFields(fields Fields) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &logger{
		slogger: l.slogger.With(args...),
	}
}

// WithError creates a new logger with an error field
func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// Global logger instance
var defaultLogger = Default()

// GetDefault returns the default logger
func GetDefault() Logger {
	return defaultLogger
}

// Convenience functions using the default logger

// Debug logs a debug message using the default logger
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs an info message using the default logger
func Info(msg string) {
	defaultLogger.Info(msg)
}

// Warn logs a warning message using the default logger
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// Error logs an error message using the default logger
func Error(msg string) {
	defaultLogger.Error(msg)
}

// Fatal logs an error message and exits
func Fatal(msg string) {
	slog.Error(msg)
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits
func Fatalf(format string, args ...interface{}) {
	slog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Infof logs a formatted info message using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// WithField creates a new logger with a field using the default logger
func WithField(key string, value interface{}) Logger {
	return defaultLogger.WithField(key, value)
}

// WithFields creates a new logger with fields using the default logger
func WithFields(fields Fields) Logger {
	return defaultLogger.WithFields(fields)
}

// WithError creates a new logger with an error field using the default logger
func WithError(err error) Logger {
	return defaultLogger.WithError(err)
}

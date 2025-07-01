/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package logger

import (
	"errors"
	"strings"
	"testing"
)

func TestLoggerBasic(t *testing.T) {
	l := New()

	// These should not panic
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warning message")
	l.Error("error message")
}

func TestLoggerFormatting(t *testing.T) {
	l := New()

	// Test formatted methods
	l.Debugf("Hello %s, number %d", "world", 42)
	l.Infof("Hello %s, number %d", "world", 42)
	l.Warnf("Hello %s, number %d", "world", 42)
	l.Errorf("Hello %s, number %d", "world", 42)
}

func TestLoggerWithFields(t *testing.T) {
	l := New()

	// Test with single field
	l.WithField("key", "value").Info("message with field")

	// Test with multiple fields
	l.WithFields(Fields{
		"user":   "john",
		"action": "login",
		"status": "success",
	}).Info("message with fields")
}

func TestLoggerWithError(t *testing.T) {
	l := New()

	// Test with error
	err := errors.New("something went wrong")
	l.WithError(err).Error("operation failed")

	// Test with nil error
	l.WithError(nil).Info("should not add error field")
}

func TestLoggerChaining(t *testing.T) {
	l := New()

	// Test chaining multiple WithField calls
	contextLogger := l.
		WithField("component", "installer").
		WithField("version", "1.0.0")

	contextLogger.Info("installation started")

	// Add more context
	contextLogger.
		WithField("package", "test-package").
		WithError(errors.New("download failed")).
		Error("installation failed")
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Test global functions
	Debug("debug message")
	Info("info message")
	Warn("warning message")
	Error("error message")

	Debugf("formatted %s", "debug")
	Infof("formatted %s", "info")
	Warnf("formatted %s", "warning")
	Errorf("formatted %s", "error")

	// Test global WithField functions
	WithField("key", "value").Info("with field")
	WithFields(Fields{"a": 1, "b": 2}).Info("with fields")
	WithError(errors.New("test error")).Error("with error")
}

func TestFieldsType(t *testing.T) {
	// Test that Fields type works correctly
	fields := Fields{
		"string": "value",
		"int":    42,
		"bool":   true,
		"float":  3.14,
		"nil":    nil,
		"slice":  []string{"a", "b"},
		"map":    map[string]int{"x": 1},
	}

	l := New()
	l.WithFields(fields).Info("testing various field types")
}

func TestGetDefault(t *testing.T) {
	// Test getting default logger
	def := GetDefault()
	if def == nil {
		t.Error("expected default logger, got nil")
	}

	// Should be able to use it
	def.Info("from default logger")
}

func TestFatalFunctions(t *testing.T) {
	// We can't actually test Fatal/Fatalf as they call os.Exit
	// Just ensure they're callable (compile test)
	if false {
		Fatal("fatal message")
		Fatalf("fatal %s", "formatted")
	}

	// Test that Fatal methods exist on logger interface
	l := New()
	if false {
		l.Fatal("fatal message")
		l.Fatalf("fatal %s", "formatted")
	}
}

func TestLoggerOutput(t *testing.T) {
	// Since we're using slog internally, we can't easily capture output
	// This test just ensures methods work without panicking
	l := New()

	// Test various combinations
	l.Info("simple message")
	l.WithField("key", "value").Info("with field")
	l.WithFields(Fields{"a": 1, "b": "two"}).Warn("with fields")
	l.WithError(errors.New("test")).Error("with error")

	// Test empty messages
	l.Info("")
	l.Debug("")
	l.Warn("")
	l.Error("")

	// Test special characters
	l.Info("message with\nnewline")
	l.Info("message with\ttab")
	l.Info("message with \"quotes\"")
}

func TestLoggerConcurrency(t *testing.T) {
	l := New()

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			logger := l.WithField("goroutine", n)
			for j := 0; j < 10; j++ {
				logger.Infof("message %d from goroutine %d", j, n)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoggerMethodsExist(t *testing.T) {
	// This test ensures all interface methods are implemented
	var _ Logger = &logger{}
	var _ Logger = New()
	var _ Logger = Default()
	var _ Logger = GetDefault()
}

func TestExampleUsage(t *testing.T) {
	// Example from actual usage in the codebase
	log := WithField("component", "installer")
	log.Info("Installing package")

	// Simulate error scenario
	err := errors.New("download failed")
	log.WithError(err).Error("Installation failed")

	// With multiple fields
	log.WithFields(Fields{
		"package": "test-cmd",
		"version": "1.0.0",
		"source":  "github.com/user/repo",
	}).Info("Installation complete")
}

func TestLoggerNilHandling(t *testing.T) {
	l := New()

	// Test nil field values
	l.WithField("nil_value", nil).Info("with nil field")

	// Test empty fields
	l.WithFields(Fields{}).Info("with empty fields")
	l.WithFields(nil).Info("with nil fields map")

	// Test method chaining with nil
	var nilError error
	l.WithError(nilError).Info("with nil error")

	// Test empty strings
	l.WithField("", "empty key")
	l.WithField("key", "")
	l.WithField("", "")
}

func TestLoggerSpecialValues(t *testing.T) {
	l := New()

	// Test various special values that might cause issues
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"zero int", "count", 0},
		{"negative int", "balance", -100},
		{"empty slice", "items", []string{}},
		{"nil slice", "items", []string(nil)},
		{"empty map", "data", map[string]int{}},
		{"nil map", "data", map[string]int(nil)},
		{"very long string", "text", strings.Repeat("a", 1000)},
		{"unicode", "emoji", "🚀 Hello 世界"},
		{"special chars", "path", "/path/with spaces/and-dashes_underscores.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			l.WithField(tt.key, tt.value).Info("testing special value")
		})
	}
}

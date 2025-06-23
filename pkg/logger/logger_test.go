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
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  Level
		msgLevel  Level
		shouldLog bool
	}{
		{"debug level logs all", DebugLevel, DebugLevel, true},
		{"debug level logs info", DebugLevel, InfoLevel, true},
		{"info level skips debug", InfoLevel, DebugLevel, false},
		{"info level logs info", InfoLevel, InfoLevel, true},
		{"error level skips warn", ErrorLevel, WarnLevel, false},
		{"error level logs error", ErrorLevel, ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := New(&buf, tt.logLevel)

			// Log based on message level
			switch tt.msgLevel {
			case DebugLevel:
				l.Debug("test message")
			case InfoLevel:
				l.Info("test message")
			case WarnLevel:
				l.Warn("test message")
			case ErrorLevel:
				l.Error("test message")
			}

			output := buf.String()
			if tt.shouldLog && output == "" {
				t.Error("expected log output but got none")
			}
			if !tt.shouldLog && output != "" {
				t.Errorf("expected no output but got: %s", output)
			}
		})
	}
}

func TestLoggerFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, DebugLevel)

	l.Infof("Hello %s, number %d", "world", 42)
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Error("expected [INFO] in output")
	}
	if !strings.Contains(output, "Hello world, number 42") {
		t.Error("expected formatted message in output")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	l.InfoWithFields("user action", Fields{
		"user_id": 123,
		"action":  "login",
	})

	output := buf.String()
	if !strings.Contains(output, "user action") {
		t.Error("expected message in output")
	}
	if !strings.Contains(output, "user_id=123") {
		t.Error("expected user_id field in output")
	}
	if !strings.Contains(output, "action=login") {
		t.Error("expected action field in output")
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	l2 := l.WithField("request_id", "abc123")
	l2.Info("processing request")

	output := buf.String()
	if !strings.Contains(output, "request_id=abc123") {
		t.Error("expected request_id field in output")
	}
}

func TestLoggerWithFields_Chaining(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	l2 := l.WithFields(Fields{
		"service": "api",
		"version": "1.0.0",
	})
	l3 := l2.WithField("endpoint", "/users")

	l3.Info("handling request")

	output := buf.String()
	if !strings.Contains(output, "service=api") {
		t.Error("expected service field in output")
	}
	if !strings.Contains(output, "version=1.0.0") {
		t.Error("expected version field in output")
	}
	if !strings.Contains(output, "endpoint=/users") {
		t.Error("expected endpoint field in output")
	}
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	err := errors.New("something went wrong")
	l.WithError(err).Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "error=something went wrong") {
		t.Error("expected error field in output")
	}
}

func TestLoggerWithError_Nil(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	l2 := l.WithError(nil)
	if l2 != l {
		t.Error("WithError(nil) should return the same logger")
	}
}

func TestSetGetLevel(t *testing.T) {
	l := New(&bytes.Buffer{}, InfoLevel)

	if l.GetLevel() != InfoLevel {
		t.Errorf("expected InfoLevel, got %v", l.GetLevel())
	}

	l.SetLevel(DebugLevel)
	if l.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel after SetLevel, got %v", l.GetLevel())
	}
}

func TestDefaultLogger(t *testing.T) {
	// Save original
	original := defaultLogger
	defer func() {
		defaultLogger = original
	}()

	var buf bytes.Buffer
	SetDefault(New(&buf, InfoLevel))

	Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("expected message in output from default logger")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Save original
	original := defaultLogger
	defer func() {
		defaultLogger = original
	}()

	var buf bytes.Buffer
	SetDefault(New(&buf, DebugLevel))

	// Test all convenience functions
	Debug("debug message")
	Debugf("debug %s", "formatted")
	Info("info message")
	Infof("info %s", "formatted")
	Warn("warn message")
	Warnf("warn %s", "formatted")
	Error("error message")
	Errorf("error %s", "formatted")

	output := buf.String()
	expectedMessages := []string{
		"debug message",
		"debug formatted",
		"info message",
		"info formatted",
		"warn message",
		"warn formatted",
		"error message",
		"error formatted",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("expected '%s' in output", expected)
		}
	}
}

func TestLogEntryFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, InfoLevel)

	l.Info("test message")
	output := buf.String()

	// Check for timestamp format (ISO8601)
	// The timestamp should contain 'T' as separator and timezone info
	if !strings.Contains(output, "T") {
		t.Error("expected ISO8601 timestamp with 'T' separator in output")
	}

	// Check for level
	if !strings.Contains(output, "[INFO]") {
		t.Error("expected [INFO] level in output")
	}

	// Check for message
	if !strings.Contains(output, "test message") {
		t.Error("expected message in output")
	}

	// Check for newline
	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestSourceLocation(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, DebugLevel)

	// Debug and Error should include source location
	l.Debug("debug message")
	debugOutput := buf.String()
	if !strings.Contains(debugOutput, "logger_test.go:") {
		t.Error("expected source location in debug output")
	}

	buf.Reset()
	l.Error("error message")
	errorOutput := buf.String()
	if !strings.Contains(errorOutput, "logger_test.go:") {
		t.Error("expected source location in error output")
	}

	// Info should not include source location
	buf.Reset()
	l.Info("info message")
	infoOutput := buf.String()
	if strings.Contains(infoOutput, "logger_test.go:") {
		t.Error("did not expect source location in info output")
	}
}

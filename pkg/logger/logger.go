// Package logger provides structured logging capabilities for CCMD.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	// DebugLevel logs everything
	DebugLevel Level = iota
	// InfoLevel logs info, warnings and errors
	InfoLevel
	// WarnLevel logs warnings and errors
	WarnLevel
	// ErrorLevel logs only errors
	ErrorLevel
	// FatalLevel logs fatal errors and exits
	FatalLevel
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Fields represents structured fields for logging
type Fields map[string]interface{}

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	DebugWithFields(msg string, fields Fields)

	Info(msg string)
	Infof(format string, args ...interface{})
	InfoWithFields(msg string, fields Fields)

	Warn(msg string)
	Warnf(format string, args ...interface{})
	WarnWithFields(msg string, fields Fields)

	Error(msg string)
	Errorf(format string, args ...interface{})
	ErrorWithFields(msg string, fields Fields)

	Fatal(msg string)
	Fatalf(format string, args ...interface{})
	FatalWithFields(msg string, fields Fields)

	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	SetLevel(level Level)
	GetLevel() Level
}

// logger is the default implementation
type logger struct {
	level      Level
	output     io.Writer
	fields     Fields
	mu         sync.RWMutex
	timeFormat string
}

// New creates a new logger instance
func New(output io.Writer, level Level) Logger {
	return &logger{
		level:      level,
		output:     output,
		fields:     make(Fields),
		timeFormat: "2006-01-02T15:04:05.000Z07:00",
	}
}

// Default creates a logger with default settings
func Default() Logger {
	level := InfoLevel
	if os.Getenv("CCMD_DEBUG") == "1" || os.Getenv("CCMD_LOG_LEVEL") == "debug" {
		level = DebugLevel
	}
	return New(os.Stderr, level)
}

// SetLevel sets the logging level
func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// log is the core logging function
func (l *logger) log(level Level, msg string, fields Fields) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	output := l.output
	l.mu.RUnlock()

	// Merge fields
	allFields := make(Fields)
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Format the log entry
	entry := l.formatEntry(level, msg, allFields)

	// Write to output
	_, _ = fmt.Fprint(output, entry)

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// formatEntry formats a log entry
func (l *logger) formatEntry(level Level, msg string, fields Fields) string {
	var sb strings.Builder

	// Timestamp
	sb.WriteString(time.Now().Format(l.timeFormat))
	sb.WriteString(" ")

	// Level
	sb.WriteString(fmt.Sprintf("[%s]", level.String()))
	sb.WriteString(" ")

	// Source location for debug and error levels
	if level == DebugLevel || level == ErrorLevel {
		if file, line := getSourceLocation(); file != "" {
			sb.WriteString(fmt.Sprintf("%s:%d ", file, line))
		}
	}

	// Message
	sb.WriteString(msg)

	// Fields
	if len(fields) > 0 {
		sb.WriteString(" ")
		first := true
		for k, v := range fields {
			if !first {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// getSourceLocation returns the file and line number of the caller
func getSourceLocation() (file string, line int) {
	// Skip 3 frames: getSourceLocation, formatEntry, log
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return "", 0
	}
	// Return only the file name, not the full path
	return filepath.Base(file), line
}

// Debug logs a debug message
func (l *logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// DebugWithFields logs a debug message with fields
func (l *logger) DebugWithFields(msg string, fields Fields) {
	l.log(DebugLevel, msg, fields)
}

// Info logs an info message
func (l *logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// Infof logs a formatted info message
func (l *logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// InfoWithFields logs an info message with fields
func (l *logger) InfoWithFields(msg string, fields Fields) {
	l.log(InfoLevel, msg, fields)
}

// Warn logs a warning message
func (l *logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// Warnf logs a formatted warning message
func (l *logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// WarnWithFields logs a warning message with fields
func (l *logger) WarnWithFields(msg string, fields Fields) {
	l.log(WarnLevel, msg, fields)
}

// Error logs an error message
func (l *logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// Errorf logs a formatted error message
func (l *logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// ErrorWithFields logs an error message with fields
func (l *logger) ErrorWithFields(msg string, fields Fields) {
	l.log(ErrorLevel, msg, fields)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

// Fatalf logs a formatted fatal message and exits
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), nil)
}

// FatalWithFields logs a fatal message with fields and exits
func (l *logger) FatalWithFields(msg string, fields Fields) {
	l.log(FatalLevel, msg, fields)
}

// WithField creates a new logger with an additional field
func (l *logger) WithField(key string, value interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &logger{
		level:      l.level,
		output:     l.output,
		fields:     newFields,
		timeFormat: l.timeFormat,
	}
}

// WithFields creates a new logger with additional fields
func (l *logger) WithFields(fields Fields) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &logger{
		level:      l.level,
		output:     l.output,
		fields:     newFields,
		timeFormat: l.timeFormat,
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

// SetDefault sets the default logger
func SetDefault(l Logger) {
	defaultLogger = l
}

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

// Infof logs a formatted info message using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs an error message using the default logger
func Error(msg string) {
	defaultLogger.Error(msg)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

// Fatalf logs a formatted fatal message using the default logger and exits
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
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

/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

// Package output provides colored output utilities for command-line interface.
package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Color functions for different message types
var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgBlue).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()
)

// PrintSuccessf prints a formatted success message.
func PrintSuccessf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, Success(format)+"\n", a...)
}

// PrintErrorf prints a formatted error message.
func PrintErrorf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, Error(format)+"\n", a...)
}

// PrintWarningf prints a formatted warning message.
func PrintWarningf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, Warning(format)+"\n", a...)
}

// PrintInfof prints a formatted info message.
func PrintInfof(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, Info(format)+"\n", a...)
}

// Printf prints a formatted message.
func Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", a...)
}

// Fatalf prints an error message and exits with code 1.
func Fatalf(format string, a ...interface{}) {
	PrintErrorf(format, a...)
	os.Exit(1)
}

// Prompt asks the user for input with a colored prompt
func Prompt(prompt string) string {
	fmt.Print(Info(prompt + ": "))
	var input string
	_, _ = fmt.Scanln(&input)
	return input
}

// Debugf prints a debug message if debug mode is enabled.
func Debugf(format string, a ...interface{}) {
	if os.Getenv("CCMD_DEBUG") == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", a...)
	}
}

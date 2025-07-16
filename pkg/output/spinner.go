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
	"time"
)

// Spinner provides a simple loading indicator
type Spinner struct {
	message string
	chars   []string
	delay   time.Duration
	active  bool
	done    chan bool
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		chars:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:   100 * time.Millisecond,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.active = true
	go func() {
		i := 0
		for s.active {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("\r%s %s ", Info(s.chars[i]), s.message)
				i = (i + 1) % len(s.chars)
				time.Sleep(s.delay)
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	if s.active {
		s.done <- true
		s.active = false
	}
	fmt.Print("\r\033[K") // Clear the line
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	PrintSuccessf("✓ %s", message)
}

// Error stops the spinner and shows an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	PrintErrorf("✗ %s", message)
}

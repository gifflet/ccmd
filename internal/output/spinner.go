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
	s.active = false
	s.done <- true
	fmt.Print("\r\033[K") // Clear the line
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	PrintSuccess("✓ %s", message)
}

// Error stops the spinner and shows an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	PrintError("✗ %s", message)
}

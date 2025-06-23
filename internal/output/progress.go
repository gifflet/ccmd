/*
 * This file is part of ccmd.
 *
 * Copyright (c) 2025 Guilherme Silva Sousa
 *
 * Licensed under the MIT License
 * See LICENSE file in the project root for full license information.
 */

package output

import (
	"fmt"
	"strings"
)

// ProgressBar provides a simple progress indicator
type ProgressBar struct {
	total   int
	current int
	width   int
	message string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		current: 0,
		width:   40,
		message: message,
	}
}

// Update updates the progress bar with the current value
func (p *ProgressBar) Update(current int) {
	p.current = current
	p.render()
}

// Increment increments the progress by 1
func (p *ProgressBar) Increment() {
	p.current++
	p.render()
}

// Complete marks the progress as complete
func (p *ProgressBar) Complete() {
	p.current = p.total
	p.render()
	fmt.Println() // New line after completion
}

func (p *ProgressBar) render() {
	if p.total == 0 {
		return
	}

	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	fmt.Printf("\r%s [%s] %d/%d (%.0f%%)",
		p.message,
		Info(bar),
		p.current,
		p.total,
		percent*100,
	)
}

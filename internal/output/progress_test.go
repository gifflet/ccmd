// Copyright (c) 2025 Guilherme Silva Sousa
// Licensed under the MIT License
// See LICENSE file in the project root for full license information.
package output

import (
	"testing"
)

func TestProgressBar(t *testing.T) {
	// Test creation
	pb := NewProgressBar(100, "Testing")
	if pb.total != 100 {
		t.Errorf("Expected total 100, got %d", pb.total)
	}
	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}
	if pb.message != "Testing" {
		t.Errorf("Expected message 'Testing', got '%s'", pb.message)
	}

	// Test update
	pb.Update(50)
	if pb.current != 50 {
		t.Errorf("Expected current 50 after update, got %d", pb.current)
	}

	// Test increment
	pb.Increment()
	if pb.current != 51 {
		t.Errorf("Expected current 51 after increment, got %d", pb.current)
	}

	// Test complete
	pb.Complete()
	if pb.current != pb.total {
		t.Errorf("Expected current to equal total after complete, got %d", pb.current)
	}

	// Test with zero total (should not panic)
	pbZero := NewProgressBar(0, "Zero test")
	pbZero.Update(10) // Should not panic
	pbZero.render()   // Should not panic
}

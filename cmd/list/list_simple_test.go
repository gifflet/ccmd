package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandStructure(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List all installed commands", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Equal(t, "Show detailed information", verboseFlag.Usage)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestFormatTimeFunctions(t *testing.T) {
	// Test truncateSource
	tests := []struct {
		name     string
		source   string
		maxLen   int
		expected string
	}{
		{
			name:     "short source",
			source:   "github.com/test",
			maxLen:   30,
			expected: "github.com/test",
		},
		{
			name:     "long source",
			source:   "github.com/very/long/path/repository",
			maxLen:   20,
			expected: "github.com/very/l...",
		},
		{
			name:     "exact length",
			source:   "github.com/exact",
			maxLen:   16,
			expected: "github.com/exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateSource(tt.source, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

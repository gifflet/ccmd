package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandStructure(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List all commands managed by ccmd", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	longFlag := cmd.Flags().Lookup("long")
	assert.NotNil(t, longFlag)
	assert.Equal(t, "l", longFlag.Shorthand)
	assert.Equal(t, "Show detailed output including metadata", longFlag.Usage)
	assert.Equal(t, "false", longFlag.DefValue)
}

func TestFormatTimeFunctions(t *testing.T) {
	// Test truncateText
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "short text",
			text:     "github.com/test",
			maxLen:   30,
			expected: "github.com/test",
		},
		{
			name:     "long text",
			text:     "github.com/very/long/path/repository",
			maxLen:   20,
			expected: "github.com/very/l...",
		},
		{
			name:     "exact length",
			text:     "github.com/exact",
			maxLen:   16,
			expected: "github.com/exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

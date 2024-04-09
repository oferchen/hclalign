package config_test

import (
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/assert"
)

func TestIsValidOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    []string
		expected bool
	}{
		{
			name:     "valid order with all attributes",
			order:    []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expected: true,
		},
		{
			name:     "invalid order with duplicates",
			order:    []string{"description", "type", "default", "sensitive", "nullable", "sensitive"},
			expected: false,
		},
		{
			name:     "invalid order with missing attributes",
			order:    []string{"description", "type", "default"},
			expected: false,
		},
		{
			name:     "invalid order with extra attributes",
			order:    []string{"description", "type", "default", "sensitive", "nullable", "validation", "extra"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsValidOrder(tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTargetDynamically(t *testing.T) {
	// This test assumes that fileprocessing and patternmatching packages exist and work as expected.
	// It is also assumed that the implementation of ProcessTargetDynamically only validates the inputs
	// and forwards them to the fileprocessing.ProcessFiles function.
	// Therefore, these tests focus on input validation.

	tests := []struct {
		name        string
		target      string
		criteria    []string
		order       []string
		expectError bool
	}{
		{
			name:        "valid input",
			target:      "path/to/target",
			criteria:    []string{"*.tf"},
			order:       []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expectError: false,
		},
		{
			name:        "invalid criteria",
			target:      "path/to/target",
			criteria:    []string{"invalid[regex"},
			order:       []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expectError: true,
		},
		{
			name:        "empty target",
			target:      "",
			criteria:    []string{"*.tf"},
			order:       []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expectError: true,
		},
		{
			name:        "invalid order",
			target:      "path/to/target",
			criteria:    []string{"*.tf"},
			order:       []string{"default", "type"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ProcessTargetDynamically(tt.target, tt.criteria, tt.order)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

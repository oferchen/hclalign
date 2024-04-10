package patternmatching

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilePatterns(t *testing.T) {
	tests := []struct {
		name        string
		criteria    []string
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "compile single pattern",
			criteria:    []string{"*.txt"},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name:        "compile multiple patterns",
			criteria:    []string{"*.txt", "*.go", "main.*"},
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:     "compile invalid pattern",
			criteria: []string{"[a-z"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := CompilePatterns(tt.criteria)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(patterns))
			}
		})
	}
}

func TestMatchesFileCriteria(t *testing.T) {
	compiledPatterns, _ := CompilePatterns([]string{"*.txt", "README.*"})
	tests := []struct {
		name          string
		filePath      string
		expectToMatch bool
	}{
		{
			name:          "match txt file",
			filePath:      "document.txt",
			expectToMatch: true,
		},
		{
			name:          "match README",
			filePath:      "README.md",
			expectToMatch: true,
		},
		{
			name:          "no match",
			filePath:      "image.jpg",
			expectToMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesFileCriteria(tt.filePath, compiledPatterns)
			assert.Equal(t, tt.expectToMatch, result)
		})
	}
}

func TestIsValidCriteria(t *testing.T) {
	tests := []struct {
		name          string
		criteria      []string
		expectedValid bool
	}{
		{
			name:          "valid single wildcard",
			criteria:      []string{"*.tf"},
			expectedValid: true,
		},
		{
			name:          "valid multiple criteria",
			criteria:      []string{"main.tf", "*.jpg", "directory/"},
			expectedValid: true,
		},
		{
			name:          "invalid pattern",
			criteria:      []string{"???.>tf"},
			expectedValid: false,
		},
		{
			name:          "empty criterion",
			criteria:      []string{""},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCriteria(tt.criteria)
			assert.Equal(t, tt.expectedValid, result)
		})
	}
}

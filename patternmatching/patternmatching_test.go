package patternmatching

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesFileCriteria(t *testing.T) {
	criteria := []string{"*.txt", "README.*"}
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
			result := MatchesFileCriteria(tt.filePath, criteria)
			assert.Equal(t, tt.expectToMatch, result)
		})
	}
}

func TestIsValidCriteria(t *testing.T) {
	tests := []struct {
		name       string
		criteria   []string
		expectErr  bool
		errMessage string
	}{
		{
			name:      "valid single wildcard",
			criteria:  []string{"*.tf"},
			expectErr: false,
		},
		{
			name:      "valid multiple criteria",
			criteria:  []string{"main.tf", "*.jpg", "directory/"},
			expectErr: false,
		},
		{
			name:       "invalid pattern",
			criteria:   []string{"["},
			expectErr:  true,
			errMessage: "invalid criterion",
		},
		{
			name:       "empty criterion",
			criteria:   []string{""},
			expectErr:  true,
			errMessage: "criterion is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidCriteria(tt.criteria)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

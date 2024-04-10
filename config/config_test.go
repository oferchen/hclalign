package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/fileprocessing"
)

func TestIsValidOrder(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		order     []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid order",
			order:     []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expectErr: false,
		},
		{
			name:      "duplicate item",
			order:     []string{"description", "type", "default", "type"},
			expectErr: true,
			errMsg:    "duplicate attribute 'type' found in order",
		},
		{
			name:      "incorrect length",
			order:     []string{"description", "type"},
			expectErr: true,
			errMsg:    "provided order length 2 doesn't match expected 6",
		},
		{
			name:      "missing item",
			order:     []string{"description", "type", "default", "sensitive", "nullable"},
			expectErr: true,
			errMsg:    "provided order length 5 doesn't match expected 6",
		},
		{
			name:      "empty order",
			order:     []string{},
			expectErr: true,
			errMsg:    "provided order length 0 doesn't match expected 6",
		},
		{
			name:      "extra items not in default order",
			order:     []string{"description", "type", "default", "sensitive", "nullable", "validation", "unicorn"},
			expectErr: true,
			errMsg:    "provided order length 7 doesn't match expected 6",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := config.IsValidOrder(tc.order)
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				assert.False(t, valid)
			} else {
				require.NoError(t, err)
				assert.True(t, valid)
			}
		})
	}
}

func TestProcessTargetDynamically(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		criteria  []string
		order     []string
		expectErr bool
	}{
		{
			name: "valid target and criteria",
			setupFunc: func(t *testing.T) string {
				return t.TempDir() // Use the testing framework to create a temp directory
			},
			criteria:  []string{"*.tf"},
			order:     []string{"description", "type", "default"},
			expectErr: false,
		},
		{
			name: "non-existent target",
			setupFunc: func(t *testing.T) string {
				// Create and then remove a directory to ensure it's non-existent
				dir := t.TempDir()
				require.NoError(t, os.RemoveAll(dir))
				return dir
			},
			criteria:  []string{"*.tf"},
			order:     []string{"description", "type", "default"},
			expectErr: true,
		},
		{
			name: "invalid criteria",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			criteria:  []string{"invalid_criteria]"},
			order:     []string{"description", "type", "default"},
			expectErr: true,
		},
		{
			name: "empty target string",
			setupFunc: func(t *testing.T) string {
				// Directly return an empty string for the target
				return ""
			},
			criteria:  []string{"*.tf"},
			order:     []string{"description", "type", "default"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.setupFunc(t) // Setup the target using the provided setup function
			err := config.ProcessTargetDynamically(target, tc.criteria, tc.order)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/assert"
)

func TestIsValidOrder(t *testing.T) {
	tests := []struct {
		name      string
		order     []string
		expected  bool
		expectErr bool
	}{
		{
			name:      "valid order with all attributes",
			order:     []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expected:  true,
			expectErr: false,
		},
		{
			name:      "invalid order with duplicates",
			order:     []string{"description", "type", "default", "sensitive", "nullable", "sensitive"},
			expected:  false,
			expectErr: true,
		},
		{
			name:      "invalid order with missing attributes",
			order:     []string{"description", "type", "default"},
			expected:  false,
			expectErr: true,
		},
		{
			name:      "invalid order with extra attributes",
			order:     []string{"description", "type", "default", "sensitive", "nullable", "validation", "extra"},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := config.IsValidOrder(tt.order)
			assert.Equal(t, tt.expected, valid)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTempHCLFile(t *testing.T, tempDir, content string) string {
	t.Helper()

	hclFilePath := filepath.Join(tempDir, "test.tf")
	err := os.WriteFile(hclFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp HCL file: %v", err)
	}

	return hclFilePath
}

func TestProcessTargetDynamically(t *testing.T) {
	tempDir := t.TempDir()
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
			target:      createTempHCLFile(t, tempDir, `variable "test" {}`),
			criteria:    []string{"*.tf"},
			order:       []string{"description", "type", "default", "sensitive", "nullable", "validation"},
			expectError: false,
		},
		{
			name:        "invalid criteria",
			target:      createTempHCLFile(t, tempDir, `variable "test" {}`),
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
			target:      createTempHCLFile(t, tempDir, `variable "test" {}`),
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

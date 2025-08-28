package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "testcli",
		Args: cobra.MinimumNArgs(1),
		RunE: cli.RunE,
	}
	cmd.Flags().StringSlice("criteria", config.DefaultCriteria, "Set the criteria")
	cmd.Flags().StringSlice("order", config.DefaultOrder, "Set the order")
	return cmd
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

func TestRunE_Extended(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, []string)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "Order has a duplicate entry",
			setup: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				hclFilePath := createTempHCLFile(t, tempDir, `resource "test" {}`)
				return hclFilePath, []string{hclFilePath, "--order=description,description"}
			},
			expectError:    true,
			expectedErrMsg: "invalid order: duplicate attribute 'description' found in order",
		},
		{
			name: "Order has a missing entry",
			setup: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				hclFilePath := createTempHCLFile(t, tempDir, `resource "test" {}`)
				return hclFilePath, []string{hclFilePath, "--order=type,default"}
			},
			expectError:    true,
			expectedErrMsg: "invalid order: provided order length 2 doesn't match expected 6",
		},
		{
			name: "Order has an entry which is not allowed",
			setup: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				hclFilePath := createTempHCLFile(t, tempDir, `resource "test" {}`)
				return hclFilePath, []string{hclFilePath, "--order=description,unicorn"}
			},
			expectError:    true,
			expectedErrMsg: "invalid order: provided order length 2 doesn't match expected 6",
		},
		{
			name: "Order is reversed of the DefaultOrder",
			setup: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				hclFilePath := createTempHCLFile(t, tempDir, `resource "test" {}`)
				return hclFilePath, []string{hclFilePath, "--order=validation,nullable,sensitive,default,type,description"}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, args := tt.setup(t)
			cmd := setupTestCommand()
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

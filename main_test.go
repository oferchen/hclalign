package main

import (
	"os"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
)

// Helper function to create a temporary HCL file.
func createTempHCLFile(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", "*.hcl")
	if err != nil {
		t.Fatalf("Failed to create temp HCL file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp HCL file: %v", err)
	}

	return tempFile.Name()
}

func TestMainFunctionality(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Missing Target Argument",
			setup: func(t *testing.T) []string {
				return []string{} // No args means the target is missing
			},
			wantErr: true,
			errMsg:  "accepts 1 arg(s), received 0",
		},
		{
			name: "Valid Single File",
			setup: func(t *testing.T) []string {
				// Create a temp file with valid HCL content
				filePath := createTempHCLFile(t, `variable "test" {}`)
				return []string{filePath}
			},
			wantErr: false,
		},
		{
			name: "Multiple Files",
			setup: func(t *testing.T) []string {
				// Create two temp files, but since the command only accepts one, this should cause an error
				filePath1 := createTempHCLFile(t, `variable "test1" {}`)
				filePath2 := createTempHCLFile(t, `variable "test2" {}`)
				return []string{filePath1, filePath2}
			},
			wantErr: true,
			errMsg:  "accepts 1 arg(s), received 2",
		},
	}

	// Iterate over test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			args := tc.setup(t)

			rootCmd := &cobra.Command{
				Use:   "hcl_align [target file or directory]",
				Short: "Aligns HCL files based on given criteria",
				Args:  cobra.ExactArgs(1),
				RunE:  cli.RunE,
			}
			rootCmd.Flags().StringSliceP("criteria", "c", config.DefaultCriteria, "List of file criteria to align")
			rootCmd.Flags().StringSliceP("order", "o", config.DefaultOrder, "Comma-separated list of the order of variable block fields")

			rootCmd.SetArgs(args)

			_, err := rootCmd.ExecuteC()

			if (err != nil) != tc.wantErr {
				t.Fatalf("Unexpected error status: got error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.wantErr && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Expected error message to contain '%s', but got '%s'", tc.errMsg, err.Error())
			}
		})
	}
}

// main_test.go
package main

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func createTempHCLFile(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", "*.tf")
	if err != nil {
		t.Fatalf("Failed to create temp HCL file: %v", err)
	}
	defer func() { _ = tempFile.Close() }()

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
				return []string{}
			},
			wantErr: true,
			errMsg:  config.ErrMissingTarget,
		},
		{
			name: "Valid Single File",
			setup: func(t *testing.T) []string {

				filePath := createTempHCLFile(t, `variable "test" {}`)
				return []string{filePath}
			},
			wantErr: false,
		},
		{
			name: "Multiple Files",
			setup: func(t *testing.T) []string {

				filePath1 := createTempHCLFile(t, `variable "test1" {}`)
				filePath2 := createTempHCLFile(t, `variable "test2" {}`)
				return []string{filePath1, filePath2}
			},
			wantErr: true,
			errMsg:  "accepts at most 1 arg(s)",
		},
		{
			name: "Mutually Exclusive Flags",
			setup: func(t *testing.T) []string {
				filePath := createTempHCLFile(t, `variable "test" {}`)
				return []string{filePath, "--check", "--diff"}
			},
			wantErr: true,
			errMsg:  "if any flags in the group [write check diff] are set none of the others can be",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			args := tc.setup(t)

			rootCmd := &cobra.Command{
				Use:          "hclalign [target file or directory]",
				Short:        "Aligns HCL files based on given criteria",
				Args:         cobra.ArbitraryArgs,
				RunE:         cli.RunE,
				SilenceUsage: true,
			}
			rootCmd.Flags().Bool("write", false, "write result to file(s)")
			rootCmd.Flags().Bool("check", false, "check if files are formatted")
			rootCmd.Flags().Bool("diff", false, "print the diff of required changes")
			rootCmd.Flags().Bool("stdin", false, "read from STDIN")
			rootCmd.Flags().Bool("stdout", false, "write result to STDOUT")
			rootCmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
			rootCmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
			rootCmd.Flags().StringSlice("order", config.DefaultOrder, "order of variable block fields")
			rootCmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
			rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
			rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
			rootCmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
			rootCmd.MarkFlagsMutuallyExclusive("write", "check", "diff")

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

func TestCLIOrderFlagInfluencesProcessing(t *testing.T) {
	content := `variable "test" {
  description = "d"
  default     = "v"
}`
	filePath := createTempHCLFile(t, content)

	rootCmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Short:        "Aligns HCL files based on given criteria",
		Args:         cobra.ArbitraryArgs,
		RunE:         cli.RunE,
		SilenceUsage: true,
	}
	rootCmd.Flags().Bool("write", false, "write result to file(s)")
	rootCmd.Flags().Bool("check", false, "check if files are formatted")
	rootCmd.Flags().Bool("diff", false, "print the diff of required changes")
	rootCmd.Flags().Bool("stdin", false, "read from STDIN")
	rootCmd.Flags().Bool("stdout", false, "write result to STDOUT")
	rootCmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
	rootCmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
	rootCmd.Flags().StringSlice("order", config.DefaultOrder, "order of variable block fields")
	rootCmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
	rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	rootCmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	rootCmd.MarkFlagsMutuallyExclusive("write", "check", "diff")

	rootCmd.SetArgs([]string{filePath, "--order=default", "--order=description"})

	_, err := rootCmd.ExecuteC()
	require.NoError(t, err)

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	expected := "variable \"test\" {\n  default     = \"v\"\n  description = \"d\"\n}"
	require.Equal(t, expected, string(data))
}

func TestCLIStrictOrderUnknownAttribute(t *testing.T) {
	filePath := createTempHCLFile(t, `variable "test" {}`)

	rootCmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Short:        "Aligns HCL files based on given criteria",
		Args:         cobra.ArbitraryArgs,
		RunE:         cli.RunE,
		SilenceUsage: true,
	}
	rootCmd.Flags().Bool("write", false, "write result to file(s)")
	rootCmd.Flags().Bool("check", false, "check if files are formatted")
	rootCmd.Flags().Bool("diff", false, "print the diff of required changes")
	rootCmd.Flags().Bool("stdin", false, "read from STDIN")
	rootCmd.Flags().Bool("stdout", false, "write result to STDOUT")
	rootCmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
	rootCmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
	rootCmd.Flags().StringSlice("order", config.DefaultOrder, "order of variable block fields")
	rootCmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
	rootCmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	rootCmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	rootCmd.MarkFlagsMutuallyExclusive("write", "check", "diff")

	rootCmd.SetArgs([]string{filePath, "--order=description", "--order=unknown", "--strict-order"})

	_, err := rootCmd.ExecuteC()
	require.Error(t, err)
}

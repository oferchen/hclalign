package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Args:         cobra.ArbitraryArgs,
		RunE:         RunE,
		SilenceUsage: true,
	}
	cmd.Flags().Bool("write", false, "write result to file(s)")
	cmd.Flags().Bool("check", false, "check if files are formatted")
	cmd.Flags().Bool("diff", false, "print the diff of required changes")
	cmd.Flags().Bool("stdin", false, "read from STDIN")
	cmd.Flags().Bool("stdout", false, "write result to STDOUT")
	cmd.Flags().StringSlice("include", config.DefaultInclude, "glob patterns to include")
	cmd.Flags().StringSlice("exclude", config.DefaultExclude, "glob patterns to exclude")
	cmd.Flags().StringSlice("order", config.CanonicalOrder, "order of variable block fields")
	cmd.Flags().Bool("strict-order", false, "enforce strict attribute ordering")
	cmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	cmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	cmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	return cmd
}

func TestRunEUsageError(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEFormattingNeeded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.tf")
	content := "variable \"a\" {\n  type = string\n  description = \"d\"\n}"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cmd := newRootCmd()
	cmd.SetArgs([]string{path, "--check"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.Code)
}

func TestRunERuntimeError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.tf")
	// Invalid HCL to trigger processing error
	require.NoError(t, os.WriteFile(path, []byte("variable \"a\" {"), 0o644))

	cmd := newRootCmd()
	cmd.SetArgs([]string{path})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 3, exitErr.Code)
}

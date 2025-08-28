// cli/cli_test.go
package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)


func newRootCmd(exclusive bool) *cobra.Command {
=======
func newRootCmd() *cobra.Command { return newTestRootCmd(true) }

func newTestRootCmd(exclusive bool) *cobra.Command {

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
	if exclusive {
		cmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	}
	return cmd
}

func TestRunEUsageError(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunETargetWithStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.tf")

	cmd := newRootCmd(true)
	cmd.SetArgs([]string{path, "--stdin"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEMultipleModeFlags(t *testing.T) {
	cmd := newRootCmd(false)
	cmd.SetArgs([]string{"--check", "--diff"})
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

	cmd := newRootCmd(true)
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
	invalidHCL := []byte("variable \"a\" {")
	require.NoError(t, os.WriteFile(path, invalidHCL, 0o644))

	cmd := newRootCmd(true)
	cmd.SetArgs([]string{path})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 3, exitErr.Code)
}

func TestRunEInvalidConcurrency(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--stdin", "--concurrency", "0"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEModes(t *testing.T) {
	unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"
	formatted := "variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n"

	tests := []struct {
		name       string
		flags      []string
		stdin      string
		wantCode   int
		wantStdout string
		contains   string
		withFile   bool
		wantFile   string
	}{
		{
			name:     "diff",
			flags:    []string{"--diff"},
			wantCode: 1,
			contains: "@@",
			withFile: true,
			wantFile: unformatted,
		},
		{
			name:       "stdin",
			flags:      []string{"--stdin", "--stdout"},
			stdin:      unformatted,
			wantCode:   0,
			wantStdout: formatted,
		},
		{
			name:       "stdout",
			flags:      []string{"--stdout"},
			wantCode:   0,
			wantStdout: formatted,
			withFile:   true,
			wantFile:   formatted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []string
			var filePath string
			if tt.withFile {
				dir := t.TempDir()
				filePath = filepath.Join(dir, "test.tf")
				require.NoError(t, os.WriteFile(filePath, []byte(unformatted), 0o644))
				args = append([]string{filePath}, tt.flags...)
			} else {
				args = tt.flags
			}

			// capture stdout
			oldStdout := os.Stdout
			rOut, wOut, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = wOut
			t.Cleanup(func() { os.Stdout = oldStdout })
			outChan := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, rOut)
				outChan <- buf.String()
			}()

			// set stdin if needed
			if tt.stdin != "" {
				oldStdin := os.Stdin
				rIn, wIn, err := os.Pipe()
				require.NoError(t, err)
				os.Stdin = rIn
				t.Cleanup(func() { os.Stdin = oldStdin })
				_, err = wIn.Write([]byte(tt.stdin))
				require.NoError(t, err)
				wIn.Close()
			}

			cmd := newRootCmd(true)
			cmd.SetArgs(args)
			_, err = cmd.ExecuteC()

			if tt.wantCode == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				var exitErr *ExitCodeError
				require.ErrorAs(t, err, &exitErr)
				require.Equal(t, tt.wantCode, exitErr.Code)
			}

			wOut.Close()
			stdout := <-outChan

			if tt.wantStdout != "" {
				require.Equal(t, tt.wantStdout, stdout)
			}
			if tt.contains != "" {
				require.Contains(t, stdout, tt.contains)
			}

			if tt.withFile {
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				if tt.wantFile != "" {
					require.Equal(t, tt.wantFile, string(data))
				}
			}
		})
	}
}

func TestRunEMultipleModeFlags(t *testing.T) {
	cmd := newTestRootCmd(false)
	cmd.SetArgs([]string{"--write", "--check"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
	require.Contains(t, exitErr.Error(), "cannot specify more than one")
}

func TestRunEInvalidConcurrency(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--concurrency", "0", "target.tf"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
	require.Contains(t, exitErr.Error(), "concurrency must be at least 1")
}

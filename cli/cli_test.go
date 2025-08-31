// cli/cli_test.go
package cli

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newRootCmd(exclusive bool) *cobra.Command { return newTestRootCmd(exclusive) }

func newTestRootCmd(exclusive bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Args:         cobra.MaximumNArgs(1),
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
	cmd.Flags().String("providers-schema", "", "path to providers schema file")
	cmd.Flags().Bool("use-terraform-schema", false, "use terraform schema for providers")
	cmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	cmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	cmd.Flags().StringSlice("types", []string{"variable"}, "comma-separated list of block types to align")
	cmd.Flags().Bool("all", false, "align all block types")
	cmd.MarkFlagsMutuallyExclusive("types", "all")
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

func TestRunETooManyArgs(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"one", "two"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	require.Contains(t, err.Error(), "accepts at most 1 arg(s)")
}

func TestRunETargetWithStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.tf")

	cmd := newRootCmd(true)
	cmd.SetArgs([]string{path, "--stdin", "--stdout"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEStdinRequiresStdout(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--stdin"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEMutuallyExclusiveFlags(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--check", "--diff"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
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

func TestRunEStdinRuntimeError(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--stdin", "--stdout"})

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	_, err = w.Write([]byte("variable \"a\" {"))
	require.NoError(t, err)
	w.Close()

	_, err = cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 3, exitErr.Code)
}

func TestRunEInvalidConcurrency(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--stdin", "--stdout", "--concurrency", "0"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestRunEInvalidGlob(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		message string
	}{
		{name: "include", flag: "--include", message: "invalid include"},
		{name: "exclude", flag: "--exclude", message: "invalid exclude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd(true)
			cmd.SetArgs([]string{"--stdin", "--stdout", tt.flag, "["})
			_, err := cmd.ExecuteC()
			require.Error(t, err)
			var exitErr *ExitCodeError
			require.ErrorAs(t, err, &exitErr)
			require.Equal(t, 2, exitErr.Code)
			require.Contains(t, exitErr.Error(), tt.message)
		})
	}
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
			name:     "write",
			flags:    []string{"--write"},
			wantCode: 0,
			withFile: true,
			wantFile: formatted,
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
				expected := tt.wantStdout
				if tt.withFile {
					expected = fmt.Sprintf("\n--- %s ---\n%s", filePath, tt.wantStdout)
				}
				require.Equal(t, expected, stdout)
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

func TestRunEVerbose(t *testing.T) {
	unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"

	tests := []struct {
		name    string
		verbose bool
		wantLog bool
	}{
		{name: "verbose", verbose: true, wantLog: true},
		{name: "silent", verbose: false, wantLog: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.tf")
			require.NoError(t, os.WriteFile(path, []byte(unformatted), 0o644))

			cmd := newRootCmd(true)
			args := []string{path}
			if tt.verbose {
				args = append(args, "--verbose")
			}
			cmd.SetArgs(args)

			oldOut := log.Writer()
			r, w, err := os.Pipe()
			require.NoError(t, err)
			log.SetOutput(w)
			t.Cleanup(func() { log.SetOutput(oldOut) })
			logCh := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				logCh <- buf.String()
			}()

			_, err = cmd.ExecuteC()
			require.NoError(t, err)

			w.Close()
			got := <-logCh
			if tt.wantLog {
				require.Contains(t, got, "processed file: ")
			} else {
				require.Empty(t, got)
			}
		})
	}
}

func TestRunEFollowSymlinks(t *testing.T) {
	unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"
	formatted := "variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n"

	tests := []struct {
		name   string
		follow bool
		want   string
	}{
		{name: "follow", follow: true, want: formatted},
		{name: "no_follow", follow: false, want: unformatted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			target := t.TempDir()
			realFile := filepath.Join(target, "file.tf")
			require.NoError(t, os.WriteFile(realFile, []byte(unformatted), 0o644))
			link := filepath.Join(base, "link")
			require.NoError(t, os.Symlink(target, link))

			cmd := newRootCmd(true)
			args := []string{base}
			if tt.follow {
				args = append(args, "--follow-symlinks")
			}
			cmd.SetArgs(args)

			_, err := cmd.ExecuteC()
			require.NoError(t, err)

			data, err := os.ReadFile(realFile)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(data))
		})
	}
}

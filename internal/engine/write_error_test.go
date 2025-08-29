// internal/engine/write_error_test.go
package engine_test

import (
	"context"
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	enginepkg "github.com/oferchen/hclalign/internal/engine"
	terraformfmt "github.com/oferchen/hclalign/internal/fmt"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func newRootCmd(exclusive bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "hclalign [target file or directory]",
		Args:         cobra.MaximumNArgs(1),
		RunE:         cli.RunE,
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
	cmd.Flags().String("fmt-strategy", string(terraformfmt.StrategyAuto), "formatter strategy: auto, binary, go")
	cmd.Flags().Bool("fmt-only", false, "run formatter only")
	cmd.Flags().Bool("no-fmt", false, "disable formatter")
	if exclusive {
		cmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	}
	cmd.MarkFlagsMutuallyExclusive("fmt-only", "no-fmt")
	return cmd
}

func TestProcessWriteFileError(t *testing.T) {
	dir := t.TempDir()
	casesDir := filepath.Join("..", "..", "tests", "cases")
	data, err := os.ReadFile(filepath.Join(casesDir, "simple", "in.tf"))
	require.NoError(t, err)
	filePath := filepath.Join(dir, "example.tf")
	require.NoError(t, os.WriteFile(filePath, data, 0o644))

	original := enginepkg.WriteFileAtomic
	enginepkg.WriteFileAtomic = func(ctx context.Context, path string, b []byte, perm iofs.FileMode, hints internalfs.Hints) error {
		return errors.New("write error")
	}
	defer func() { enginepkg.WriteFileAtomic = original }()

	cmd := newRootCmd(true)
	cmd.SetArgs([]string{filePath, "--write"})
	_, err = cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *cli.ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 3, exitErr.Code)

	got, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, string(data), string(got))
}

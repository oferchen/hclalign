// filename: internal/engine/write_error_test.go
package engine_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"

	"github.com/oferchen/hclalign/cli"
	"github.com/oferchen/hclalign/config"
	enginepkg "github.com/oferchen/hclalign/internal/engine"
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
	cmd.Flags().String("providers-schema", "", "path to providers schema file")
	cmd.Flags().Bool("use-terraform-schema", false, "use terraform schema for providers")
	cmd.Flags().Int("concurrency", runtime.GOMAXPROCS(0), "maximum concurrency")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose logging")
	cmd.Flags().Bool("follow-symlinks", false, "follow symlinks when traversing directories")
	cmd.Flags().StringSlice("types", []string{"variable"}, "comma-separated list of block types to align")
	cmd.Flags().Bool("all", false, "align all block types")
	cmd.Flags().Bool("prefix-order", false, "lexicographically sort unknown attributes")
	cmd.MarkFlagsMutuallyExclusive("types", "all")
	if exclusive {
		cmd.MarkFlagsMutuallyExclusive("write", "check", "diff")
	}
	return cmd
}

func TestProcessWriteFileError(t *testing.T) {
	dir := t.TempDir()
	casesDir := filepath.Join("..", "..", "tests", "cases")
	data, err := os.ReadFile(filepath.Join(casesDir, "variable", "in.tf"))
	require.NoError(t, err)
	filePath := filepath.Join(dir, "example.tf")
	require.NoError(t, os.WriteFile(filePath, data, 0o644))

	original := enginepkg.WriteFileAtomic
	enginepkg.WriteFileAtomic = func(ctx context.Context, _ internalfs.WriteOpts) error {
		return errors.New("write error")
	}
	defer func() { enginepkg.WriteFileAtomic = original }()

	cmd := newRootCmd(true)
	cmd.SetArgs([]string{filePath, "--write", "--prefix-order"})
	_, err = cmd.ExecuteC()
	require.Error(t, err)
	var exitErr *cli.ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 3, exitErr.Code)

	got, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, string(data), string(got))
}

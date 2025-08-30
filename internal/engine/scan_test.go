// internal/engine/scan_test.go
package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestScanDefaultExcludeDirectories(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("..", "..", "tests", "cases", "default_excludes")
	gitIgnored := filepath.Join(dir, ".git", "ignored.tf")
	require.NoError(t, os.MkdirAll(filepath.Dir(gitIgnored), 0o755))
	require.NoError(t, os.WriteFile(gitIgnored, []byte(""), 0o644))
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(gitIgnored)) })

	cfg := &config.Config{
		Target:  dir,
		Include: config.DefaultInclude,
		Exclude: config.DefaultExclude,
	}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(dir, "main.tf")}, files)
}

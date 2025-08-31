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

	cfg := &config.Config{
		Target:  dir,
		Include: config.DefaultInclude,
		Exclude: config.DefaultExclude,
	}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{
		filepath.Join(dir, "main.tf"),
		filepath.Join(dir, "nested", ".terraform", "ignored.tf"),
	}, files)
}

func TestScanFollowSymlinksSelfCycle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(""), 0o644))
	require.NoError(t, os.Symlink(".", filepath.Join(dir, "self")))

	cfg := &config.Config{
		Target:         dir,
		Include:        config.DefaultInclude,
		Exclude:        config.DefaultExclude,
		FollowSymlinks: true,
	}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(dir, "main.tf")}, files)
}

func TestScanFollowSymlinksCycle(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	require.NoError(t, os.Mkdir(a, 0o755))
	require.NoError(t, os.Mkdir(b, 0o755))
	fileA := filepath.Join(a, "a.tf")
	fileB := filepath.Join(b, "b.tf")
	require.NoError(t, os.WriteFile(fileA, []byte(""), 0o644))
	require.NoError(t, os.WriteFile(fileB, []byte(""), 0o644))
	require.NoError(t, os.Symlink(b, filepath.Join(a, "link")))
	require.NoError(t, os.Symlink(a, filepath.Join(b, "link")))

	cfg := &config.Config{
		Target:         root,
		Include:        config.DefaultInclude,
		Exclude:        config.DefaultExclude,
		FollowSymlinks: true,
	}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{fileA, filepath.Join(a, "link", "b.tf")}, files)
}

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
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)
	dir = absDir

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
		filepath.Join(dir, "nested", "vendor", "included.tf"),
	}, files)
}

func TestScanFollowSymlinks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.tf")
	require.NoError(t, os.WriteFile(target, []byte(""), 0o644))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, "link.tf")))

	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside.tf")
	require.NoError(t, os.WriteFile(outside, []byte(""), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "out.tf")))

	cfg := &config.Config{Target: dir, Include: config.DefaultInclude}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{target}, files)

	cfg.FollowSymlinks = true
	files, err = scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{target, target}, files)
}

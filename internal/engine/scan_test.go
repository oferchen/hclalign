package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestScanDefaultExcludeDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// root tf file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(""), 0o644))

	// create default excluded directories each with a tf file
	excluded := []string{".terraform", "vendor", ".git", "node_modules"}
	for _, d := range excluded {
		sub := filepath.Join(dir, d)
		require.NoError(t, os.MkdirAll(sub, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(sub, "ignored.tf"), []byte(""), 0o644))
	}

	cfg := &config.Config{
		Target:  dir,
		Include: config.DefaultInclude,
		Exclude: config.DefaultExclude,
	}

	files, err := scan(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(dir, "main.tf")}, files)
}

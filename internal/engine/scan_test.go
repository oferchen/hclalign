// filename: internal/engine/scan_test.go
package engine

import (
	"context"
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
	require.Equal(t, []string{filepath.Join(dir, "main.tf")}, files)
}

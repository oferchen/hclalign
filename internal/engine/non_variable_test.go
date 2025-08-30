// filename: internal/engine/non_variable_test.go
package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestNonVariableBlocksUnchanged(t *testing.T) {
	t.Parallel()

	casesDir := filepath.Join("..", "..", "tests", "cases", "non_variable")
	inPath := filepath.Join(casesDir, "in.tf")
	inBytes, err := os.ReadFile(inPath)
	require.NoError(t, err)

	dir := t.TempDir()
	file := filepath.Join(dir, "test.tf")
	require.NoError(t, os.WriteFile(file, inBytes, 0o644))

	cfg := &config.Config{
		Target:      file,
		Include:     []string{"**/*.tf"},
		Concurrency: 1,
		Types:       []string{"variable"},
	}

	for i := 0; i < 2; i++ {
		changed, err := Process(context.Background(), cfg)
		require.NoError(t, err)
		require.False(t, changed)

		got, err := os.ReadFile(file)
		require.NoError(t, err)
		require.Equal(t, string(inBytes), string(got))
	}
}

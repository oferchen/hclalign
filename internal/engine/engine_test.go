package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestProcessMissingTarget(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Target:      "nonexistent.hcl",
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
	require.False(t, changed)
}

func TestProcessContextCancelled(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("testdata", "idempotent_input.hcl"))
	require.NoError(t, err)
	filePath := filepath.Join(dir, "example.hcl")
	require.NoError(t, os.WriteFile(filePath, data, 0o644))

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	testHookAfterParse = func() { cancel() }
	defer func() { testHookAfterParse = nil }()

	changed, err := Process(ctx, cfg)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, changed)
}

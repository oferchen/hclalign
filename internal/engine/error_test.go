// internal/engine/error_test.go
package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestProcessInvalidHCLFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.hcl")
	orig := "variable \"a\" {"
	require.NoError(t, os.WriteFile(path, []byte(orig), 0o644))

	cfg := &config.Config{
		Target:      path,
		Include:     []string{"**/*.hcl"},
		Concurrency: 1,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing error")
	require.False(t, changed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, orig, string(data))
}

func TestProcessReaderMalformed(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("variable \"a\" {")
	var buf bytes.Buffer
	cfg := &config.Config{Stdout: true}

	changed, err := ProcessReader(context.Background(), r, &buf, cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing error")
	require.False(t, changed)
	require.Empty(t, buf.String())
}

func TestProcessReaderEmpty(t *testing.T) {
	t.Parallel()

	r := strings.NewReader("")
	var buf bytes.Buffer
	cfg := &config.Config{Stdout: true}

	changed, err := ProcessReader(context.Background(), r, &buf, cfg)
	require.NoError(t, err)
	require.False(t, changed)
	require.Empty(t, buf.String())
}

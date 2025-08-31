// internal/fmt/formatfile_test.go
package terraformfmt

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatFileUsesCLI(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "terraform")
	script := []byte("#!/bin/sh\nout=$5\nprintf 'cli\\n' > \"$out\"")
	require.NoError(t, os.WriteFile(bin, script, 0o755))
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", dir)
	f := filepath.Join(dir, "file.tf")
	require.NoError(t, os.WriteFile(f, []byte("orig\n"), 0o644))
	ran, err := FormatFile(context.Background(), f)
	require.NoError(t, err)
	require.True(t, ran)
	out, err := os.ReadFile(f)
	require.NoError(t, err)
	require.Equal(t, "cli\n", string(out))
}

func TestFormatFileMissingCLI(t *testing.T) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "")
	ran, err := FormatFile(context.Background(), "file.tf")
	require.NoError(t, err)
	require.False(t, ran)
}

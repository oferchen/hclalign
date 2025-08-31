// internal/engine/error_test.go
package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestProcessInvalidHCLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.tf")
	orig := "variable \"a\" {\n"
	require.NoError(t, os.WriteFile(path, []byte(orig), 0o644))

	cfg := &config.Config{
		Target:      path,
		Include:     []string{"**/*.tf"},
		Concurrency: 1,
	}

	origFmt := terraformFmtFormatFile
	terraformFmtFormatFile = func(ctx context.Context, p string) (bool, error) {
		return false, fmt.Errorf("boom")
	}
	defer func() { terraformFmtFormatFile = origFmt }()

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("parsing error in file %s", path))
	require.False(t, changed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, orig, string(data))
}

func TestProcessAggregatesErrors(t *testing.T) {
	dir := t.TempDir()
	bad1 := filepath.Join(dir, "bad1.tf")
	bad2 := filepath.Join(dir, "bad2.tf")
	goodPath := filepath.Join(dir, "good.tf")

	bad := "variable \"bad\" {"
	require.NoError(t, os.WriteFile(bad1, []byte(bad), 0o644))
	require.NoError(t, os.WriteFile(bad2, []byte(bad), 0o644))

	good := "variable \"good\" {\n  description = \"foo\"\n  default     = 1\n}\n"
	require.NoError(t, os.WriteFile(goodPath, []byte(good), 0o644))

	cfg := &config.Config{
		Target:      dir,
		Include:     []string{"**/*.tf"},
		Concurrency: 2,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.False(t, changed)
	require.Contains(t, err.Error(), "bad1.tf")
	require.Contains(t, err.Error(), "bad2.tf")

	data, readErr := os.ReadFile(goodPath)
	require.NoError(t, readErr)
	require.Equal(t, good, string(data))
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

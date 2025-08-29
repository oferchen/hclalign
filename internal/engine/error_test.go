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
		Target:		path,
		Include:	[]string{"**/*.hcl"},
		Concurrency:	1,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing error")
	require.False(t, changed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, orig, string(data))
}

func TestProcessFilesAggregatesErrors(t *testing.T) {
	dir := t.TempDir()
	goodPath := filepath.Join(dir, "good.hcl")
	bad1Path := filepath.Join(dir, "bad1.hcl")
	bad2Path := filepath.Join(dir, "bad2.hcl")

	good := "variable \"good\" {\n  default     = 1\n  description = \"foo\"\n}\n"
	require.NoError(t, os.WriteFile(goodPath, []byte(good), 0o644))

	bad1 := "variable \"bad1\" {"
	bad2 := "variable \"bad2\" { default = [ }"
	require.NoError(t, os.WriteFile(bad1Path, []byte(bad1), 0o644))
	require.NoError(t, os.WriteFile(bad2Path, []byte(bad2), 0o644))

	cfg := &config.Config{
		Target:		dir,
		Include:	[]string{"**/*.hcl"},
		Concurrency:	2,
	}

	changed, err := Process(context.Background(), cfg)
	require.Error(t, err)
	require.True(t, changed)
	require.Contains(t, err.Error(), "bad1.hcl")
	require.Contains(t, err.Error(), "bad2.hcl")

	data, readErr := os.ReadFile(goodPath)
	require.NoError(t, readErr)
	expected := "variable \"good\" {\n  description = \"foo\"\n  default     = 1\n}\n"
	require.Equal(t, expected, string(data))
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


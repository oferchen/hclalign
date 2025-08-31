// internal/engine/pipeline_test.go
package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestProcessFileTerraformFmt(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.tf")
	require.NoError(t, os.WriteFile(file, []byte("variable \"a\" { type = string }"), 0o644))

	var formatCalls, runCalls int
	origFormat := terraformFmtFormatFile
	origRun := terraformFmtRun
	terraformFmtFormatFile = func(ctx context.Context, path string) ([]byte, internalfs.Hints, bool, error) {
		formatCalls++
		return nil, internalfs.Hints{}, false, nil
	}
	terraformFmtRun = func(ctx context.Context, b []byte) ([]byte, internalfs.Hints, error) {
		runCalls++
		return b, internalfs.Hints{}, nil
	}
	t.Cleanup(func() {
		terraformFmtFormatFile = origFormat
		terraformFmtRun = origRun
	})

	p := &Processor{cfg: &config.Config{Mode: config.ModeWrite}}
	_, _, err := p.processFile(context.Background(), file)
	require.NoError(t, err)
	require.Equal(t, 1, formatCalls)
	require.Equal(t, 2, runCalls)
}

func TestProcessFileTerraformFmtPreservesCRLF(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.tf")
	hints := internalfs.Hints{Newline: "\r\n"}
	content := internalfs.ApplyHints([]byte("variable \"a\" {type=string}\n"), hints)
	require.NoError(t, os.WriteFile(file, content, 0o644))

	origFormat := terraformFmtFormatFile
	origRun := terraformFmtRun
	terraformFmtFormatFile = func(ctx context.Context, path string) ([]byte, internalfs.Hints, bool, error) {
		formatted := []byte("variable \"a\" {\n  type = string\n}\n")
		return formatted, internalfs.Hints{}, true, nil
	}
	terraformFmtRun = func(ctx context.Context, b []byte) ([]byte, internalfs.Hints, error) {
		return b, internalfs.Hints{}, nil
	}
	t.Cleanup(func() {
		terraformFmtFormatFile = origFormat
		terraformFmtRun = origRun
	})

	p := &Processor{cfg: &config.Config{Mode: config.ModeWrite}}
	changed, _, err := p.processFile(context.Background(), file)
	require.NoError(t, err)
	require.True(t, changed)

	out, err := os.ReadFile(file)
	require.NoError(t, err)
	h := internalfs.DetectHintsFromBytes(out)
	require.Equal(t, "\r\n", h.Newline)
	require.False(t, h.HasBOM)
}

func TestProcessFileTerraformFmtPreservesBOM(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.tf")
	hints := internalfs.Hints{HasBOM: true, Newline: "\r\n"}
	content := internalfs.ApplyHints([]byte("variable \"a\" {type=string}\n"), hints)
	require.NoError(t, os.WriteFile(file, content, 0o644))

	origFormat := terraformFmtFormatFile
	origRun := terraformFmtRun
	terraformFmtFormatFile = func(ctx context.Context, path string) ([]byte, internalfs.Hints, bool, error) {
		formatted := []byte("variable \"a\" {\n  type = string\n}\n")
		return formatted, internalfs.Hints{}, true, nil
	}
	terraformFmtRun = func(ctx context.Context, b []byte) ([]byte, internalfs.Hints, error) {
		return b, internalfs.Hints{}, nil
	}
	t.Cleanup(func() {
		terraformFmtFormatFile = origFormat
		terraformFmtRun = origRun
	})

	p := &Processor{cfg: &config.Config{Mode: config.ModeWrite}}
	changed, _, err := p.processFile(context.Background(), file)
	require.NoError(t, err)
	require.True(t, changed)

	out, err := os.ReadFile(file)
	require.NoError(t, err)
	h := internalfs.DetectHintsFromBytes(out)
	require.Equal(t, "\r\n", h.Newline)
	require.True(t, h.HasBOM)
}

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
	terraformFmtFormatFile = func(ctx context.Context, path string) (bool, error) {
		formatCalls++
		return false, nil
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

func TestProcessFileSkipTerraformFmt(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.tf")
	require.NoError(t, os.WriteFile(file, []byte("variable \"a\" { type = string }"), 0o644))

	var formatCalls, runCalls int
	origFormat := terraformFmtFormatFile
	origRun := terraformFmtRun
	terraformFmtFormatFile = func(ctx context.Context, path string) (bool, error) {
		formatCalls++
		return false, nil
	}
	terraformFmtRun = func(ctx context.Context, b []byte) ([]byte, internalfs.Hints, error) {
		runCalls++
		return b, internalfs.Hints{}, nil
	}
	t.Cleanup(func() {
		terraformFmtFormatFile = origFormat
		terraformFmtRun = origRun
	})

	p := &Processor{cfg: &config.Config{Mode: config.ModeWrite, SkipTerraformFmt: true}}
	_, _, err := p.processFile(context.Background(), file)
	require.NoError(t, err)
	require.Equal(t, 0, formatCalls)
	require.Equal(t, 0, runCalls)
}

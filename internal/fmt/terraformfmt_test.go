// internal/fmt/terraformfmt_test.go
package terraformfmt

import (
        "bytes"
        "context"
        "os"
        "os/exec"
        "path/filepath"
        "testing"

        internalfs "github.com/oferchen/hclalign/internal/fs"
        "github.com/oferchen/hclalign/formatter"
        "github.com/stretchr/testify/require"
)

func TestRunUsesTerraformCLI(t *testing.T) {
        dir := t.TempDir()
        bin := filepath.Join(dir, "terraform")
        script := []byte("#!/bin/sh\ncat >/dev/null\nprintf 'bin\\n'")
        if err := os.WriteFile(bin, script, 0o755); err != nil {
                t.Fatalf("write fake terraform: %v", err)
        }
        oldPath := os.Getenv("PATH")
        defer os.Setenv("PATH", oldPath)
        os.Setenv("PATH", dir)
        out, hints, err := Run(context.Background(), []byte("input\n"))
        require.NoError(t, err)
        require.Equal(t, "bin\n", string(out))
        require.Equal(t, internalfs.Hints{Newline: "\n"}, hints)
}

func TestRunFallsBackToGoFormatter(t *testing.T) {
        oldPath := os.Getenv("PATH")
        defer os.Setenv("PATH", oldPath)
        os.Setenv("PATH", "")
        src := []byte("variable \"a\" {\n  type = string\n}\n")
        want, wantHints, err := formatter.Format(src, "test.tf")
        require.NoError(t, err)
        got, gotHints, err := Run(context.Background(), src)
        require.NoError(t, err)
        require.Equal(t, want, got)
        require.Equal(t, wantHints, gotHints)
}

func TestRunPropagatesHints(t *testing.T) {
        src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {}\r\n")...)
        formatted, hints, err := Run(context.Background(), src)
        require.NoError(t, err)
        require.True(t, hints.HasBOM)
        require.Equal(t, "\r\n", hints.Newline)
        require.False(t, bytes.HasPrefix(formatted, []byte{0xef, 0xbb, 0xbf}))
        require.False(t, bytes.Contains(formatted, []byte("\r\n")))
        styled := internalfs.ApplyHints(formatted, hints)
        require.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {}\r\n")...), styled)
}

func TestRunContextCanceled(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        cancel()
        _, _, err := Run(ctx, []byte("variable \"a\" {}\n"))
        require.ErrorIs(t, err, context.Canceled)
}

func TestGoMatchesBinary(t *testing.T) {
        if _, err := exec.LookPath("terraform"); err != nil {
                t.Skip("terraform binary not found")
        }
        src := []byte("variable \"a\" {\n  type = string\n}\n")
        goFmt, _, err := Format(src, "test.tf", string(StrategyGo))
        require.NoError(t, err)
        binFmt, _, err := Format(src, "test.tf", string(StrategyBinary))
        require.NoError(t, err)
        require.Equal(t, goFmt, binFmt)
}

func TestIdempotent(t *testing.T) {
        src := []byte("variable \"a\" {\n  type = string\n}\n")
        first, _, err := Format(src, "test.tf", string(StrategyGo))
        require.NoError(t, err)
        second, _, err := Format(first, "test.tf", string(StrategyGo))
        require.NoError(t, err)
        require.Equal(t, first, second)
}

func TestBinaryPreservesHints(t *testing.T) {
        if _, err := exec.LookPath("terraform"); err != nil {
                t.Skip("terraform binary not found")
        }
        src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {\r\n  type = string\r\n}\r\n")...)
        formatted, hints, err := Format(src, "test.tf", string(StrategyBinary))
        require.NoError(t, err)
        require.True(t, hints.HasBOM)
        require.Equal(t, "\r\n", hints.Newline)
        require.False(t, bytes.HasPrefix(formatted, []byte{0xef, 0xbb, 0xbf}))
        require.False(t, bytes.Contains(formatted, []byte("\r\n")))
        styled := internalfs.ApplyHints(formatted, hints)
        require.Equal(t, src, styled)
}

func TestUnknownStrategy(t *testing.T) {
        _, _, err := Format([]byte("{}"), "test.tf", "bogus")
        require.Error(t, err)
}

func TestBinaryInvalidUTF8(t *testing.T) {
        _, _, err := Format([]byte{0xff, 0xfe}, "test.tf", string(StrategyBinary))
        require.Error(t, err)
}

func TestBinaryMissingTerraform(t *testing.T) {
        oldPath := os.Getenv("PATH")
        defer os.Setenv("PATH", oldPath)
        os.Setenv("PATH", "")
        _, _, err := Format([]byte("variable \"a\" {}\n"), "test.tf", string(StrategyBinary))
        require.Error(t, err)
}

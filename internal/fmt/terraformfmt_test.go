// filename: internal/fmt/terraformfmt_test.go
package terraformfmt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestGoMatchesBinary(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	goFmt, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	binFmt, err := Format(src, "test.tf", string(StrategyBinary))
	require.NoError(t, err)
	require.Equal(t, string(binFmt), string(goFmt))
}

func TestAutoPrefersTerraformBinary(t *testing.T) {
	tmpDir := t.TempDir()
	argsOut := filepath.Join(tmpDir, "args.txt")
	script := filepath.Join(tmpDir, "terraform")
	content := "#!/bin/sh\n" + "cat -\n" + "echo \"$@\" > \"$MOCK_TF_ARGS_OUT\"\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	oldPath := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath))
	defer os.Setenv("PATH", oldPath)
	require.NoError(t, os.Setenv("MOCK_TF_ARGS_OUT", argsOut))
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	formatted, err := Format(src, "test.tf", string(StrategyAuto))
	require.NoError(t, err)
	require.Equal(t, string(src), string(formatted))
	argsBytes, err := os.ReadFile(argsOut)
	require.NoError(t, err)
	fields := strings.Fields(string(argsBytes))
	require.Greater(t, len(fields), 0)
	require.Equal(t, "-", fields[len(fields)-1])
}

func TestIdempotent(t *testing.T) {
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	first, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	second, err := Format(first, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
}

func TestBinaryPreservesHints(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {\r\n  type = string\r\n}\r\n")...)
	formatted, err := Format(src, "test.tf", string(StrategyBinary))
	require.NoError(t, err)
	hints := internalfs.DetectHintsFromBytes(formatted)
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
}

func TestRunPreservesHints(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {\r\n  type = string\r\n}\r\n")...)
	formatted, err := Run(context.Background(), src)
	require.NoError(t, err)
	hints := internalfs.DetectHintsFromBytes(formatted)
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
}

func TestUnknownStrategy(t *testing.T) {
	_, err := Format([]byte("{}"), "test.tf", "bogus")
	require.Error(t, err)
}

func TestBinaryInvalidUTF8(t *testing.T) {
	_, err := Format([]byte{0xff, 0xfe}, "test.tf", string(StrategyBinary))
	require.Error(t, err)
}

func TestBinaryMissingTerraform(t *testing.T) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "")
	_, err := Format([]byte("variable \"a\" {}\n"), "test.tf", string(StrategyBinary))
	require.Error(t, err)
}

func TestRunMissingTerraform(t *testing.T) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "")
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	formatted, err := Run(context.Background(), src)
	require.NoError(t, err)
	goFmt, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(goFmt), string(formatted))
}

func TestAutoFallsBackToGo(t *testing.T) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "")
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	autoFmt, err := Format(src, "test.tf", string(StrategyAuto))
	require.NoError(t, err)
	goFmt, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(goFmt), string(autoFmt))
}

// internal/fmt/terraformfmt_test.go
package terraformfmt

import (
	"os"
	"os/exec"
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

func TestAutoMatchesBinary(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	autoFmt, err := Format(src, "test.tf", string(StrategyAuto))
	require.NoError(t, err)
	binFmt, err := Format(src, "test.tf", string(StrategyBinary))
	require.NoError(t, err)
	require.Equal(t, string(binFmt), string(autoFmt))
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

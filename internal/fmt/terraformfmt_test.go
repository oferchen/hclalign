// internal/fmt/terraformfmt_test.go
package terraformfmt

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoMatchesBinary(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	goFmt, _, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	binFmt, _, err := Format(src, "test.tf", string(StrategyBinary))
	require.NoError(t, err)
	require.Equal(t, string(binFmt), string(goFmt))
}

func TestAutoUsesGoFormatter(t *testing.T) {
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	autoFmt, _, err := Format(src, "test.tf", string(StrategyAuto))
	require.NoError(t, err)
	goFmt, _, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(goFmt), string(autoFmt))
}

func TestIdempotent(t *testing.T) {
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	first, _, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	second, _, err := Format(first, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
}

func TestBinaryPreservesHints(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {\r\n  type = string\r\n}\r\n")...)
	_, hints, err := Format(src, "test.tf", string(StrategyBinary))
	require.NoError(t, err)
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
}

func TestRunPreservesHints(t *testing.T) {
	src := append([]byte{0xef, 0xbb, 0xbf}, []byte("variable \"a\" {\r\n  type = string\r\n}\r\n")...)
	_, hints, err := Run(context.Background(), src)
	require.NoError(t, err)
	require.True(t, hints.HasBOM)
	require.Equal(t, "\r\n", hints.Newline)
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

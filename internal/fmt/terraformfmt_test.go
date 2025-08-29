package terraformfmt

import (
	"os/exec"
	"testing"

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

func TestIdempotent(t *testing.T) {
	src := []byte("variable \"a\" {\n  type = string\n}\n")
	first, err := Format(src, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	second, err := Format(first, "test.tf", string(StrategyGo))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
}

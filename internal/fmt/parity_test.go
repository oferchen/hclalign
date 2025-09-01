// internal/fmt/parity_test.go
package terraformfmt

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixturesMatchTerraform(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	casesDir := filepath.Join("..", "..", "tests", "cases")
	globs, err := filepath.Glob(filepath.Join(casesDir, "*", "in.tf"))
	require.NoError(t, err)
	for _, in := range globs {
		src, err := os.ReadFile(in)
		require.NoError(t, err)
		goFmt, _, err := Format(context.Background(), src, in, string(StrategyGo))
		require.NoError(t, err)
		binFmt, _, err := Format(context.Background(), src, in, string(StrategyBinary))
		require.NoError(t, err)
		require.Equalf(t, binFmt, goFmt, "fixture %s", in)
	}
}

package terraformfmt

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func readTestFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return b
}

func TestFormatGo(t *testing.T) {
	in := readTestFile(t, "unformatted.tf")
	want := readTestFile(t, "formatted.tf")
	got, err := Format(context.Background(), in, StrategyGo)
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

func TestFormatBinary(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found")
	}
	in := readTestFile(t, "unformatted.tf")
	want := readTestFile(t, "formatted.tf")
	got, err := Format(context.Background(), in, StrategyBinary)
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

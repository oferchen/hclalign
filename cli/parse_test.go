// cli/parse_test.go
package cli

import (
	"runtime"
	"strconv"
	"testing"

	"github.com/oferchen/hclalign/config"
	"github.com/stretchr/testify/require"
)

func TestParseConfigValid(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{"--check"}))
	cfg, err := parseConfig(cmd, []string{"target"})
	require.NoError(t, err)
	require.Equal(t, "target", cfg.Target)
	require.Equal(t, config.ModeCheck, cfg.Mode)
}

func TestParseConfigTargetWithStdin(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{"--stdin", "--stdout"}))
	_, err := parseConfig(cmd, []string{"file"})
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestParseConfigStdinRequiresStdout(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{"--stdin"}))
	_, err := parseConfig(cmd, nil)
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestParseConfigMutuallyExclusiveFlags(t *testing.T) {
	cmd := newRootCmd(true)
	cmd.SetArgs([]string{"--check", "--diff"})
	_, err := cmd.ExecuteC()
	require.Error(t, err)
}

func TestParseConfigNoTarget(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{}))
	_, err := parseConfig(cmd, nil)
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestParseConfigConcurrencyValidation(t *testing.T) {
	cmd := newRootCmd(true)
	bad := runtime.GOMAXPROCS(0) + 1
	require.NoError(t, cmd.ParseFlags([]string{"--concurrency", strconv.Itoa(bad)}))
	_, err := parseConfig(cmd, []string{"target"})
	require.Error(t, err)
	var exitErr *ExitCodeError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.Code)
}

func TestParseConfigPrefixOrder(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{"--prefix-order"}))
	cfg, err := parseConfig(cmd, []string{"target"})
	require.NoError(t, err)
	require.True(t, cfg.PrefixOrder)
}

func TestParseConfigSkipTerraformFmt(t *testing.T) {
	cmd := newRootCmd(true)
	require.NoError(t, cmd.ParseFlags([]string{"--skip-terraform-fmt"}))
	cfg, err := parseConfig(cmd, []string{"target"})
	require.NoError(t, err)
	require.True(t, cfg.SkipTerraformFmt)
}

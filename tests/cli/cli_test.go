package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "hclalign")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}

func TestCLI(t *testing.T) {
	bin := buildBinary(t)

	t.Run("write", func(t *testing.T) {
		unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"
		want := "variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n"

		dir := t.TempDir()
		file := filepath.Join(dir, "test.tf")
		require.NoError(t, os.WriteFile(file, []byte(unformatted), 0o644))

		cmd := exec.Command(bin, file, "--write")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.NoError(t, err)
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())

		data, err := os.ReadFile(file)
		require.NoError(t, err)
		require.Equal(t, want, string(data))
	})

	t.Run("check_success", func(t *testing.T) {
		formatted := "variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n"
		dir := t.TempDir()
		file := filepath.Join(dir, "test.tf")
		require.NoError(t, os.WriteFile(file, []byte(formatted), 0o644))

		cmd := exec.Command(bin, file, "--check")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.NoError(t, err)
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String())
	})

	t.Run("diff_failure", func(t *testing.T) {
		unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"
		dir := t.TempDir()
		file := filepath.Join(dir, "test.tf")
		require.NoError(t, os.WriteFile(file, []byte(unformatted), 0o644))

		cmd := exec.Command(bin, file, "--diff")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.Error(t, err)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Equal(t, 1, exitErr.ExitCode())
		outStr := stdout.String()
		require.Contains(t, outStr, "-  type = string")
		require.Contains(t, outStr, "+  type        = string")
		require.Contains(t, stderr.String(), "files need formatting")
	})

	t.Run("stdin_stdout", func(t *testing.T) {
		unformatted := "variable \"a\" {\n  type = string\n  description = \"d\"\n}\n"
		want := "variable \"a\" {\n  description = \"d\"\n  type        = string\n}\n"

		cmd := exec.Command(bin, "--stdin", "--stdout")
		cmd.Stdin = strings.NewReader(unformatted)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.NoError(t, err)
		require.Equal(t, want, stdout.String())
		require.Empty(t, stderr.String())
	})

	t.Run("missing_target", func(t *testing.T) {
		cmd := exec.Command(bin)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.Error(t, err)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Equal(t, 2, exitErr.ExitCode())
		require.NotEmpty(t, stderr.String())
	})

	t.Run("invalid_hcl", func(t *testing.T) {
		invalid := "variable \"a\" {"
		dir := t.TempDir()
		file := filepath.Join(dir, "bad.tf")
		require.NoError(t, os.WriteFile(file, []byte(invalid), 0o644))

		cmd := exec.Command(bin, file)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.Error(t, err)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Equal(t, 3, exitErr.ExitCode())
		require.Empty(t, stdout.String())
		require.NotEmpty(t, stderr.String())
	})
}
